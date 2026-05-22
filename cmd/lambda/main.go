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
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"

	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
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

	tableName := envOrDefault("GUITARS_TABLE", "Guitars")

	ddbOpts := []func(*dynamodb.Options){}
	smOpts := []func(*secretsmanager.Options){}
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		ddbOpts = append(ddbOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
		smOpts = append(smOpts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	ddb := dynamodb.NewFromConfig(awsCfg, ddbOpts...)
	repo := persistence.NewDynamoRepository(ddb, tableName)

	authn, err := auth.BuildAuthenticator(ctx, awsCfg, smOpts)
	if err != nil {
		log.Fatalf("build authenticator: %v", err)
	}

	svc := application.NewService(repo, uuidGen{})
	handler := httpapi.NewHandler(svc, authn)

	log.Printf("guitars lambda starting (table=%s, auth=%s)", tableName, auth.AuthenticatorMode())
	lambda.Start(handler.Handle)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
