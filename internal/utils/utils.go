package utils

import (
	"fmt"
	"pills-taking-reminder/internal/models"
	"time"
)

const (
	startTimeString string = "08:00"
	endTimeString   string = "22:00"
)

func RoundTime(timeString string) (string, error) {
	t, err := time.Parse("15:04", timeString)
	if err != nil {
		return "", fmt.Errorf("failed to parse time: %w", err)
	}

	hour := t.Hour()
	minute := t.Minute()

	remainder := minute % 15

	if remainder > 0 {
		minute += 15 - remainder
	}

	if minute == 60 {
		minute = 0
		hour = (hour + 1) % 24
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}

func CountTakings(frequency models.Frequency) ([]string, error) {
	startTime, err := time.Parse("15:04", startTimeString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime, err := time.Parse("15:04", endTimeString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	duration := endTime.Sub(startTime)

	var count int
	switch frequency {
	case models.Once:
		count = 1
	case models.Twice:
		count = 2
	case models.Thrice:
		count = 3
	case models.Fourth:
		count = 4
	case models.Hourly:
		count = 15
	}
	takingTimes := make([]string, count)

	if count == 1 {
		interval := duration / time.Duration(count)
		takingTimes[0] = startTime.Add(interval / 2).Format("15:04")
	} else {
		interval := duration / time.Duration(count-1)
		takingTime := startTime
		for i := range count {
			takingTimes[i], err = RoundTime(takingTime.Format("15:04"))
			if err != nil {
				return nil, fmt.Errorf("failed to calculate taking times: %w", err)
			}
			takingTime = takingTime.Add(interval)
		}
	}
	return takingTimes, nil
}
