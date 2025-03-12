package pg

import (
	"database/sql"
	"fmt"
	"pills-taking-reminder/internal/models"
	"pills-taking-reminder/internal/utils"
	"time"
)

func createTable(db *sql.DB, statement, operation string) error {
	stmt, err := db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}

func createInfinitySchedule(s *Storage, schedule models.ScheduleRequest, operation string) (int64, error) {

	stmt, err := s.db.Prepare(addInfiniteScheduleQuery)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	var id int64
	var startTime string
	if time.Now().Hour() < 22 {
		startTime = time.Now().Format("2006-01-02")
	} else {
		startTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}
	err = stmt.QueryRow(schedule.MedicineName, startTime, schedule.UserID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	takingTimes, err := utils.CountTakings(schedule.Frequency)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	err = insertTakingTimes(s, takingTimes, operation, id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}

func createEndingSchedule(s *Storage, schedule models.ScheduleRequest, operation string) (int64, error) {

	stmt, err := s.db.Prepare(addTemporaryScheduleQuery)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	var id int64
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, schedule.Duration)
	err = stmt.QueryRow(schedule.MedicineName, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), schedule.UserID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	takingTimes, err := utils.CountTakings(schedule.Frequency)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	err = insertTakingTimes(s, takingTimes, operation, id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}

func insertTakingTimes(s *Storage, takingTimes []string, operation string, id int64) error {
	for _, takingTime := range takingTimes {
		stmt, err := s.db.Prepare(addTakingTimeQuery)
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}
		_, err = stmt.Exec(id, takingTime)
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}
	}
	return nil
}
