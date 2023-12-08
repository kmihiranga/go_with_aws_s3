package main

import (
	awsClient "aws-with-go/s3client"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	iamService := awsClient.GetS3ClientInstance().IAMClient
	bucketToAdd := fmt.Sprintf("arn:aws:s3:::%v", os.Getenv("AWS_BUCKET_NAME"))

	// check if policy exists
	policy, errPolicy := awsClient.CheckIfPolicyExists(context.TODO(), iamService, os.Getenv("AWS_POLICY_ARN"))
	if errPolicy != nil {
		panic(errPolicy)
	}
	log.Println("checked existing policy")

	// get policy version
	policyVersion, errPolicyVersion := awsClient.GetPolicyVersion(context.TODO(), iamService, os.Getenv("AWS_POLICY_ARN"), policy)
	if errPolicyVersion != nil {
		panic(errPolicyVersion)
	}
	log.Println("Retrieved existing policy version")

	// parse and modify the policy document
	decoded, errDecode := url.QueryUnescape(*policyVersion.PolicyVersion.Document)
	if errDecode != nil {
		panic(errDecode)
	}
	log.Println("Decoded policy version details")

	var policyDocument map[string]interface{}
	errJsonMarshal := json.Unmarshal([]byte(decoded), &policyDocument)
	if errJsonMarshal != nil {
		panic(errJsonMarshal)
	}

	// modify the policy document
	if statements, ok := policyDocument["Statement"].([]interface{}); ok {
		for _, stmt := range statements {
			if statement, ok := stmt.(map[string]interface{}); ok {
				if resource, ok := statement["Resource"].([]interface{}); ok {
					statement["Resource"] = append(resource, bucketToAdd)
				}
			}
		}
	}
	log.Println("Added resource bucket to the policy")

	modifiedPolicy, errModifiedPolicy := json.Marshal(policyDocument)
	if errModifiedPolicy != nil {
		panic(errModifiedPolicy)
	}

	// create a new policy version
	errNewPolicyVersion := awsClient.CreatePolicyVersion(context.TODO(), iamService, os.Getenv("AWS_POLICY_ARN"), modifiedPolicy)
	if errNewPolicyVersion != nil {
		panic(errNewPolicyVersion)
	}
	log.Println("Created a new policy version")

	// check if bucket exists
	checkBucketExists, _ := awsClient.CheckIfBucketExists(context.TODO(), os.Getenv("AWS_BUCKET_NAME"))
	if checkBucketExists {
		// use the S3 client from the singleton instance
		bucketOutPut, errBucketOutPut := awsClient.ListBucketObjects(context.TODO(), os.Getenv("AWS_BUCKET_NAME"))
		if errBucketOutPut != nil {
			log.Fatal(errBucketOutPut)
		}
	
		for _, object := range bucketOutPut.Contents {
			log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
		}
	} else {
		_, err := awsClient.CreateBucket(context.TODO(), os.Getenv("AWS_BUCKET_NAME"), os.Getenv("AWS_REGION"))
		if err != nil {
			log.Fatal(err)
		}
	}
}
