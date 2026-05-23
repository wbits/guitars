// Package auth contains the authentication adapters used by the
// interfaces/http layer. Production uses Cognito-issued JWT bearer tokens;
// local development can still fall back to a shared bearer token in
// Secrets Manager when Cognito env vars are not configured.
package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// ErrUnauthorized is returned by Authenticator.Authenticate when the supplied
// Authorization header is missing, malformed, or does not match the expected
// bearer token.
var ErrUnauthorized = errors.New("unauthorized")

// Authenticator validates an HTTP Authorization header value and returns the
// authenticated caller.
type Authenticator interface {
	Authenticate(ctx context.Context, authorizationHeader string) (Principal, error)
}

// TokenLoader abstracts how the expected bearer token is fetched. Production
// uses Secrets Manager; tests can inject a function that returns a fixed value.
type TokenLoader interface {
	Load(ctx context.Context) (string, error)
}

// TokenLoaderFunc adapts a plain function to the TokenLoader interface.
type TokenLoaderFunc func(ctx context.Context) (string, error)

// Load implements TokenLoader.
func (f TokenLoaderFunc) Load(ctx context.Context) (string, error) { return f(ctx) }

// BearerAuthenticator implements Authenticator using a cached token fetched
// via a TokenLoader.
//
// The token is cached for cacheTTL so that the hot path of every Lambda
// invocation does not hit Secrets Manager. A zero cacheTTL disables caching.
type BearerAuthenticator struct {
	loader   TokenLoader
	cacheTTL time.Duration

	mu      sync.RWMutex
	token   string
	loadedAt time.Time
}

// NewBearerAuthenticator constructs a BearerAuthenticator that retrieves the
// expected token via loader and caches it for ttl.
func NewBearerAuthenticator(loader TokenLoader, ttl time.Duration) *BearerAuthenticator {
	return &BearerAuthenticator{loader: loader, cacheTTL: ttl}
}

// Authenticate validates the supplied Authorization header. It returns the
// local-dev principal on success and ErrUnauthorized otherwise.
func (a *BearerAuthenticator) Authenticate(ctx context.Context, header string) (Principal, error) {
	supplied, ok := extractBearer(header)
	if !ok {
		return Principal{}, ErrUnauthorized
	}
	expected, err := a.getToken(ctx)
	if err != nil {
		return Principal{}, fmt.Errorf("load bearer token: %w", err)
	}
	if expected == "" {
		return Principal{}, ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) != 1 {
		return Principal{}, ErrUnauthorized
	}
	return Principal{UserID: bearerUserID()}, nil
}

func (a *BearerAuthenticator) getToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	if a.token != "" && (a.cacheTTL == 0 || time.Since(a.loadedAt) < a.cacheTTL) {
		t := a.token
		a.mu.RUnlock()
		return t, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()
	// Re-check after acquiring the write lock.
	if a.token != "" && (a.cacheTTL == 0 || time.Since(a.loadedAt) < a.cacheTTL) {
		return a.token, nil
	}
	t, err := a.loader.Load(ctx)
	if err != nil {
		return "", err
	}
	a.token = strings.TrimSpace(t)
	a.loadedAt = time.Now()
	return a.token, nil
}

func extractBearer(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) && !strings.HasPrefix(header, strings.ToLower(prefix)) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

// --- Secrets Manager loader ------------------------------------------------

// SecretsManagerAPI is the subset of the Secrets Manager client used here.
type SecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManagerLoader loads the expected bearer token from an AWS Secrets
// Manager secret. The secret value is expected to be the raw token string.
type SecretsManagerLoader struct {
	Client   SecretsManagerAPI
	SecretID string
}

// Load implements TokenLoader.
func (l *SecretsManagerLoader) Load(ctx context.Context) (string, error) {
	out, err := l.Client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(l.SecretID),
	})
	if err != nil {
		return "", err
	}
	if out.SecretString == nil {
		return "", errors.New("secret has no string value")
	}
	return strings.TrimSpace(*out.SecretString), nil
}

func bearerUserID() string {
	if id := strings.TrimSpace(os.Getenv("LOCAL_DEV_USER_ID")); id != "" {
		return id
	}
	return "local-dev-user"
}
