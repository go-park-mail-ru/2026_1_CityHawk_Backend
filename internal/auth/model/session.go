package model

import usermodel "cityhawk/backend/internal/user/model"

type SessionResult struct {
	User   usermodel.User
	Tokens TokenPair
}
