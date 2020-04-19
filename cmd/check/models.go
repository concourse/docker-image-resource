package main

import "encoding/json"

type Source struct {
	Repository         string          `json:"repository"`
	Tag                Tag             `json:"tag"`
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

// Tag refers to a tag for an image in the registry.
type Tag string

// UnmarshalJSON accepts numeric and string values.
func (tag *Tag) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err == nil {
		*tag = Tag(s)
	} else {
		var n json.RawMessage
		if err = json.Unmarshal(b, &n); err == nil {
			*tag = Tag(n)
		}
	}
	return err
}
