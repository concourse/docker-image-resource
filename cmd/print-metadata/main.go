package main

import (
	"encoding/json"
	"os"
	"strings"
)

type imageMetadata struct {
	Env []string `json:"env"`
}

var blacklistedEnv = map[string]bool{
	"HOSTNAME": true,
}

func main() {
	var envVars []string
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		name := parts[0]

		if !blacklistedEnv[name] {
			envVars = append(envVars, e)
		}
	}

	err := json.NewEncoder(os.Stdout).Encode(imageMetadata{
		Env: envVars,
	})
	if err != nil {
		panic(err)
	}
}
