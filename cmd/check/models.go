package main

import "encoding/json"

type Source struct {
	Repository         string          `json:"repository"`
	Tag                json.Number     `json:"tag"`
	Username           string          `json:"username"`
	Password           string          `json:"password"`
	InsecureRegistries []string        `json:"insecure_registries"`
	RegistryMirror     string          `json:"registry_mirror"`
	DomainCerts        []DomainCert    `json:"ca_certs"`
	ClientCerts        []ClientCertKey `json:"client_certs"`

	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
}

type Version struct {
	Digest string `json:"digest"`
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
