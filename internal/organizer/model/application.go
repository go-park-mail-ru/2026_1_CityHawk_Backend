package model

import "time"

type ApplicationStatus string

const (
	StatusPending   ApplicationStatus = "pending"
	StatusNeedsInfo ApplicationStatus = "needs_info"
	StatusApproved  ApplicationStatus = "approved"
	StatusRejected  ApplicationStatus = "rejected"
)

func (s ApplicationStatus) Valid() bool {
	switch s {
	case StatusPending, StatusNeedsInfo, StatusApproved, StatusRejected:
		return true
	default:
		return false
	}
}

type Application struct {
	ID            string
	UserID        string
	Status        ApplicationStatus
	Name          string
	Email         string
	Phone         string
	City          string
	ProjectName   string
	Categories    string
	Links         string
	About         string
	ReviewComment string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
