package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"code.cloudfoundry.org/lager/v3"
	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	ecrapi "github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/cihub/seelog"
	"github.com/concourse/retryhttp"
	"github.com/distribution/reference"
	"github.com/docker/distribution"
	_ "github.com/docker/distribution/manifest/schema1"
	_ "github.com/docker/distribution/manifest/schema2"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/hashicorp/go-multierror"
	digest "github.com/opencontainers/go-digest"
)

func main() {
	logger := lager.NewLogger("http")
	rECRRepo, err := regexp.Compile(`[a-zA-Z0-9][a-zA-Z0-9_-]*\.dkr\.ecr\.[a-zA-Z0-9][a-zA-Z0-9_-]*\.amazonaws\.com(\.cn)?[^ ]*`)
	fatalIf("failed to compile ECR regex", err)

	var request CheckRequest
	err = json.NewDecoder(os.Stdin).Decode(&request)
	fatalIf("failed to read request", err)

	os.Setenv("AWS_ACCESS_KEY_ID", request.Source.AWSAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", request.Source.AWSSecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", request.Source.AWSSessionToken)

	// silence benign ecr-login errors/warnings
	seelog.UseLogger(seelog.Disabled)

	if rECRRepo.MatchString(request.Source.Repository) == true {
		ecrUser, ecrPass, err := ecr.NewECRHelper(
			ecr.WithClientFactory(ecrapi.DefaultClientFactory{}),
		).Get(request.Source.Repository)
		fatalIf("failed to get ECR credentials", err)
		request.Source.Username = ecrUser
		request.Source.Password = ecrPass
	}

	registryHost, repo := parseRepository(request.Source.Repository)

	explicitlyDeclaredRegistryHost := hasExplicitlyDeclaredRegistryHost(registryHost)
	if len(request.Source.RegistryMirror) > 0 && !explicitlyDeclaredRegistryHost {
		registryMirrorURL, err := url.Parse(request.Source.RegistryMirror)
		fatalIf("failed to parse registry mirror URL", err)
		registryHost = registryMirrorURL.Host
	}

	tag := string(request.Source.Tag)
	if tag == "" {
		tag = "latest"
	}

	transport, registryURL := makeTransport(logger, request, registryHost, repo)

	client := &http.Client{
		Transport: retryRoundTripper(logger, transport),
	}

	ub, err := v2.NewURLBuilderFromString(registryURL, false)
	fatalIf("failed to construct registry URL builder", err)

	namedRef, err := reference.WithName(repo)
	fatalIf("failed to construct named reference", err)

	var response CheckResponse

	taggedRef, err := reference.WithTag(namedRef, tag)
	fatalIf("failed to construct tagged reference", err)

	latestManifestURL, err := ub.BuildManifestURL(taggedRef)
	fatalIf("failed to build latest manifest URL", err)

	latestDigest, foundLatest := headDigest(client, latestManifestURL, request.Source.Repository, tag)

	if request.Version.Digest != "" {
		digestRef, err := reference.WithDigest(namedRef, digest.Digest(request.Version.Digest))
		fatalIf("failed to build cursor manifest URL", err)

		cursorManifestURL, err := ub.BuildManifestURL(digestRef)
		fatalIf("failed to build manifest URL", err)

		cursorDigest, foundCursor := headDigest(client, cursorManifestURL, request.Source.Repository, tag)

		if foundCursor && cursorDigest != latestDigest {
			response = append(response, Version{cursorDigest})
		}
	}

	if foundLatest {
		response = append(response, Version{latestDigest})
	}

	json.NewEncoder(os.Stdout).Encode(response)
}

func headDigest(client *http.Client, manifestURL, repository, tag string) (string, bool) {
	manifestRequest, err := http.NewRequest("HEAD", manifestURL, nil)
	fatalIf("failed to build manifest request", err)
	manifestRequest.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	manifestRequest.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	manifestRequest.Header.Add("Accept", "application/json")

	manifestResponse, err := client.Do(manifestRequest)
	fatalIf("failed to fetch manifest", err)

	defer manifestResponse.Body.Close()

	if manifestResponse.StatusCode == http.StatusNotFound {
		return "", false
	}

	if manifestResponse.StatusCode != http.StatusOK {
		fatal(fmt.Sprintf("failed to fetch digest for image '%s:%s': %s\ndoes the image exist?", repository, tag, manifestResponse.Status))
	}

	digest := manifestResponse.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fetchDigest(client, manifestURL, repository, tag)
	}

	return digest, true
}

func fetchDigest(client *http.Client, manifestURL, repository, tag string) (string, bool) {
	manifestRequest, err := http.NewRequest("GET", manifestURL, nil)
	fatalIf("failed to build manifest request", err)
	manifestRequest.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	manifestRequest.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	manifestRequest.Header.Add("Accept", "application/json")

	manifestResponse, err := client.Do(manifestRequest)
	fatalIf("failed to fetch manifest", err)

	defer manifestResponse.Body.Close()

	if manifestResponse.StatusCode == http.StatusNotFound {
		return "", false
	}

	if manifestResponse.StatusCode != http.StatusOK {
		fatal(fmt.Sprintf("failed to fetch digest for image '%s:%s': %s\ndoes the image exist?", repository, tag, manifestResponse.Status))
	}

	ctHeader := manifestResponse.Header.Get("Content-Type")

	bytes, err := io.ReadAll(manifestResponse.Body)
	fatalIf("failed to read response body", err)

	_, desc, err := distribution.UnmarshalManifest(ctHeader, bytes)
	fatalIf("failed to unmarshal manifest", err)

	return string(desc.Digest), true
}

