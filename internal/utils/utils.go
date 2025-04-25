package utils

import (
	"fmt"
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

func CountTakings(frequency int) ([]string, error) {

	if frequency <= 0 || frequency > 15 {
		return nil, fmt.Errorf("frequency must be between 1 and 15")
	}

	startTime, err := time.Parse("15:04", startTimeString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime, err := time.Parse("15:04", endTimeString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	duration := endTime.Sub(startTime)

	takingTimes := make([]string, frequency)

	if frequency == 1 {
		interval := duration / time.Duration(frequency)
		takingTimes[0] = startTime.Add(interval / 2).Format("15:04")
	} else {
		interval := duration / time.Duration(frequency-1)
		takingTime := startTime
		for i := range frequency {
			takingTimes[i], err = RoundTime(takingTime.Format("15:04"))
			if err != nil {
				return nil, fmt.Errorf("failed to calculate taking times: %w", err)
			}
			takingTime = takingTime.Add(interval)
		}
	}
	return takingTimes, nil
}
