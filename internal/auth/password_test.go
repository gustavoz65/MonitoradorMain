package auth

import "testing"

func TestValidateCredentials(t *testing.T) {
	if err := ValidateCredentials("gustavo", "gustavo@example.com", "Senha123"); err != nil {
		t.Fatalf("expected valid credentials, got %v", err)
	}
	if err := ValidateCredentials("gu", "gustavo@example.com", "Senha123"); err == nil {
		t.Fatal("expected short username to fail")
	}
	if err := ValidateCredentials("gustavo", "gustavo@example.com", "fraca"); err == nil {
		t.Fatal("expected weak password to fail")
	}
}
