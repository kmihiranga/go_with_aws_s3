package s3client

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	env "github.com/joho/godotenv"
)

// S3ClientSingleton struct holds the instance of the S3 client.
type S3ClientSingleton struct {
	Client *s3.Client
	IAMClient *iam.Client
	PreSignedClient *s3.PresignClient
}

var instance *S3ClientSingleton
var once sync.Once

// GetS3ClientInstance | return the singleton instance of the S3 client
func GetS3ClientInstance() *S3ClientSingleton {
	// load .env files
	errEnvRead := env.Load()
	if errEnvRead != nil {
		log.Fatal("error loading .env file")
	}
	// retrieve environment variable
	accessKeyID := os.Getenv("AWS_ACCESS_KEY")
	secretAccessKey := os.Getenv("AWS_SECRET_KEY")
	region := os.Getenv("AWS_REGION")

	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")), config.WithRegion(region))
		if err != nil {
			panic(err)
		}
		s3Client := s3.NewFromConfig(cfg)
		instance = &S3ClientSingleton{
			Client: s3Client,
			IAMClient: iam.NewFromConfig(cfg),
			PreSignedClient: s3.NewPresignClient(s3Client),
		}
		log.Println("AWS S3 initialized")
	})
	return instance
}