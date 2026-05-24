package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGroupsFromClaims(t *testing.T) {
	got := GroupsFromClaims(jwt.MapClaims{
		"cognito:groups": []any{"guitars-admins", "other"},
	})
	if len(got) != 2 || got[0] != "guitars-admins" {
		t.Fatalf("unexpected groups: %#v", got)
	}
}

func TestIsAdmin(t *testing.T) {
	p := Principal{Groups: []string{"guitars-admins"}}
	if !IsAdmin(p, "guitars-admins") {
		t.Fatal("expected admin membership")
	}
	if IsAdmin(p, "other-group") {
		t.Fatal("expected non-member")
	}
}
