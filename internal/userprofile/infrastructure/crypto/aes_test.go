package crypto

import "testing"

func TestKeyStore_EncryptDecrypt(t *testing.T) {
	store, err := NewKeyStoreFromBase64("MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := store.Encrypt("sk-test-key-12345")
	if err != nil {
		t.Fatal(err)
	}
	if encoded == "sk-test-key-12345" {
		t.Fatal("expected ciphertext")
	}
	plain, err := store.Decrypt(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "sk-test-key-12345" {
		t.Fatalf("got %q", plain)
	}
}

func TestNewKeyStoreFromBase64_RejectsWrongSize(t *testing.T) {
	if _, err := NewKeyStoreFromBase64("c2hvcnQ="); err == nil {
		t.Fatal("expected error for short key")
	}
}
