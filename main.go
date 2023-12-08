package main

import (
	awsClient "aws-with-go/s3client"
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func main() {
	iamService := awsClient.GetS3ClientInstance().IAMClient

	// check if policy exists
	policy, errPolicy := awsClient.CheckIfPolicyExists(context.TODO(), iamService, os.Getenv("AWS_POLICY_ARN"))
	if errPolicy != nil {
		panic(errPolicy)
	}
	// get policy version
	policyVersion, errPolicyVersion := awsClient.GetPolicyVersion(context.TODO(), iamService, os.Getenv("AWS_POLICY_ARN"), policy)
	if errPolicyVersion != nil {
		panic(errPolicyVersion)
	}
	// parse and modify the policy document
	decoded, errDecode := url.QueryUnescape(*policyVersion.PolicyVersion.Document)
	if errDecode != nil {
		panic(errDecode)
	}
	var policyDocument map[string]interface{}
	errJsonMarshal := json.Unmarshal([]byte(decoded), &policyDocument)
	if errJsonMarshal != nil {
		panic(errJsonMarshal)
	}
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
