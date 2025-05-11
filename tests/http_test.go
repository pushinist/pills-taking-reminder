package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	httpHandler "pills-taking-reminder/internal/api/http"
	"pills-taking-reminder/pkg/logger"

	"pills-taking-reminder/internal/api/dto"
	"pills-taking-reminder/internal/domain/entities"
	"pills-taking-reminder/internal/domain/usecase"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestCreateScheduleHTTP(t *testing.T) {
	cleanupDatabase()

	tests := []struct {
		name           string
		request        dto.ScheduleRequest
		wantStatusCode int
		wantError      bool
	}{
		{
			name: "Just valid frequency",
			request: dto.ScheduleRequest{
				MedicineName: "Aspirin",
				Frequency:    2,
				Duration:     7,
				UserID:       1001,
			},
			wantStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name: "Minimum frequency",
			request: dto.ScheduleRequest{
				MedicineName: "Ibuprofen",
				Frequency:    1,
				Duration:     30,
				UserID:       1002,
			},
			wantStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name: "Maximum frequency",
			request: dto.ScheduleRequest{
				MedicineName: "Ingavirin",
				Frequency:    15,
				Duration:     5,
				UserID:       1003,
			},
			wantStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name: "Infinite duration",
			request: dto.ScheduleRequest{
				MedicineName: "Jelly bears",
				Frequency:    1,
				Duration:     0,
				UserID:       1004,
			},
			wantStatusCode: http.StatusOK,
			wantError:      false,
		},
	}

	logger := logger.SetupLogger("local")
	useCase := usecase.NewScheduleUseCase(testRepo, 90*time.Minute)
	handler := httpHandler.NewScheduleHandler(useCase, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupDatabase()

			reqJSON, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			resp, err := http.Post(
				server.URL+"/schedule",
				"application/json",
				bytes.NewBuffer(reqJSON),
			)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, resp.StatusCode)
				return
			}

			if !tt.wantError {
				var scheduleID int64
				err = json.NewDecoder(resp.Body).Decode(&scheduleID)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if scheduleID <= 0 {
					t.Errorf("Expected positive schedule ID, got %d", scheduleID)
				}

				getURL := fmt.Sprintf("%s/schedule?user_id=%d&schedule_id=%d",
					server.URL, tt.request.UserID, scheduleID)
				getResp, err := http.Get(getURL)
				if err != nil {
					t.Fatalf("Failed to retrieve schedule: %v", err)
				}
				defer getResp.Body.Close()

				if getResp.StatusCode != http.StatusOK {
					t.Errorf("Expected to retrieve schedule with status 200, got %d", getResp.StatusCode)
					return
				}

				var schedule dto.ScheduleResponse
				err = json.NewDecoder(getResp.Body).Decode(&schedule)
				if err != nil {
					t.Fatalf("Failed to decode schedule: %v", err)
				}

				if schedule.MedicineName != tt.request.MedicineName {
					t.Errorf("Expected medicine name %s, got %s",
						tt.request.MedicineName, schedule.MedicineName)
				}

				if schedule.UserID != tt.request.UserID {
					t.Errorf("Expected user ID %d, got %d",
						tt.request.UserID, schedule.UserID)
				}

				if len(schedule.TakingTime) != tt.request.Frequency {
					t.Errorf("Expected %d taking times, got %d",
						tt.request.Frequency, len(schedule.TakingTime))
				}
			}
		})
	}
}

func TestScheduleGetNextTakingsHTTP(t *testing.T) {
	now := time.Date(2025, 5, 11, 14, 0, 0, 0, time.UTC)
	interval := 2 * time.Hour

	tests := []struct {
		name      string
		schedule  *entities.Schedule
		wantCount int
		wantTimes []string
	}{
		{
			name: "single",
			schedule: &entities.Schedule{
				MedicineName: "Test Med",
				Frequency:    1,
				TakingTimes: []entities.TakingTime{
					{Time: time.Date(0, 0, 0, 15, 0, 0, 0, time.UTC)},
				},
				StartDate: now.AddDate(0, 0, -1),
			},
			wantCount: 1,
			wantTimes: []string{"15:00"},
		},
		{
			name: "Multiple but only one should be",
			schedule: &entities.Schedule{
				MedicineName: "Test Med",
				Frequency:    3,
				TakingTimes: []entities.TakingTime{
					{Time: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC)},
					{Time: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},
					{Time: time.Date(0, 0, 0, 22, 0, 0, 0, time.UTC)},
				},
				StartDate: now.AddDate(0, 0, -1),
			},
			wantCount: 1,
			wantTimes: []string{"15:30"},
		},
		{
			name: "multiple but zero should be",
			schedule: &entities.Schedule{
				MedicineName: "Test Med",
				Frequency:    2,
				TakingTimes: []entities.TakingTime{
					{Time: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC)},
					{Time: time.Date(0, 0, 0, 22, 0, 0, 0, time.UTC)},
				},
				StartDate: now.AddDate(0, 0, -1),
			},
			wantCount: 0,
			wantTimes: []string{},
		},
		{
			name: "tomorrow",
			schedule: &entities.Schedule{
				MedicineName: "Test Med",
				Frequency:    1,
				TakingTimes: []entities.TakingTime{
					{Time: time.Date(0, 0, 0, 15, 0, 0, 0, time.UTC)},
				},
				StartDate: now.AddDate(0, 0, 1),
			},
			wantCount: 0,
			wantTimes: []string{},
		},
		{
			name: "past",
			schedule: &entities.Schedule{
				MedicineName: "Test Med",
				Frequency:    1,
				TakingTimes: []entities.TakingTime{
					{Time: time.Date(0, 0, 0, 15, 0, 0, 0, time.UTC)},
				},
				StartDate: now.AddDate(0, 0, -10),
				EndDate:   func() *time.Time { t := now.AddDate(0, 0, -1); return &t }(),
			},
			wantCount: 0,
			wantTimes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			takings := tt.schedule.GetNextTakings(now, interval)

			if len(takings) != tt.wantCount {
				t.Errorf("Expected %d takings, got %d", tt.wantCount, len(takings))
			}

			for i, want := range tt.wantTimes {
				if i < len(takings) {
					got := takings[i].FormatTime()
					if got != want {
						t.Errorf("Taking %d: expected %s, got %s", i, want, got)
					}
				}
			}
		})
	}
}
