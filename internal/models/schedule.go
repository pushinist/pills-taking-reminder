package models

import (
	"database/sql"
	"time"
)

type ScheduleRequest struct {
	MedicineName string `json:"medicine_name" validate:"required"`
	Frequency    int    `json:"frequency" validate:"required,gte=1,lte=15"`
	Duration     int    `json:"duration" validate:"gte=0"`
	UserID       int64  `json:"user_id" validate:"required,gte=1"`
}

type ScheduleResponse struct {
	ID           int64        `json:"id"`
	MedicineName string       `json:"medicine_name"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      sql.NullTime `json:"end_date"`
	UserID       int64        `json:"user_id"`
	TakingTime   []string     `json:"taking_time"`
}
