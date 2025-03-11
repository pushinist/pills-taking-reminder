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

type Schedule struct {
	ID           int64        `json:"schedule_id"`
	MedicineName string       `json:"medicine_name"`
	Frequency    Frequency    `json:"frequency"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      sql.NullTime `json:"end_date"`
	UserID       int64        `json:"user_id"`
}
