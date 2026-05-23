package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// CognitoJWTAuthenticator validates Cognito-issued JWT bearer tokens against
// the user pool JWKS endpoint.
type CognitoJWTAuthenticator struct {
	issuer   string
	clientID string
	keyFunc  jwt.Keyfunc
}

// NewCognitoJWTAuthenticator builds an authenticator that fetches signing
// keys from the Cognito user pool JWKS document.
func NewCognitoJWTAuthenticator(region, userPoolID, clientID string) (*CognitoJWTAuthenticator, error) {
	issuer := CognitoIssuer(region, userPoolID)
	jwksURL := issuer + "/.well-known/jwks.json"
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			// JWKS refresh failures are logged by the runtime; the cached keys
			// remain usable until the next successful refresh.
			_ = err
		},
	})
	if err != nil {
		return nil, fmt.Errorf("load cognito jwks: %w", err)
	}
	return NewCognitoJWTAuthenticatorWithKeyFunc(issuer, clientID, jwks.Keyfunc), nil
}

// NewCognitoJWTAuthenticatorWithKeyFunc constructs an authenticator with an
// injected key resolver. Intended for unit tests.
func NewCognitoJWTAuthenticatorWithKeyFunc(issuer, clientID string, keyFunc jwt.Keyfunc) *CognitoJWTAuthenticator {
	return &CognitoJWTAuthenticator{
		issuer:   issuer,
		clientID: clientID,
		keyFunc:  keyFunc,
	}
}

// CognitoIssuer returns the issuer URL for a Cognito user pool.
func CognitoIssuer(region, userPoolID string) string {
	return fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolID)
}

// Authenticate validates a Cognito access or ID token supplied as a bearer
// token and returns the caller's Cognito sub as UserID.
func (a *CognitoJWTAuthenticator) Authenticate(_ context.Context, header string) (Principal, error) {
	raw, ok := extractBearer(header)
	if !ok {
		return Principal{}, ErrUnauthorized
	}
	parsed, err := jwt.Parse(raw, a.keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	if err != nil || !parsed.Valid {
		return Principal{}, ErrUnauthorized
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Principal{}, ErrUnauthorized
	}
	if !claimMatches(claims, "iss", a.issuer) {
		return Principal{}, ErrUnauthorized
	}
	switch tokenUse, _ := claims["token_use"].(string); tokenUse {
	case "access":
		if !claimMatches(claims, "client_id", a.clientID) {
			return Principal{}, ErrUnauthorized
		}
	case "id":
		if !claimMatchesAudience(claims, a.clientID) {
			return Principal{}, ErrUnauthorized
		}
	default:
		return Principal{}, ErrUnauthorized
	}
	sub, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(sub) == "" {
		return Principal{}, ErrUnauthorized
	}
	return Principal{UserID: strings.TrimSpace(sub), Email: emailFromClaims(claims)}, nil
}

func emailFromClaims(claims jwt.MapClaims) string {
	if email, ok := claims["email"].(string); ok && strings.TrimSpace(email) != "" {
		return strings.TrimSpace(email)
	}
	if username, ok := claims["username"].(string); ok && strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	return ""
}

func claimMatches(claims jwt.MapClaims, name, expected string) bool {
	actual, ok := claims[name].(string)
	return ok && actual == expected
}

func claimMatchesAudience(claims jwt.MapClaims, clientID string) bool {
	aud, ok := claims["aud"]
	if !ok {
		return false
	}
	switch v := aud.(type) {
	case string:
		return v == clientID
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == clientID {
				return true
			}
		}
	}
	return false
}

// IsCognitoConfigured reports whether Cognito JWT authentication should be
// used based on environment variables.
func IsCognitoConfigured(poolID, clientID string) bool {
	return strings.TrimSpace(poolID) != "" && strings.TrimSpace(clientID) != ""
}
