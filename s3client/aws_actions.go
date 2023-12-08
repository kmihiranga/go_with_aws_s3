package s3client

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var s3ClientSingleton = GetS3ClientInstance().Client

// create a new bucket
func CreateBucket(ctx context.Context,bucketName string, bucketRegion string) (bool, error) {
	_, err := s3ClientSingleton.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(bucketRegion),
		},
	})
	if err != nil {
		return false, fmt.Errorf("couldn't create bucket %v in region %v. Here's why: %v", bucketName, bucketRegion, err)
	}
	return true, nil
}

// check if bucket exists
func CheckIfBucketExists(ctx context.Context,bucketName string) (bool, error) {
	_, err := s3ClientSingleton.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var nbe *types.NoSuchBucket
		if errors.As(err, &nbe) {
			return false, fmt.Errorf("error accessing bucket %v. %v", bucketName, err)
		}
		return false, fmt.Errorf("bucket name %v is not existing in the S3 service. %v", bucketName, err)
	}
	return true, nil
}

// list bucket objects
func ListBucketObjects(ctx context.Context,bucketName string) (*s3.ListObjectsV2Output, error) {
	output, err := s3ClientSingleton.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing bucket objects in bucket %v. %v", bucketName, err)
	}
	return output, nil
}

// check if policy exists
func CheckIfPolicyExists(ctx context.Context, client *iam.Client, policyArn string) (*iam.GetPolicyOutput, error) {
	policy, err := client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		return nil, fmt.Errorf("error checking policy %v. %v", policyArn, err)
	}
	return policy, nil
}

// get policy version
func GetPolicyVersion(ctx context.Context, client *iam.Client, policyArn string, policy *iam.GetPolicyOutput) (*iam.GetPolicyVersionOutput, error) {
	policyVersion, err := client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
		PolicyArn: &policyArn,
		VersionId: policy.Policy.DefaultVersionId,
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving policy version. %v", err)
	}
	return policyVersion, nil
}

// create new policy version
func CreatePolicyVersion(ctx context.Context, client *iam.Client, policyArn string, modifiedPolicy []byte) error {
	_, err := client.CreatePolicyVersion(ctx, &iam.CreatePolicyVersionInput{
		PolicyArn: &policyArn,
		PolicyDocument: aws.String(string(modifiedPolicy)),
		SetAsDefault: *aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("error create policy version. %v", err)
	}
	return nil
}