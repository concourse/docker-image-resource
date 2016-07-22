package main

import (
	"fmt"
	"os"
	"strings"

   "github.com/aws/aws-sdk-go/aws"
   "github.com/aws/aws-sdk-go/aws/session"
   "github.com/aws/aws-sdk-go/service/ecr"
)

func getEcrResponse(request CheckRequest, registryHost string, repo string, tag string) (response CheckResponse) {

   source := request.Source

   // infer registery id from registry host
   frags := strings.Split(registryHost, ".")
   registryId := frags[0]

   // sdk pulls creds from environment
   os.Setenv("AWS_ACCESS_KEY_ID", source.AwsAccessKeyId)
   os.Setenv("AWS_SECRET_ACCESS_KEY", source.AwsSecretAccessKey)

   // create session 
   svc := ecr.New(session.New(&aws.Config{
      Region: aws.String(source.AwsRegion),
   }))

   // this only pulls first 100 images, won't find older images; iterate using the NextToken
   // field if it becomes a problem.
	params := &ecr.ListImagesInput{
		 RepositoryName: aws.String(repo), // Required
		 MaxResults:     aws.Int64(100),
       // NextToken:      aws.String("NextToken"),
		 RegistryId:     aws.String(registryId),
	}

	resp, err := svc.ListImages(params)
	if err != nil {
		 // Print the error, cast err to awserr.Error to get the Code and Message from an error.
		 fmt.Println(err.Error())
		 return
	}

   // iterate over the image ids to find the first matching supplied tag
   response = CheckResponse{}
   image_ids := resp.ImageIds
   for i := 0; i < len(image_ids); i++ {

      if image_ids[i].ImageTag == nil {
         continue
      }

      image_tag :=  *image_ids[i].ImageTag
      if tag == image_tag {
         digest := *image_ids[i].ImageDigest
         response = CheckResponse{Version{digest}}
      }
   }

   return response
}
