package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	usermodel "cityhawk/backend/internal/user/model"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email string `json:"email,omitempty"`
	Type  string `json:"type"`
	jwt.RegisteredClaims
}

type JWTTokenService struct {
	secret []byte
}

func NewJWTTokenService(secret []byte) *JWTTokenService {
	return &JWTTokenService{secret: secret}
}

func (s *JWTTokenService) SignAccessToken(u usermodel.User, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	jti, err := randomHex(16)
	if err != nil {
		return "", err
	}

	claims := Claims{
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
	return token.SignedString(s.secret)
}

func (s *JWTTokenService) GenerateOpaqueToken(size int) (string, error) {
	return randomHex(size)
}

func (s *JWTTokenService) Parse(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}

	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return *c, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
