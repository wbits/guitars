package application

import (
	"context"
	"testing"

	"github.com/wbits/guitars/internal/userprofile/infrastructure/crypto"
	"github.com/wbits/guitars/internal/userprofile/infrastructure/persistence"
)

func testEncryptor(t *testing.T) BYOKEncryptor {
	t.Helper()
	store, err := crypto.NewKeyStoreFromBase64("MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestService_SetAssistantBYOK_EncryptsAndClears(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := NewService(repo, testEncryptor(t))
	ctx := context.Background()

	updated, err := svc.SetAssistantBYOK(ctx, "user-1", "user@example.com", "sk-test", "https://example.com/v1", "gpt-test")
	if err != nil {
		t.Fatal(err)
	}
	if !updated.AssistantBYOKConfigured() {
		t.Fatal("expected configured")
	}
	if updated.AssistantLLMModel() != "gpt-test" {
		t.Fatalf("model: %q", updated.AssistantLLMModel())
	}

	stored, err := repo.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AssistantEncryptedAPIKey() == "sk-test" {
		t.Fatal("expected encrypted storage")
	}

	creds, ok, err := svc.AssistantBYOKCredentialsForUser(ctx, "user-1")
	if err != nil || !ok || creds.APIKey != "sk-test" {
		t.Fatalf("credentials: ok=%v err=%v creds=%+v", ok, err, creds)
	}

	cleared, err := svc.ClearAssistantBYOK(ctx, "user-1", "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if cleared.AssistantBYOKConfigured() {
		t.Fatal("expected cleared")
	}
}

func TestService_SetAssistantBYOK_RequiresServerKey(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := NewService(repo, nil)
	_, err := svc.SetAssistantBYOK(context.Background(), "user-1", "", "sk-test", "", "")
	if !IsBYOKNotConfigured(err) {
		t.Fatalf("got %v", err)
	}
}
