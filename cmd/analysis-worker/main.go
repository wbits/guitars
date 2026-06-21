package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/wbits/guitars/internal/guitaranalysis"
	analysispersistence "github.com/wbits/guitars/internal/guitaranalysis/persistence"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
	profileapp "github.com/wbits/guitars/internal/userprofile/application"
	profilepersistence "github.com/wbits/guitars/internal/userprofile/infrastructure/persistence"
	profilecrypto "github.com/wbits/guitars/internal/userprofile/infrastructure/crypto"
)

func main() {
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		awsCfg.Region = "us-east-1"
	}

	ddbOpts := []func(*dynamodb.Options){}
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		ddbOpts = append(ddbOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}
	ddb := dynamodb.NewFromConfig(awsCfg, ddbOpts...)

	guitarsTable := envOrDefault("GUITARS_TABLE", "Guitars")
	profilesTable := envOrDefault("USER_PROFILES_TABLE", "UserProfiles")
	analysisTable := envOrDefault("GUITAR_ANALYSIS_TABLE", "GuitarAnalysis")

	guitarRepo := persistence.NewDynamoRepository(ddb, guitarsTable)
	profileRepo := profilepersistence.NewDynamoRepository(ddb, profilesTable, "usernameIndex")

	var byokEncryptor profileapp.BYOKEncryptor
	if keyB64 := os.Getenv("ASSISTANT_BYOK_ENCRYPTION_KEY"); keyB64 != "" {
		store, err := profilecrypto.NewKeyStoreFromBase64(keyB64)
		if err != nil {
			log.Fatalf("assistant BYOK encryption key: %v", err)
		}
		byokEncryptor = store
	}

	profiles := profileapp.NewService(profileRepo, byokEncryptor)
	analysisRepo := analysispersistence.NewDynamoRepository(ddb, analysisTable)
	analysisSvc := guitaranalysis.NewService(
		analysisRepo,
		&guitaranalysis.ProfileOwnerLoader{Profiles: profiles},
		&guitaranalysis.VisionAnalyzer{},
		nil,
		guitaranalysis.GuitarLoaderFunc(guitarRepo.FindByID),
	)

	log.Printf("analysis worker starting (guitars=%s, profiles=%s, analysis=%s)", guitarsTable, profilesTable, analysisTable)
	lambda.Start(func(ctx context.Context, event events.SQSEvent) error {
		for _, record := range event.Records {
			var job guitaranalysis.Job
			if err := json.Unmarshal([]byte(record.Body), &job); err != nil {
				log.Printf("skip invalid job body: %v", err)
				continue
			}
			if err := analysisSvc.ProcessJob(ctx, job); err != nil {
				return err
			}
		}
		return nil
	})
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
