package security

import (
	"testing"
	"time"

	usermodel "cityhawk/backend/internal/user/model"
)

func TestJWTTokenServiceSignAndParse(t *testing.T) {
	svc := NewJWTTokenService([]byte("secret"))
	user := usermodel.User{ID: "user-1", Email: "user@example.com"}

	token, err := svc.SignAccessToken(user, time.Minute)
	if err != nil {
		t.Fatalf("SignAccessToken() error = %v", err)
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != "user-1" || claims.Email != "user@example.com" || claims.Type != "access" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTTokenServiceParseInvalidToken(t *testing.T) {
	svc := NewJWTTokenService([]byte("secret"))
	if _, err := svc.Parse("bad-token"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestBcryptPasswordServiceAndSHA(t *testing.T) {
	passwords := NewBcryptPasswordService()
	hash, err := passwords.Hash("verysecret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !passwords.Verify("verysecret", hash) {
		t.Fatal("Verify() = false, want true")
	}
	if passwords.Verify("wrong", hash) {
		t.Fatal("Verify() = true for wrong password")
	}

	if got := SHA256Hex("value"); got != "cd42404d52ad55ccfa9aca4adc828aa5800ad9d385a0671fbcbf724118320619" {
		t.Fatalf("SHA256Hex() = %q", got)
	}
}
