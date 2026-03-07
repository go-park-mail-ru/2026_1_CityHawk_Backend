package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authService struct {
	secret          []byte
	accessTTL       time.Duration
	refreshTTL      time.Duration
	refreshSessions *refreshStore
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func (a *authService) issueTokenPair(u user) (refreshResponse, error) {
	accessToken, err := a.signToken(u, "access", a.accessTTL)
	if err != nil {
		return refreshResponse{}, err
	}

	refreshToken, err := a.signToken(u, "refresh", a.refreshTTL)
	if err != nil {
		return refreshResponse{}, err
	}

	a.storeRefreshToken(refreshToken, u.ID, time.Now().UTC().Add(a.refreshTTL))

	return refreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(a.accessTTL.Seconds()),
	}, nil
}

func (a *authService) rotateRefresh(oldRefreshToken string) (refreshResponse, error) {
	c, err := a.parseToken(oldRefreshToken)
	if err != nil || c.Type != "refresh" {
		return refreshResponse{}, errors.New("invalid token")
	}

	if !a.isRefreshActive(oldRefreshToken) {
		return refreshResponse{}, errors.New("token revoked")
	}

	a.revokeRefresh(oldRefreshToken)

	u := user{ID: c.UserID, Email: c.Email}
	return a.issueTokenPair(u)
}

func (a *authService) signToken(u user, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}

	claims := claims{
		UserID: u.ID,
		Email:  u.Email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

func (a *authService) parseToken(tokenString string) (claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return a.secret, nil
	})
	if err != nil {
		return claims{}, err
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return claims{}, errors.New("invalid token")
	}

	return *c, nil
}

func (a *authService) storeRefreshToken(token, userID string, expiresAt time.Time) {
	a.refreshSessions.mu.Lock()
	defer a.refreshSessions.mu.Unlock()
	a.refreshSessions.session[tokenHash(token)] = refreshSession{UserID: userID, ExpiresAt: expiresAt}
}

func (a *authService) revokeRefresh(token string) {
	a.refreshSessions.mu.Lock()
	defer a.refreshSessions.mu.Unlock()
	delete(a.refreshSessions.session, tokenHash(token))
}

func (a *authService) isRefreshActive(token string) bool {
	a.refreshSessions.mu.Lock()
	defer a.refreshSessions.mu.Unlock()

	hash := tokenHash(token)
	s, ok := a.refreshSessions.session[hash]
	if !ok {
		return false
	}

	if time.Now().UTC().After(s.ExpiresAt) {
		delete(a.refreshSessions.session, hash)
		return false
	}

	return true
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
