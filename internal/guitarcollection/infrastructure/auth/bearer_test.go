package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBearerAuthenticator_AcceptsCorrectToken(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), 0)
	p, err := a.Authenticate(context.Background(), "Bearer secret-token")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if p.UserID != "local-dev-user" {
		t.Fatalf("want local-dev-user, got %q", p.UserID)
	}
}

func TestBearerAuthenticator_RejectsWrongToken(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), 0)
	_, err := a.Authenticate(context.Background(), "Bearer not-the-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestBearerAuthenticator_RejectsMissingHeader(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), 0)
	_, err := a.Authenticate(context.Background(), "")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestBearerAuthenticator_RejectsMalformedHeader(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), 0)
	for _, h := range []string{"secret-token", "Basic secret-token", "Bearer"} {
		if _, err := a.Authenticate(context.Background(), h); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("header %q: expected ErrUnauthorized, got %v", h, err)
		}
	}
}

func TestBearerAuthenticator_PropagatesLoaderErrors(t *testing.T) {
	want := errors.New("boom")
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "", want
	}), 0)
	_, err := a.Authenticate(context.Background(), "Bearer x")
	if !errors.Is(err, want) {
		t.Fatalf("expected loader error, got %v", err)
	}
}

func TestBearerAuthenticator_CachesToken(t *testing.T) {
	calls := 0
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		calls++
		return "secret-token", nil
	}), time.Minute)
	for i := 0; i < 2; i++ {
		if _, err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected 1 loader call, got %d", calls)
	}
}

func TestBearerAuthenticator_RefreshesAfterTTL(t *testing.T) {
	calls := 0
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		calls++
		return "secret-token", nil
	}), time.Nanosecond)
	if _, err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	if _, err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 loader calls, got %d", calls)
	}
}
