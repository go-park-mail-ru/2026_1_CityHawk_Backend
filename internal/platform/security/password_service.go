package security

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordService struct{}

func NewBcryptPasswordService() *BcryptPasswordService {
	return &BcryptPasswordService{}
}

func (BcryptPasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (BcryptPasswordService) Verify(password, stored string) bool {
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
}
