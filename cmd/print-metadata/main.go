package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/concourse/docker-image-resource/cmd/print-metadata/passwd"
)

type imageMetadata struct {
	User string   `json:"user,omitempty"`
	Env  []string `json:"env"`
}

var blacklistedEnv = map[string]bool{
	"HOSTNAME": true,
}

var userFile = flag.String("userFile", "/etc/passwd", "")

func main() {
	flag.Parse()

	username, err := getUsername(*userFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to determine username, will not be included in metadata")
	}

	err = json.NewEncoder(os.Stdout).Encode(imageMetadata{
		User: username,
		Env:  env(),
	})
	if err != nil {
		panic(err)
	}
}

func getUsername(userFile string) (string, error) {
	users, err := passwd.ReadUsers(userFile)
	if err != nil {
		return "", err
	}

	name, found := users.NameForID(syscall.Getuid())
	if !found {
		return "", fmt.Errorf("could not find user in %s", userFile)
	}

	return name, nil
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
