package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestCognitoAuthenticator(t *testing.T) (*CognitoJWTAuthenticator, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	authn := NewCognitoJWTAuthenticatorWithKeyFunc(
		"https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_TestPool",
		"test-client-id",
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodRS256 {
				return nil, errors.New("unexpected signing method")
			}
			return &key.PublicKey, nil
		},
	)
	return authn, key
}

func signCognitoToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestCognitoJWTAuthenticator_AcceptsAccessToken(t *testing.T) {
	authn, key := newTestCognitoAuthenticator(t)
	token := signCognitoToken(t, key, jwt.MapClaims{
		"iss":        authn.issuer,
		"token_use":  "access",
		"client_id":  "test-client-id",
		"sub":        "user-123",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	})
	if err := authn.Authenticate(context.Background(), "Bearer "+token); err != nil {
		t.Fatalf("expected valid access token, got %v", err)
	}
}

func TestCognitoJWTAuthenticator_AcceptsIDToken(t *testing.T) {
	authn, key := newTestCognitoAuthenticator(t)
	token := signCognitoToken(t, key, jwt.MapClaims{
		"iss":       authn.issuer,
		"token_use": "id",
		"aud":       "test-client-id",
		"sub":       "user-123",
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	})
	if err := authn.Authenticate(context.Background(), "Bearer "+token); err != nil {
		t.Fatalf("expected valid id token, got %v", err)
	}
}

func TestCognitoJWTAuthenticator_RejectsWrongClient(t *testing.T) {
	authn, key := newTestCognitoAuthenticator(t)
	token := signCognitoToken(t, key, jwt.MapClaims{
		"iss":       authn.issuer,
		"token_use": "access",
		"client_id": "other-client",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	err := authn.Authenticate(context.Background(), "Bearer "+token)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestCognitoJWTAuthenticator_RejectsExpiredToken(t *testing.T) {
	authn, key := newTestCognitoAuthenticator(t)
	token := signCognitoToken(t, key, jwt.MapClaims{
		"iss":       authn.issuer,
		"token_use": "access",
		"client_id": "test-client-id",
		"exp":       time.Now().Add(-time.Hour).Unix(),
	})
	err := authn.Authenticate(context.Background(), "Bearer "+token)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestCognitoJWTAuthenticator_RejectsSharedSecretToken(t *testing.T) {
	authn, _ := newTestCognitoAuthenticator(t)
	err := authn.Authenticate(context.Background(), "Bearer local-dev-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
