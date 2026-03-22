package id

import (
	"time"
)

const TimeUserIDLayout = "20060102150405.000000000"

type TimeUserIDProvider struct {
	layout string
}

func NewTimeUserIDProvider(layout string) *TimeUserIDProvider {
	if layout == "" {
		layout = TimeUserIDLayout
	}
	return &TimeUserIDProvider{layout: layout}
}

func (p *TimeUserIDProvider) New() string {
	return time.Now().UTC().Format(p.layout)
}
