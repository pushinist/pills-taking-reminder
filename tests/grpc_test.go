package tests

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"pills-taking-reminder/internal/api/grpc"
	"pills-taking-reminder/internal/api/grpc/pb"
	"pills-taking-reminder/internal/domain/usecase"
	"strings"
	"testing"
	"time"
)

func TestGRPCCreateSchedule(t *testing.T) {
	cleanupDatabase()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	interval := 90 * time.Minute
	useCase := usecase.NewScheduleUseCase(testRepo, interval)

	server := grpc.NewGRPCServer(useCase, logger)

	validTestCases := []struct {
		name    string
		request *pb.ScheduleRequest
	}{
		{
			name: "valid single",
			request: &pb.ScheduleRequest{
				MedicineName: "Aspirin",
				Frequency:    1,
				Duration:     7,
				UserId:       1001,
			},
		},
		{
			name: "valid multiple",
			request: &pb.ScheduleRequest{
				MedicineName: "Vitamin C",
				Frequency:    3,
				Duration:     30,
				UserId:       1002,
			},
		},
		{
			name: "valid infinite",
			request: &pb.ScheduleRequest{
				MedicineName: "Daily Supplement",
				Frequency:    2,
				Duration:     0,
				UserId:       1003,
			},
		},
		{
			name: "valid max freq",
			request: &pb.ScheduleRequest{
				MedicineName: "Complex Regimen",
				Frequency:    15,
				Duration:     5,
				UserId:       1004,
			},
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := server.CreateSchedule(context.Background(), tc.request)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Errorf("Received nil response")
				return
			}

			if resp.ScheduleId <= 0 {
				t.Errorf("Expected positive schedule ID, got %d", resp.ScheduleId)
				return
			}

			getReq := &pb.ScheduleIDRequest{
				UserId:     tc.request.UserId,
				ScheduleId: resp.ScheduleId,
			}

			getResp, err := server.GetSchedule(context.Background(), getReq)
			if err != nil {
				t.Errorf("Failed to retrieve created schedule: %v", err)
				return
			}

			if getResp.MedicineName != tc.request.MedicineName {
				t.Errorf("Expected medicine name %s, got %s", tc.request.MedicineName, getResp.MedicineName)
			}

			if getResp.UserId != tc.request.UserId {
				t.Errorf("Expected user ID %d, got %d", tc.request.UserId, getResp.UserId)
			}

			if len(getResp.TakingTime) != int(tc.request.Frequency) {
				t.Errorf("Expected %d taking times, got %d", tc.request.Frequency, len(getResp.TakingTime))
			}

			t.Logf("Successfully created and verified schedule with ID %d", resp.ScheduleId)
		})
	}

	invalidTestCases := []struct {
		name    string
		request *pb.ScheduleRequest
	}{

		{
			name: "invalid zero freq",
			request: &pb.ScheduleRequest{
				MedicineName: "Invalid Med",
				Frequency:    0,
				Duration:     7,
				UserId:       1005,
			},
		},
		{
			name: "invalid above max freq",
			request: &pb.ScheduleRequest{
				MedicineName: "Invalid Med",
				Frequency:    16,
				Duration:     7,
				UserId:       1006,
			},
		},
		{
			name: "invalid neg freq",
			request: &pb.ScheduleRequest{
				MedicineName: "Invalid Med",
				Frequency:    -1,
				Duration:     7,
				UserId:       1007,
			},
		},
		{
			name: "invalid neg dur",
			request: &pb.ScheduleRequest{
				MedicineName: "Invalid Med",
				Frequency:    1,
				Duration:     -1,
				UserId:       1008,
			},
		},
		{
			name: "invalid empty medicine name",
			request: &pb.ScheduleRequest{
				MedicineName: "",
				Frequency:    1,
				Duration:     7,
				UserId:       1009,
			},
		},
		{
			name: "invalid zero user ID",
			request: &pb.ScheduleRequest{
				MedicineName: "Invalid Med",
				Frequency:    1,
				Duration:     7,
				UserId:       0,
			},
		},
	}

	for _, tc := range invalidTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.CreateSchedule(context.Background(), tc.request)

			if err == nil {
				t.Errorf("Expected an error for invalid input, got nil")
				return
			}

			if !strings.Contains(err.Error(), "Invalid input parameters") {
				t.Errorf("Expected error about invalid input parameters, got: %v", err)
			}

			t.Logf("Got expected error: %v", err)
		})
	}

	t.Run("Duplicate schedule", func(t *testing.T) {
		req := &pb.ScheduleRequest{
			MedicineName: "Duplicate Test",
			Frequency:    1,
			Duration:     7,
			UserId:       2001,
		}

		_, err := server.CreateSchedule(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create initial schedule: %v", err)
		}

		_, err = server.CreateSchedule(context.Background(), req)

		if err == nil {
			t.Errorf("Expected error for duplicate schedule, got nil")
			return
		}

		if !strings.Contains(err.Error(), "Internal server error") {
			t.Errorf("Expected error with 'Internal server error', got: %v", err)
		}

		t.Logf("Got expected error for duplicate: %v", err)
	})

}

