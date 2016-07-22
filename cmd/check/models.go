package main

type Source struct {
   Repository         string       `json:"repository"`
   Tag                string       `json:"tag"`
   Username           string       `json:"username"`
   Password           string       `json:"password"`
   AwsAccessKeyId     string       `json:"aws_access_key_id"`
   AwsSecretAccessKey string       `json:"aws_secret_access_key"`
   AwsRegion          string       `json:"aws_region"`
   InsecureRegistries []string     `json:"insecure_registries"`
   RegistryMirror     string       `json:"registry_mirror"`
   DomainCerts        []DomainCert `json:"ca_certs"`
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
