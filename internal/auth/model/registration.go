package model

import usermodel "cityhawk/backend/internal/user/model"

type RegisterInput struct {
	Email    string
	Username string
	Password string
	Birthday string
	CityID   string
}

type LoginInput struct {
	Email    string
	Password string
}

type RegistrationResult struct {
	User   usermodel.User
	Tokens TokenPair
}
