package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

var VersionRegex = regexp.MustCompile("^\\S*$")

type Source struct {
	Repository         string          `json:"repository"`
	Tag                json.Number     `json:"tag"`
	Regex              string          `json:"regex"`
	Username           string          `json:"username"`
	Password           string          `json:"password"`
	InsecureRegistries []string        `json:"insecure_registries"`
	RegistryMirror     string          `json:"registry_mirror"`
	DomainCerts        []DomainCert    `json:"ca_certs"`
	ClientCerts        []ClientCertKey `json:"client_certs"`

	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	AWSSessionToken    string `json:"aws_session_token"`
}

type Version struct {
	Digest  string           `json:"digest,omitempty"`
	Tag     string           `json:"tag,omitempty"`
}

// CheckedVersion parses the given version and returns a new
// Version.
func CheckedVersion(v string) (string, error) {
	matches := VersionRegex.FindStringSubmatch(v)
	if matches == nil {
		fmt.Errorf("Malformed version: %s", v)
		return v, nil
	}
	fmt.Fprintf(os.Stderr, "** matches : %s\n", matches)
	segments := matches[0]

	return segments, nil
}

func ParsedVersion(raw string) (string error) {
	v := raw
	err := ParseGroup(v)
	return err
}

func ParseGroup(v string) (err error) {
	match := VersionRegex.FindStringSubmatch(v)

	if match == nil {
		return fmt.Errorf("Tag %s filtered by regex", v)
	}
	switch len(match) {
	case 1:
		v, err = CheckedVersion(match[0])
	case 2:
		v, err = CheckedVersion(match[1])
	case 3:
		v, err = CheckedVersion(fmt.Sprintf("%s-%s", match[1], match[2]))
	default:
		v, err = CheckedVersion(fmt.Sprintf("%s-%s+%s", match[1], match[2], match[3]))
	}
	return
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type DomainCert struct {
	Domain string `json:"domain"`
	Cert   string `json:"cert"`
}

type ClientCertKey struct {
	Domain string `json:"domain"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}

type V1Compatibility struct {
	ID              string `json:"id"`
	Parent          string `json:"parent"`
	Created         string `json:"created"`
}
type ManifestResponse struct {
	Name         string `json:"name"`
	Tag          string `json:"tag"`
	Architecture string `json:"architecture"`

	FsLayers []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`

	History []struct {
		V1CompatibilityRaw string `json:"v1Compatibility"`
		V1Compatibility V1Compatibility
	} `json:"history"`
}
