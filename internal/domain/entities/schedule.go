package entities

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidFrequency = errors.New("frequency must be between 1 and 15")
	ErrInvalidDuration  = errors.New("duration must be more than 0")
)

type Schedule struct {
	ID           int64
	MedicineName string
	Frequency    int
	Duration     int
	StartDate    time.Time
	EndDate      *time.Time
	UserID       int64
	TakingTimes  []TakingTime
}

func NewSchedule(medicineName string, frequency, duration int, userID int64) (*Schedule, error) {
	if frequency < 1 || frequency > 15 {
		return nil, ErrInvalidFrequency
	}

	startDate := TimeNow()
	var endDate *time.Time

	if duration > 0 {
		end := startDate.AddDate(0, 0, duration)
		endDate = &end
	}

	takingTimes, err := CalculateTakingTimes(frequency)
	if err != nil {
		return nil, err
	}

	schedule := &Schedule{
		MedicineName: medicineName,
		Frequency:    frequency,
		Duration:     duration,
		StartDate:    startDate,
		EndDate:      endDate,
		UserID:       userID,
		TakingTimes:  takingTimes,
	}

	return schedule, nil

}

func (s *Schedule) IsActive(date time.Time) bool {
	if date.Before(s.StartDate) {
		return false
	}

	if s.EndDate != nil && date.After(*s.EndDate) {
		return false
	}

	return true
}

func (s *Schedule) GetNextTakings(from time.Time, interval time.Duration) []Taking {
	if !s.IsActive(from) {
		return nil
	}

	var takings []Taking
	to := from.Add(interval)

	for _, takeTime := range s.TakingTimes {
		takingTime := time.Date(from.Year(), from.Month(), from.Day(), takeTime.Time.Hour(), takeTime.Time.Minute(), 0, 0, from.Location())

		if takingTime.Before(from) {
			takingTime = takingTime.AddDate(0, 0, 1)
		}

		if takingTime.Before(to) {
			takings = append(takings, Taking{
				MedicineName: s.MedicineName,
				TakingTime:   takingTime,
			})
		}
	}
	return takings
}

func CalculateTakingTimes(frequency int) ([]TakingTime, error) {
	if frequency < 1 || frequency > 15 {
		return nil, fmt.Errorf("incorrect frequency")
	}

	startHour := 8
	endHour := 22

	takingTimes := make([]TakingTime, 0, frequency)
	if frequency == 1 {
		takingTime := (startHour + endHour) / 2
		takingTimes = append(takingTimes, TakingTime{
			Time: time.Date(0, 0, 0, takingTime, 0, 0, 0, time.UTC),
		})
	} else {
		dayDuration := endHour - startHour
		intervalHours := float64(dayDuration) / float64(frequency-1)

		for i := range frequency {
			hour := startHour + int(float64(i)*intervalHours)
			minute := int((float64(i)*intervalHours - float64(int(float64(i)*intervalHours))) * 60)

			minute = (minute + 7) / 15 * 15
			if minute == 60 {
				minute = 0
				hour++
			}

			takingTimes = append(takingTimes, TakingTime{
				Time: time.Date(0, 0, 0, hour, minute, 0, 0, time.UTC),
			})
		}
	}
	return takingTimes, nil
}
