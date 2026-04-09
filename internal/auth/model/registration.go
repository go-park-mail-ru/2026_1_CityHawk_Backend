package model

import usermodel "cityhawk/backend/internal/user/model"

type RegistrationResult struct {
	User   usermodel.User
	Tokens TokenPair
}
