package entities

import "time"

type TakingTime struct {
	Time time.Time
}

type Taking struct {
	MedicineName string
	TakingTime   time.Time
}

func (t Taking) IsUpcoming(from time.Time) bool {
	return t.TakingTime.After(from)
}

func (t Taking) TimeUntil(from time.Time) time.Duration {
	if t.TakingTime.Before(from) {
		return 0
	}
	return t.TakingTime.Sub(from)
}

func (t Taking) FormatTime() string {
	return t.TakingTime.Format("15:04")
}

func (t Taking) IsDue(from time.Time, window time.Duration) bool {
	if t.TakingTime.Before(from) {
		return false
	}

	due := t.TakingTime.Sub(from)
	return due <= window
}
