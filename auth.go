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
	Email string `json:"email,omitempty"`
	Type  string `json:"type"`
	jwt.RegisteredClaims
}

func (a *authService) issueTokenPair(u user) (refreshResponse, error) {
	accessToken, err := a.signAccessToken(u, a.accessTTL)
	if err != nil {
		return refreshResponse{}, err
	}

	refreshToken, err := a.generateOpaqueToken(32)
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

func (a *authService) rotateRefresh(oldRefreshToken string, users *userStore) (refreshResponse, error) {
	userID, err := a.consumeRefresh(oldRefreshToken)
	if err != nil {
		return refreshResponse{}, errors.New("token revoked")
	}

	return a.issueTokenPairForUserID(userID, users)
}

func (a *authService) issueTokenPairForUserID(userID string, users *userStore) (refreshResponse, error) {
	u, ok := users.getByID(userID)
	if !ok {
		return refreshResponse{}, errors.New("user not found")
	}
	return a.issueTokenPair(u)
}

func (a *authService) signAccessToken(u user, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}

	claims := claims{
		Email: u.Email,
		Type:  "access",
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
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != jwt.SigningMethodHS256.Alg() {
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

func (a *authService) consumeRefresh(token string) (string, error) {
	a.refreshSessions.mu.Lock()
	defer a.refreshSessions.mu.Unlock()

	hash := tokenHash(token)
	s, ok := a.refreshSessions.session[hash]
	if !ok {
		return "", errors.New("token revoked")
	}

	if time.Now().UTC().After(s.ExpiresAt) {
		delete(a.refreshSessions.session, hash)
		return "", errors.New("token expired")
	}

	delete(a.refreshSessions.session, hash)
	return s.UserID, nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (a *authService) generateOpaqueToken(size int) (string, error) {
	return randomHex(size)
}
