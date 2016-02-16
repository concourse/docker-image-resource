package main

import (
	"encoding/json"
	"os"
	"strings"
	"syscall"

	"github.com/concourse/docker-image-resource/cmd/print-metadata/passwd"
)

type imageMetadata struct {
	User string   `json:"user"`
	Env  []string `json:"env"`
}

var blacklistedEnv = map[string]bool{
	"HOSTNAME": true,
}

func main() {
	err := json.NewEncoder(os.Stdout).Encode(imageMetadata{
		User: username(),
		Env:  env(),
	})
	if err != nil {
		panic(err)
	}
}

func username() string {
	users, err := passwd.ReadUsers("/etc/passwd")
	if err != nil {
		panic(err)
	}

	name, found := users.NameForID(syscall.Getuid())
	if !found {
		panic("could not find user in /etc/passwd")
	}

	return name
}

func env() []string {
	var envVars []string
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		name := parts[0]

		if !blacklistedEnv[name] {
			envVars = append(envVars, e)
		}
	}

	return envVars
}
