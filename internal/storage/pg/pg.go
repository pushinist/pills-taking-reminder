package pg

import (
	"database/sql"
	"fmt"
	"log/slog"
	"pills-taking-reminder/internal/config"
	"pills-taking-reminder/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type StorageRepository interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	NextTakings(id int64) ([]models.Taking, error)
	GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type Storage struct {
	db       *sql.DB
	logger   *slog.Logger
	interval time.Duration
}

func New(dbCfg config.DB, logger *slog.Logger, interval time.Duration) (StorageRepository, error) {
	const operation = "storage.pg.new"

	logger.Info("initing db connection...",
		slog.String("operation", operation),
		slog.String("host", dbCfg.Host),
		slog.String("port", dbCfg.Port),
		slog.String("db_name", dbCfg.Name))

	DSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbCfg.Username,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.Name)
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		logger.Error("failed to connect to db",
			slog.String("operation", operation),
			slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if err = db.Ping(); err != nil {
		logger.Error("failed to ping db",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
	}

	err = createTable(db, createSchedulesQuery, operation, logger)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = createTable(db, createTakingsQuery, operation, logger)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	logger.Info("db with tables is ready to work with!",
		slog.String("operation", operation))

	return &Storage{
		db:       db,
		logger:   logger,
		interval: interval,
	}, nil
}

func (s *Storage) CreateSchedule(schedule models.ScheduleRequest) (int64, error) {
	const operation = "storage.pg.CreateSchedule"

	s.logger.Info("trying to create a schedule",
		slog.String("operation", operation),
		slog.Any("schedule", schedule))

	if schedule.Frequency <= 0 || schedule.Frequency > 15 {
		s.logger.Info("failed to create schedule",
			slog.String("operation", operation),
			slog.Any("schedule", schedule),
			slog.String("error", "frequency must be between 1 and 15"))
		return 0, fmt.Errorf("frequency must be between 1 and 15")
	}
	if schedule.Duration == 0 {
		id, err := createInfinitySchedule(s, schedule, operation)
		if err != nil {
			s.logger.Info("failed to create schedule",
				slog.String("operation", operation),
				slog.Any("schedule", schedule),
				slog.String("error", err.Error()))

			return 0, fmt.Errorf("%s: %w", operation, err)
		}
		s.logger.Info("schedule was successfuly created!",
			slog.String("operation", operation),
			slog.Any("schedule", schedule),
			slog.Int64("schedule_id", id))
		return id, nil
	} else {
		id, err := createEndingSchedule(s, schedule, operation)
		if err != nil {
			s.logger.Info("failed to create schedule",
				slog.String("operation", operation),
				slog.Any("schedule", schedule),
				slog.String("error", err.Error()))
			return 0, fmt.Errorf("%s: %w", operation, err)
		}
		s.logger.Info("schedule was successfuly created!",
			slog.String("operation", operation),
			slog.Any("schedule", schedule),
			slog.Int64("schedule_id", id))
		return id, nil
	}
}

func (s *Storage) GetSchedulesIDs(userID int64) ([]int64, error) {
	const operation = "storage.pg.GetSchedulesIDs"

	s.logger.Info("trying to get ids for user",
		slog.String("operation", operation),
		slog.Int64("user_id", userID))

	stmt, err := s.db.Prepare(getSchedulesQuery)
	if err != nil {
		s.logger.Info("failed to prepare SQL statement",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return []int64{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, time.Now().Format("2006-01-02"))
	if err != nil {
		s.logger.Info("failed to execute SQL statement",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return []int64{}, fmt.Errorf("%s: %w", operation, err)
	}

	defer rows.Close()

	var schedulesIDs []int64
	for rows.Next() {
		var scheduleID int64

		err := rows.Scan(&scheduleID)
		if err != nil {
			s.logger.Info("failed to scan value",
				slog.String("operation", operation),
				slog.String("error", err.Error()))

			return []int64{}, fmt.Errorf("%s: %w", operation, err)
		}
		schedulesIDs = append(schedulesIDs, scheduleID)
	}

	if err := rows.Err(); err != nil {
		s.logger.Info("error in SQL rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))

		return []int64{}, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("schedules ids were successfully found!",
		slog.String("operation", operation),
		slog.Any("schedules_ids", schedulesIDs))

	return schedulesIDs, nil
}

func (s *Storage) NextTakings(id int64) ([]models.Taking, error) {
	const operation = "storage.pg.NextTakings"

	s.logger.Info("trying to get next takings for user",
		slog.String("operation", operation),
		slog.Int64("user_id", id))

	stmt, err := s.db.Prepare(getNextTakingsQuery)
	if err != nil {
		s.logger.Info("failed to prepare SQL statement",
			slog.String("operation", operation),
			slog.String("error", err.Error()))

		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(time.Now().Format("15:04"), time.Now().Add(s.interval).Format("15:04"), id)
	if err != nil {

		s.logger.Info("failed to execute SQL query",
			slog.String("operation", operation),
			slog.String("error", err.Error()))

		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	var takings []models.Taking
	for rows.Next() {
		var taking models.Taking
		err = rows.Scan(&taking.MedicineName, &taking.TakingTime)
		if err != nil {
			s.logger.Info("failed to scan value",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
		}
		takings = append(takings, taking)
	}

	if err := rows.Err(); err != nil {
		s.logger.Info("errors in SQL rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return []models.Taking{}, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("next takings time were successfuly found!",
		slog.String("operation", operation),
		slog.Any("next_takings", takings))
	return takings, nil
}

func (s *Storage) GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error) {
	const operation = "storage.pg.GetSchedule"

	s.logger.Info("trying to get schedule",
		slog.String("operation", operation),
		slog.Int64("user_id", userID),
		slog.Int64("schedule_id", scheduleID))

	stmt, err := s.db.Prepare(getScheduleQuery)
	if err != nil {

		s.logger.Info("failed to prepare SQL statement",
			slog.String("operation", operation),
			slog.String("error", err.Error()))

		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)

	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, scheduleID)
	if err != nil {

		s.logger.Info("failed to execute SQL query",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()
	var schedule models.ScheduleResponse
	var count int
	for rows.Next() {
		var rawTakingTime time.Time
		err := rows.Scan(&schedule.ID, &schedule.MedicineName, &schedule.StartDate, &schedule.EndDate, &schedule.UserID, &rawTakingTime)
		if err != nil {
			s.logger.Info("failed to scan value",
				slog.String("operation", operation),
				slog.String("error", err.Error()))
			return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
		}
		if schedule.ID == 0 {
			return models.ScheduleResponse{}, nil
		}

		takingTime := rawTakingTime.Format("15:04")

		schedule.TakingTime = append(schedule.TakingTime, takingTime)
		count++
	}
	if count == 0 {
		s.logger.Info("haven't found any schedules",
			slog.Int64("user_id", userID),
			slog.Int64("schedule_id", scheduleID))
		return models.ScheduleResponse{}, sql.ErrNoRows
	}

	if err := rows.Err(); err != nil {
		s.logger.Info("errors in SQL rows",
			slog.String("operation", operation),
			slog.String("error", err.Error()))
		return models.ScheduleResponse{}, fmt.Errorf("%s: %w", operation, err)
	}

	s.logger.Info("schedule was successfully found!",
		slog.String("opertaion", operation),
		slog.Any("schedule", schedule))
	return schedule, nil

}
