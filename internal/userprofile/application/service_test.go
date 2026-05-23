package application

import (
	"context"
	"errors"
	"testing"

	"github.com/wbits/guitars/internal/userprofile/domain"
	"github.com/wbits/guitars/internal/userprofile/infrastructure/persistence"
)

func TestService_GetProfile_CreatesWithEmail(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := NewService(repo)

	profile, err := svc.GetProfile(context.Background(), "user-1", "user@example.com")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if profile.Email() != "user@example.com" || profile.DisplayName() != "user@example.com" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestService_UpdateUsername_Persists(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	if _, err := svc.GetProfile(ctx, "user-1", "user@example.com"); err != nil {
		t.Fatal(err)
	}
	updated, err := svc.UpdateUsername(ctx, "user-1", "user@example.com", "picker")
	if err != nil {
		t.Fatalf("update username: %v", err)
	}
	if updated.Username() != "picker" || updated.DisplayName() != "picker" {
		t.Fatalf("unexpected profile: email=%q username=%q display=%q", updated.Email(), updated.Username(), updated.DisplayName())
	}
}

func TestService_UpdateUsername_RejectsDuplicate(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	if _, err := svc.GetProfile(ctx, "user-1", "one@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetProfile(ctx, "user-2", "two@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateUsername(ctx, "user-1", "one@example.com", "shared"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.UpdateUsername(ctx, "user-2", "two@example.com", "shared")
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("want ErrUsernameTaken, got %v", err)
	}
}
