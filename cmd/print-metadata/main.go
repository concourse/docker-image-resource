package main

import (
	"encoding/json"
	"os"
	"os/user"
	"strings"
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
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	return currentUser.Username
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
