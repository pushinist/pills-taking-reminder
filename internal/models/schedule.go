package models

import "time"

type frequency string

const (
	Once   frequency = "once"
	Twice  frequency = "twice"
	Thrice frequency = "thrice"
	Fourth frequency = "fourth"
	Hourly frequency = "hourly"
)

type Schedule struct {
	ID           int64     `json:"schedule_id"`
	MedicineName string    `json:"medicine_name"`
	Frequency    frequency `json:"frequency"`
	StartDate    time.Time `json:"start_date"`
	TakingTime   string    `json:"taking_time"`
	EndDate      time.Time `json:"end_date"`
	UserID       int64     `json:"user_id"`
}
