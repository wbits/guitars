package marketcrawler

import (
	"testing"
)

func TestResolveAPIToken_PrefersExplicitToken(t *testing.T) {
	t.Setenv("GUITARS_API_TOKEN", "explicit-token")
	t.Setenv("COGNITO_CLIENT_ID", "")
	token, err := ResolveAPIToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if token != "explicit-token" {
		t.Fatalf("want explicit-token, got %q", token)
	}
}

func TestTokenFromCognito_RequiresEnv(t *testing.T) {
	t.Setenv("COGNITO_REGION", "")
	t.Setenv("COGNITO_CLIENT_ID", "")
	t.Setenv("COGNITO_USERNAME", "")
	t.Setenv("COGNITO_PASSWORD", "")
	_, err := TokenFromCognito(t.Context())
	if err == nil {
		t.Fatal("expected error when cognito env is missing")
	}
}
