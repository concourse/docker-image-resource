package main

import (
   "encoding/base64"
	"fmt"
	"os"
	"strings"

   "github.com/aws/aws-sdk-go/aws"
   "github.com/aws/aws-sdk-go/aws/session"
   "github.com/aws/aws-sdk-go/service/ecr"
)

func getAwsCredentials (source Source, registryHost string) (username string, password string) {

   // infer registery id from registry host
   frags := strings.Split(registryHost, ".")
   registryId := frags[0]
   fmt.Printf("Extracted registry id: %s\n", registryId)

   // sdk pulls creds from environment
   os.Setenv("AWS_ACCESS_KEY_ID", source.AwsAccessKeyId)
   os.Setenv("AWS_SECRET_ACCESS_KEY", source.AwsSecretAccessKey)

   // create session 
   svc := ecr.New(session.New(&aws.Config{
      Region: aws.String(source.AwsRegion),
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
   username = tokens[0]
   password = tokens[1]

   fmt.Printf("Aws Username: %s, Password: %s.\n", username, password)
   return username, password
}
