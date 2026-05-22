package marketcrawler

import (
	"os"
	"testing"
)

func TestResolveCognitoPassword_PrefersEnv(t *testing.T) {
	t.Setenv("COGNITO_PASSWORD", "  secret-value  ")
	t.Setenv("COGNITO_PASSWORD_SECRET_ID", "ignored")
	password, err := resolveCognitoPassword(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if password != "secret-value" {
		t.Fatalf("want trimmed password, got %q", password)
	}
}

func TestResolveCognitoPassword_EmptyWithoutEnvOrSecret(t *testing.T) {
	t.Setenv("COGNITO_PASSWORD", "")
	t.Setenv("COGNITO_PASSWORD_SECRET_ID", "")
	password, err := resolveCognitoPassword(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if password != "" {
		t.Fatalf("want empty password, got %q", password)
	}
}

func TestResolveCognitoPassword_FromSecretsManager(t *testing.T) {
	if os.Getenv("RUN_AWS_INTEGRATION_TESTS") == "" {
		t.Skip("set RUN_AWS_INTEGRATION_TESTS=1 to run against AWS Secrets Manager")
	}
	t.Setenv("COGNITO_PASSWORD", "")
	t.Setenv("COGNITO_PASSWORD_SECRET_ID", "guitars/crawler-cognito-password")
	t.Setenv("COGNITO_REGION", "eu-central-1")
	password, err := resolveCognitoPassword(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if password == "" {
		t.Fatal("expected password from secrets manager")
	}
}
