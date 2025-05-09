package dto

type ScheduleRequest struct {
	MedicineName string `json:"medicine_name" validate:"required"`
	Frequency    int    `json:"frequency" validate:"required,gte=1,lte=15"`
	Duration     int    `json:"duration" validate:"gte=0"`
	UserID       int64  `json:"user_id" validate:"required,gte=1"`
}

type ScheduleResponse struct {
	ID           int64    `json:"id"`
	MedicineName string   `json:"medicine_name"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	UserID       int64    `json:"user_id"`
	TakingTime   []string `json:"taking_time"`
}

type Taking struct {
	MedicineName string `json:"medicine_name"`
	TakingTime   string `json:"taking_time"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetScheduleParams struct {
	UserID     int64 `json:"user_id"`
	ScheduleID int64 `json:"schedule_id"`
}

type GetScheduleIDsParams struct {
	UserID int64 `json:"user_id"`
}

type GetNextTakingsParams struct {
	UserID int64 `json:"user_id"`
}
