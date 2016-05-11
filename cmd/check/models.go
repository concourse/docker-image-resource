package main

type Source struct {
	Repository         string   `json:"repository"`
	Tag                string   `json:"tag"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	Email              string   `json:"email"`
	InsecureRegistries []string `json:"insecure_registries"`
	RegistryMirror     string   `json:"registry_mirror"`
}

type Version struct {
	Digest string `json:"digest"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version
