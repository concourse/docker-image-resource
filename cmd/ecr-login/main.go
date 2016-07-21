package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strings"

   "github.com/aws/aws-sdk-go/aws"
   "github.com/aws/aws-sdk-go/aws/session"
   "github.com/aws/aws-sdk-go/service/ecr"
)

func main() {

	repositoryPtr := flag.String("repository", "", "aws ecr repository")
	awsRegionPtr := flag.String("region", "", "aws region")

	flag.Parse()

	repository := *repositoryPtr
	awsRegion := *awsRegionPtr

	registryHost, _ := parseRepository(repository)

	// infer registery id from registry host
	frags := strings.Split(registryHost, ".")
	registryId := frags[0]

	// create session 
	svc := ecr.New(session.New(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	params := &ecr.GetAuthorizationTokenInput{
		 RegistryIds: []*string{
			  aws.String(registryId),
		 },
	}

   resp, err := svc.GetAuthorizationToken(params)
   if err != nil {
       // Print the error, cast err to awserr.Error to get the Code and Message from the error.
       fmt.Println(err.Error())
       return
   }

   encodedToken := (*resp.AuthorizationData[0].AuthorizationToken)
   decodedToken, err := base64.StdEncoding.DecodeString(encodedToken)
   if err != nil {
       fmt.Println("token decode error:", err)
       return
   }

   tokens := strings.Split(string(decodedToken), ":")
   username := tokens[0]
   password := tokens[1]

	fmt.Printf("docker login -u %s -p %s %s", username, password, repository)
}

const officialRegistry = "registry-1.docker.io"
func parseRepository(repository string) (string, string) {
	segs := strings.Split(repository, "/")

	switch len(segs) {
	case 3:
		return segs[0], segs[1] + "/" + segs[2]
	case 2:
		if strings.Contains(segs[0], ":") || strings.Contains(segs[0], ".") {
			return segs[0], segs[1]
		} else {
			return officialRegistry, segs[0] + "/" + segs[1]
		}
	case 1:
		return officialRegistry, "library/" + segs[0]
	}

	panic("malformed repository url")
}
