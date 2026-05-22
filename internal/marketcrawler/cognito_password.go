package marketcrawler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func resolveCognitoPassword(ctx context.Context) (string, error) {
	if password := strings.TrimSpace(os.Getenv("COGNITO_PASSWORD")); password != "" {
		return password, nil
	}
	secretID := strings.TrimSpace(os.Getenv("COGNITO_PASSWORD_SECRET_ID"))
	if secretID == "" {
		return "", nil
	}
	region := strings.TrimSpace(os.Getenv("COGNITO_REGION"))
	if region == "" {
		region = "eu-central-1"
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("load aws config for cognito password secret: %w", err)
	}
	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("read cognito password secret %q: %w", secretID, err)
	}
	if out.SecretString == nil {
		return "", fmt.Errorf("read cognito password secret %q: empty value", secretID)
	}
	return strings.TrimSpace(*out.SecretString), nil
}
