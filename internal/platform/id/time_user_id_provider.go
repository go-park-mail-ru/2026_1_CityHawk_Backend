package id

import (
	"time"

	authusecase "cityhawk/backend/internal/auth/usecase"
)

const TimeUserIDLayout = "20060102150405.000000000"

type TimeUserIDProvider struct {
	layout string
}

var _ authusecase.UserIDProvider = (*TimeUserIDProvider)(nil)

func NewTimeUserIDProvider(layout string) *TimeUserIDProvider {
	if layout == "" {
		layout = TimeUserIDLayout
	}
	return &TimeUserIDProvider{layout: layout}
}

func (p *TimeUserIDProvider) New() string {
	return time.Now().UTC().Format(p.layout)
}
