package main

import (
	awsClient "aws-with-go/s3client"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gorilla/mux"
)

var iamService = awsClient.GetS3ClientInstance().IAMClient
var bucketToAdd = fmt.Sprintf("arn:aws:s3:::%v", os.Getenv("AWS_BUCKET_NAME"))

func main() {
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
			panic(err)
		}

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
	}

	r := mux.NewRouter()
	r.HandleFunc("/aws_upload", uploadFileHandler).Methods("POST")
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr: "127.0.0.1:3000",
		WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// maximum upload size
	const maxUploadSize = 10 * 1024 * 1024 // 10 MB

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse the multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	// retrieve the file name from form data
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename

	_, errUploadFile := awsClient.UploadFileToS3Bucket(context.TODO(), os.Getenv("AWS_BUCKET_NAME"), filename, file)
	if errUploadFile != nil {
		http.Error(w, "error uploading file", http.StatusBadRequest)
		return
	}
	log.Println("file uploaded successfully!")
}