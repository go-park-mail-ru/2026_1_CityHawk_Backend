package grpc

import (
	"context"
	"testing"

	authmodel "cityhawk/backend/internal/auth/model"
	platformsecurity "cityhawk/backend/internal/platform/security"
	usermodel "cityhawk/backend/internal/user/model"
	authv1 "cityhawk/backend/pkg/pb/auth/v1"
	"github.com/golang-jwt/jwt/v5"
)

func TestServerAuthFlow(t *testing.T) {
	user := usermodel.User{ID: "user-1", Email: "user@example.com", Username: "user", Role: usermodel.RoleAdmin}
	tokens := authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}
	server := NewServer(fakeAuthFlow{user: user, tokens: tokens}, fakeRefresh{tokens: tokens}, fakeOAuth{tokens: tokens}, fakeAuthUsers{user: user}, fakeParser{claims: platformsecurity.Claims{Type: "access", RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1"}}})

	if resp, err := server.Register(context.Background(), &authv1.RegisterRequest{Email: user.Email, Username: user.Username, UserSurname: "surname", Password: "password"}); err != nil || resp.GetUserId() != user.ID {
		t.Fatalf("Register() = (%+v, %v)", resp, err)
	}
	if resp, err := server.Login(context.Background(), &authv1.LoginRequest{Email: user.Email, Password: "password"}); err != nil || resp.GetTokens().GetAccessToken() != "access" {
		t.Fatalf("Login() = (%+v, %v)", resp, err)
	}
	if resp, err := server.OAuthLogin(context.Background(), &authv1.OAuthLoginRequest{Provider: authv1.OAuthProvider_OAUTH_PROVIDER_GOOGLE, Code: "code"}); err != nil || resp.GetUserId() != user.ID {
		t.Fatalf("OAuthLogin() = (%+v, %v)", resp, err)
	}
	if resp, err := server.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "old"}); err != nil || resp.GetTokens().GetRefreshToken() != "refresh" {
		t.Fatalf("Refresh() = (%+v, %v)", resp, err)
	}
	if resp, err := server.Logout(context.Background(), &authv1.LogoutRequest{RefreshToken: "refresh"}); err != nil || !resp.GetOk() {
		t.Fatalf("Logout() = (%+v, %v)", resp, err)
	}
	if resp, err := server.ValidateAccessToken(context.Background(), &authv1.ValidateAccessTokenRequest{AccessToken: "access"}); err != nil || !resp.GetValid() || resp.GetContext().GetUserId() != user.ID {
		t.Fatalf("ValidateAccessToken() = (%+v, %v)", resp, err)
	}
}

func TestServerAuthErrors(t *testing.T) {
	server := NewServer(fakeAuthFlow{}, fakeRefresh{}, fakeOAuth{}, fakeAuthUsers{}, fakeParser{})
	if _, err := server.OAuthLogin(context.Background(), &authv1.OAuthLoginRequest{}); err == nil {
		t.Fatal("OAuthLogin() accepted missing provider")
	}
	if _, err := server.Refresh(context.Background(), &authv1.RefreshRequest{}); err == nil {
		t.Fatal("Refresh() accepted missing refresh token")
	}
	if _, err := server.Logout(context.Background(), &authv1.LogoutRequest{}); err == nil {
		t.Fatal("Logout() accepted missing refresh token")
	}
	if resp, err := server.ValidateAccessToken(context.Background(), &authv1.ValidateAccessTokenRequest{AccessToken: "bad"}); err != nil || resp.GetValid() {
		t.Fatalf("ValidateAccessToken() = (%+v, %v), want invalid", resp, err)
	}
}

type fakeAuthFlow struct {
	user   usermodel.User
	tokens authmodel.TokenPair
}

func (f fakeAuthFlow) Register(context.Context, authmodel.RegisterInput) (authmodel.RegistrationResult, error) {
	return authmodel.RegistrationResult{User: f.user, Tokens: f.tokens}, nil
}

func (f fakeAuthFlow) Login(context.Context, authmodel.LoginInput) (authmodel.SessionResult, error) {
	return authmodel.SessionResult{User: f.user, Tokens: f.tokens}, nil
}

type fakeRefresh struct{ tokens authmodel.TokenPair }

func (f fakeRefresh) RotateRefresh(context.Context, string) (authmodel.TokenPair, error) {
	return f.tokens, nil
}
func (f fakeRefresh) RevokeRefresh(context.Context, string) error { return nil }

type fakeOAuth struct{ tokens authmodel.TokenPair }

func (f fakeOAuth) LoginWithGoogle(context.Context, string) (authmodel.TokenPair, error) {
	return f.tokens, nil
}
func (f fakeOAuth) LoginWithYandex(context.Context, string) (authmodel.TokenPair, error) {
	return f.tokens, nil
}
func (f fakeOAuth) LoginWithVK(context.Context, string) (authmodel.TokenPair, error) {
	return f.tokens, nil
}

type fakeAuthUsers struct{ user usermodel.User }

func (f fakeAuthUsers) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	if f.user.ID == id {
		return f.user, true
	}
	return usermodel.User{}, false
}

type fakeParser struct{ claims platformsecurity.Claims }

func (f fakeParser) Parse(string) (platformsecurity.Claims, error) { return f.claims, nil }