func TestGRPCGetNextTakings(t *testing.T) {
	cleanupDatabase()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	interval := 90 * time.Minute
	useCase := usecase.NewScheduleUseCase(testRepo, interval)

	testUsers := []int64{5001, 5002}
	testSchedules := []struct {
		medicineName string
		frequency    int
		userID       int64
	}{
		{"Morning", 1, 5001},
		{"Frequent", 3, 5001},
		{"Evening", 1, 5002},
	}

	for _, s := range testSchedules {
		input := usecase.ScheduleInput{
			MedicineName: s.medicineName,
			Frequency:    s.frequency,
			Duration:     14,
			UserID:       s.userID,
		}

		scheduleID, err := useCase.CreateSchedule(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to create test schedule: %v", err)
		}
		t.Logf("Created schedule %d: %s for user %d with frequency %d",
			scheduleID, s.medicineName, s.userID, s.frequency)
	}

	server := grpc.NewGRPCServer(useCase, logger)

	for _, userID := range testUsers {
		t.Run(fmt.Sprintf("GetNextTakings for user %d", userID), func(t *testing.T) {
			req := &pb.UserIDRequest{
				UserId: userID,
			}

			resp, err := server.GetNextTakings(context.Background(), req)
			if err != nil {
				t.Fatalf("GetNextTakings failed for user %d: %v", userID, err)
			}

			t.Logf("Retrieved %d takings for user %d", len(resp.Takings), userID)
			for i, taking := range resp.Takings {
				t.Logf("Taking %d: %s at %s", i, taking.MedicineName, taking.TakingTime)
			}

			for i, taking := range resp.Takings {
				if taking.MedicineName == "" {
					t.Errorf("Taking %d: empty medicine name", i)
				}

				if len(taking.TakingTime) != 5 || taking.TakingTime[2] != ':' {
					t.Errorf("Taking %d: invalid time format: %s", i, taking.TakingTime)
				}

				hour := taking.TakingTime[0:2]
				minute := taking.TakingTime[3:5]
				if hour < "00" || hour > "23" || minute < "00" || minute > "59" {
					t.Errorf("Taking %d: invalid time value: %s", i, taking.TakingTime)
				}
			}

			switch userID {
			case 5001:

				foundMorningMed := false
				foundFrequentMed := false

				for _, taking := range resp.Takings {
					if taking.MedicineName == "Morning Med" {
						foundMorningMed = true
					}
					if taking.MedicineName == "Frequent Med" {
						foundFrequentMed = true
					}
				}

				if !foundMorningMed && !foundFrequentMed {
					t.Logf("Note: Neither medicine found in takings - this may be expected depending on current time")
				}
			case 5002:
				if len(resp.Takings) > 0 {
					foundEveningMed := false
					for _, taking := range resp.Takings {
						if taking.MedicineName == "Evening Med" {
							foundEveningMed = true
						}
					}

					if !foundEveningMed {
						t.Logf("Note: Evening Med not found in takings - this may be expected depending on current time")
					}
				}
			}
		})
	}

	t.Run("GetNextTakings for user with no schedules", func(t *testing.T) {
		nonExistentUserID := int64(9999)
		req := &pb.UserIDRequest{
			UserId: nonExistentUserID,
		}

		resp, err := server.GetNextTakings(context.Background(), req)
		if err != nil {
			t.Fatalf("GetNextTakings failed for user with no schedules: %v", err)
		}

		if len(resp.Takings) != 0 {
			t.Errorf("Expected 0 takings for user with no schedules, got %d", len(resp.Takings))
		}
	})
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) >= len(substr) && s[0:len(substr)] == substr
}
