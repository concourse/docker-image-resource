package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/hashicorp/go-multierror"
)

func main() {
	var request CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	fatalIf("failed to read request", err)

	registryHost, repo := parseRepository(request.Source.Repository)

	tag := request.Source.Tag
	if tag == "" {
		tag = "latest"
	}

	transport, registryURL := makeTransport(request, registryHost, repo)

	ub, err := v2.NewURLBuilderFromString(registryURL)
	fatalIf("failed to construct registry URL builder", err)

	client := &http.Client{Transport: transport}

	manifestURL, err := ub.BuildManifestURL(repo, tag)
	fatalIf("failed to build manifest URL", err)

	manifestResponse, err := client.Get(manifestURL)
	fatalIf("failed to fetch manifest", err)

	manifestResponse.Body.Close()

	if manifestResponse.StatusCode != http.StatusOK {
		fatal("failed to fetch digest: " + manifestResponse.Status)
	}

	digest := manifestResponse.Header.Get("Docker-Content-Digest")
	if digest == "" {
		fatal("no digest returned")
	}

	response := CheckResponse{}
	if digest != request.Version.Digest {
		response = append(response, Version{digest})
	}

	json.NewEncoder(os.Stdout).Encode(response)
}

func makeTransport(request CheckRequest, registryHost string, repository string) (http.RoundTripper, string) {
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		DisableKeepAlives: true,
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

	authTransport := transport.NewTransport(baseTransport)

	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
	}

	challengeManager := auth.NewSimpleChallengeManager()

	var registryURL string

	var pingResp *http.Response
	var pingErr error
	var pingErrs error
	for _, scheme := range []string{"https", "http"} {
		registryURL = scheme + "://" + registryHost

		req, err := http.NewRequest("GET", registryURL+"/v2", nil)
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

	switch len(segs) {
	case 3:
		return segs[0], segs[1] + "/" + segs[2]
	case 2:
		if strings.Contains(segs[0], ":") {
			return segs[0], segs[1]
		} else {
			return officialRegistry, segs[0] + "/" + segs[1]
		}
	case 1:
		return officialRegistry, "library/" + segs[0]
	}

	fatal("malformed repository url")
	panic("unreachable")
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
