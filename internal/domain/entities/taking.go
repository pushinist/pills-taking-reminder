package entities

import "time"

type TakingTime struct {
	Time time.Time
}

type Taking struct {
	MedicineName string
	TakingTime   time.Time
}

func (t Taking) FormatTime() string {
	return t.TakingTime.Format("15:04")
}
