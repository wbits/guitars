package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// BuildAuthenticator selects Cognito JWT validation in production and falls
// back to the legacy shared bearer token when Cognito env vars are absent
// (local development against LocalStack).
func BuildAuthenticator(ctx context.Context, awsCfg aws.Config, smOpts []func(*secretsmanager.Options)) (Authenticator, error) {
	poolID := os.Getenv("COGNITO_USER_POOL_ID")
	clientID := os.Getenv("COGNITO_CLIENT_ID")
	if IsCognitoConfigured(poolID, clientID) {
		region := awsCfg.Region
		if region == "" {
			region = envOrDefault("AWS_REGION", "eu-central-1")
		}
		return NewCognitoJWTAuthenticator(region, poolID, clientID)
	}

	secretID := envOrDefault("BEARER_SECRET_ID", "guitars/bearer-token")
	sm := secretsmanager.NewFromConfig(awsCfg, smOpts...)
	loader := &SecretsManagerLoader{Client: sm, SecretID: secretID}
	return NewBearerAuthenticator(loader, 5*time.Minute), nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// AuthenticatorMode describes which auth adapter BuildAuthenticator selected.
func AuthenticatorMode() string {
	if IsCognitoConfigured(os.Getenv("COGNITO_USER_POOL_ID"), os.Getenv("COGNITO_CLIENT_ID")) {
		return "cognito-jwt"
	}
	return "legacy-bearer"
}

// MustBuildAuthenticator is like BuildAuthenticator but panics on error.
func MustBuildAuthenticator(ctx context.Context, awsCfg aws.Config, smOpts []func(*secretsmanager.Options)) Authenticator {
	authn, err := BuildAuthenticator(ctx, awsCfg, smOpts)
	if err != nil {
		panic(fmt.Sprintf("build authenticator: %v", err))
	}
	return authn
}
