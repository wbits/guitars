package marketcrawler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// ResolveAPIToken returns a bearer token for the GuitarCollection API.
// It prefers GUITARS_API_TOKEN and otherwise signs in to Cognito with
// USER_PASSWORD_AUTH using COGNITO_REGION, COGNITO_CLIENT_ID, COGNITO_USERNAME,
// and COGNITO_PASSWORD.
func ResolveAPIToken(ctx context.Context) (string, error) {
	if token := strings.TrimSpace(os.Getenv("GUITARS_API_TOKEN")); token != "" {
		return token, nil
	}
	return TokenFromCognito(ctx)
}

// TokenFromCognito performs a Cognito USER_PASSWORD_AUTH sign-in and returns
// the ID token JWT so email and group claims are available to the API.
func TokenFromCognito(ctx context.Context) (string, error) {
	region := strings.TrimSpace(os.Getenv("COGNITO_REGION"))
	clientID := strings.TrimSpace(os.Getenv("COGNITO_CLIENT_ID"))
	username := strings.TrimSpace(os.Getenv("COGNITO_USERNAME"))
	password, err := resolveCognitoPassword(ctx)
	if err != nil {
		return "", err
	}

	if region == "" || clientID == "" || username == "" || password == "" {
		return "", fmt.Errorf("cognito sign-in requires COGNITO_REGION, COGNITO_CLIENT_ID, COGNITO_USERNAME, and COGNITO_PASSWORD (or COGNITO_PASSWORD_SECRET_ID)")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("", "", "")),
	)
	if err != nil {
		return "", fmt.Errorf("load aws config: %w", err)
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)
	out, err := client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(clientID),
		AuthParameters: map[string]string{
			"USERNAME": username,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return "", fmt.Errorf("cognito initiate auth: %w", err)
	}
	if out.AuthenticationResult == nil || out.AuthenticationResult.IdToken == nil {
		return "", fmt.Errorf("cognito initiate auth: missing id token")
	}
	return *out.AuthenticationResult.IdToken, nil
}
