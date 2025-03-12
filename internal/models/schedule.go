package models

import (
	"database/sql"
	"time"
)

type Frequency string

const (
	Once   Frequency = "once"
	Twice  Frequency = "twice"
	Thrice Frequency = "thrice"
	Fourth Frequency = "fourth"
	Hourly Frequency = "hourly"
)

type ScheduleRequest struct {
	MedicineName string    `json:"medicine_name"`
	Frequency    Frequency `json:"frequency"`
	Duration     int       `json:"duration"`
	UserID       int64     `json:"user_id"`
}

type ScheduleResponse struct {
	ID           int64        `json:"id"`
	MedicineName string       `json:"medicine_name"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      sql.NullTime `json:"end_date"`
	UserID       int64        `json:"user_id"`
}
