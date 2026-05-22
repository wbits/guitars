// Command lambda is the AWS Lambda entrypoint for the GuitarCollection API.
//
// It is intentionally thin: all it does is wire the production adapters into
// the application service and start the Lambda runtime. Local development
// against LocalStack is enabled by the AWS_ENDPOINT_URL environment variable,
// which both the DynamoDB and Secrets Manager clients honour automatically.
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"

	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/storage"
	httpapi "github.com/wbits/guitars/internal/guitarcollection/interfaces/http"
)

// uuidGen implements application.IDGenerator using UUIDv4.
type uuidGen struct{}

func (uuidGen) NewID() string { return uuid.NewString() }

func main() {
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		// LocalStack resources are provisioned in us-east-1 (see Makefile).
		awsCfg.Region = "us-east-1"
	}

	tableName := envOrDefault("GUITARS_TABLE", "Guitars")

	ddbOpts := []func(*dynamodb.Options){}
	s3Opts := []func(*s3.Options){}
	smOpts := []func(*secretsmanager.Options){}
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		ddbOpts = append(ddbOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
		smOpts = append(smOpts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	ddb := dynamodb.NewFromConfig(awsCfg, ddbOpts...)
	repo := persistence.NewDynamoRepository(ddb, tableName)

	var presigner *storage.Presigner
	if bucket := os.Getenv("IMAGES_BUCKET"); bucket != "" {
		cdnBase := os.Getenv("IMAGES_CDN_BASE_URL")
		if cdnBase == "" {
			log.Fatal("IMAGES_CDN_BASE_URL is required when IMAGES_BUCKET is set")
		}
		s3Client := s3.NewFromConfig(awsCfg, s3Opts...)
		presigner = storage.NewPresigner(
			s3Client,
			bucket,
			cdnBase,
			os.Getenv("IMAGES_S3_PUBLIC_ENDPOINT"),
		)
	}

	authn, err := auth.BuildAuthenticator(ctx, awsCfg, smOpts)
	if err != nil {
		log.Fatalf("build authenticator: %v", err)
	}

	svc := application.NewService(repo, uuidGen{})
	handler := httpapi.NewHandler(svc, authn, presigner)

	log.Printf("guitars lambda starting (table=%s, auth=%s, uploads=%t)", tableName, auth.AuthenticatorMode(), presigner != nil)
	lambda.Start(handler.Handle)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
