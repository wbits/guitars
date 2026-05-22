package auth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBearerAuthenticator_AcceptsCorrectToken(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), time.Minute)
	if err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestBearerAuthenticator_RejectsWrongToken(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), time.Minute)
	if err := a.Authenticate(context.Background(), "Bearer not-the-token"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestBearerAuthenticator_RejectsMissingHeader(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), time.Minute)
	if err := a.Authenticate(context.Background(), ""); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestBearerAuthenticator_RejectsMalformedHeader(t *testing.T) {
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "secret-token", nil
	}), time.Minute)
	for _, h := range []string{"Basic abc", "Bearer", "Bearer  ", "secret-token"} {
		if err := a.Authenticate(context.Background(), h); !errors.Is(err, ErrUnauthorized) {
			t.Errorf("header %q: expected ErrUnauthorized, got %v", h, err)
		}
	}
}

func TestBearerAuthenticator_PropagatesLoaderErrors(t *testing.T) {
	loaderErr := errors.New("boom")
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		return "", loaderErr
	}), time.Minute)
	err := a.Authenticate(context.Background(), "Bearer x")
	if err == nil || errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected loader error to surface, got %v", err)
	}
	if !errors.Is(err, loaderErr) {
		t.Errorf("expected wrapped loader error, got %v", err)
	}
}

func TestBearerAuthenticator_CachesToken(t *testing.T) {
	var calls int32
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "secret-token", nil
	}), time.Minute)
	for i := 0; i < 5; i++ {
		if err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("loader should be called once with caching, got %d calls", got)
	}
}

func TestBearerAuthenticator_RefreshesAfterTTL(t *testing.T) {
	var calls int32
	a := NewBearerAuthenticator(TokenLoaderFunc(func(context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "secret-token", nil
	}), 1*time.Millisecond)
	if err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
		t.Fatalf("first auth: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := a.Authenticate(context.Background(), "Bearer secret-token"); err != nil {
		t.Fatalf("second auth: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Errorf("loader should refresh after TTL, got %d calls", got)
	}
}
