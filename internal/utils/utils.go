package utils

import (
	"fmt"
	"time"
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