func makeTransport(logger lager.Logger, request CheckRequest, registryHost string, repository string) (http.RoundTripper, string) {
	// for non self-signed registries, caCertPool must be nil in order to use the system certs
	var caCertPool *x509.CertPool
	if len(request.Source.DomainCerts) > 0 {
		caCertPool = x509.NewCertPool()
		for _, domainCert := range request.Source.DomainCerts {
			ok := caCertPool.AppendCertsFromPEM([]byte(domainCert.Cert))
			if !ok {
				fatal(fmt.Sprintf("failed to parse CA certificate for \"%s\"", domainCert.Domain))
			}
		}
	}

	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{RootCAs: caCertPool},
	}

	var insecure bool
	for _, hostOrCIDR := range request.Source.InsecureRegistries {
		if isInsecure(hostOrCIDR, registryHost) {
			insecure = true
		}
	}

	if insecure {
		baseTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if len(request.Source.ClientCerts) > 0 {
		baseTransport.TLSClientConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: setClientCert(registryHost, request.Source.ClientCerts),
		}
	}

	authTransport := transport.NewTransport(baseTransport)

	pingClient := &http.Client{
		Transport: retryRoundTripper(logger, authTransport),
		Timeout:   1 * time.Minute,
	}

	challengeManager := challenge.NewSimpleManager()

	var registryURL string

	var pingResp *http.Response
	var pingErr error
	var pingErrs error
	for _, scheme := range []string{"https", "http"} {
		registryURL = scheme + "://" + registryHost

		req, err := http.NewRequest("GET", registryURL+"/v2/", nil)
		fatalIf("failed to create ping request", err)

		pingResp, pingErr = pingClient.Do(req)
		if pingErr == nil {
			// clear out previous attempts' failures
			pingErrs = nil
			break
		}

		pingErrs = multierror.Append(
			pingErrs,
			fmt.Errorf("ping %s: %s", scheme, pingErr),
		)
	}
	fatalIf("failed to ping registry", pingErrs)

	defer pingResp.Body.Close()

	err := challengeManager.AddResponse(pingResp)
	fatalIf("failed to add response to challenge manager", err)

	credentialStore := dumbCredentialStore{request.Source.Username, request.Source.Password}
	tokenHandler := auth.NewTokenHandler(authTransport, credentialStore, repository, "pull")
	basicHandler := auth.NewBasicHandler(credentialStore)
	authorizer := auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler)

	return transport.NewTransport(baseTransport, authorizer), registryURL
}

type dumbCredentialStore struct {
	username string
	password string
}

func (dcs dumbCredentialStore) Basic(*url.URL) (string, string) {
	return dcs.username, dcs.password
}

func (dumbCredentialStore) RefreshToken(u *url.URL, service string) string {
	return ""
}

func (dumbCredentialStore) SetRefreshToken(u *url.URL, service, token string) {
}

func fatalIf(doing string, err error) {
	if err != nil {
		fatal(doing + ": " + err.Error())
	}
}

func fatal(message string) {
	println(message)
	os.Exit(1)
}

const officialRegistry = "registry-1.docker.io"

func parseRepository(repository string) (string, string) {
	segs := strings.Split(repository, "/")

	if len(segs) > 1 && (strings.Contains(segs[0], ":") || strings.Contains(segs[0], ".")) {
		// In a private registry pretty much anything is valid.
		return segs[0], strings.Join(segs[1:], "/")
	}
	switch len(segs) {
	case 3:
		return segs[0], segs[1] + "/" + segs[2]
	case 2:
		return officialRegistry, segs[0] + "/" + segs[1]
	case 1:
		return officialRegistry, "library/" + segs[0]
	}

	fatal("malformed repository url")
	panic("unreachable")
}

// Does the repository include an explicitly declared registry host, such as 'foo.com/baz/bar'
// that differs from the officialRegistry?
func hasExplicitlyDeclaredRegistryHost(registryHost string) bool {
	return strings.Contains(registryHost, ".") && registryHost != officialRegistry
}

func isInsecure(hostOrCIDR string, hostPort string) bool {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostOrCIDR == hostPort
	}

	_, cidr, err := net.ParseCIDR(hostOrCIDR)
	if err == nil {
		ip := net.ParseIP(host)
		if ip != nil {
			return cidr.Contains(ip)
		}
	}

	return hostOrCIDR == hostPort
}

func retryRoundTripper(logger lager.Logger, rt http.RoundTripper) http.RoundTripper {
	return &retryhttp.RetryRoundTripper{
		Logger:         logger,
		BackOffFactory: retryhttp.NewExponentialBackOffFactory(5 * time.Minute),
		RoundTripper:   rt,
	}
}

func setClientCert(registry string, list []ClientCertKey) []tls.Certificate {
	var clientCert []tls.Certificate
	for _, r := range list {
		if r.Domain == registry {
			certKey, err := tls.X509KeyPair([]byte(r.Cert), []byte(r.Key))
			if err != nil {
				fatal(fmt.Sprintf("failed to parse client certificate and/or key for \"%s\"", r.Domain))
			}
			clientCert = append(clientCert, certKey)
		}
	}
	return clientCert
}
