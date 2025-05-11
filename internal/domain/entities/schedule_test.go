package entities_test

import (
	"pills-taking-reminder/internal/domain/entities"
	"testing"
	"time"
)

func TestCalculateTakingTimes(t *testing.T) {
	createTime := func(hour, minute int) time.Time {
		return time.Date(0, 0, 0, hour, minute, 0, 0, time.UTC)
	}

	assertTimesEqual := func(t *testing.T, expected, actual []entities.TakingTime) {
		t.Helper()
		if len(expected) != len(actual) {
			t.Fatalf("expected %d taking times, got %d", len(expected), len(actual))
		}
		for i, expTime := range expected {
			actTime := actual[i]
			if expTime.Time.Hour() != actTime.Time.Hour() || expTime.Time.Minute() != actTime.Time.Minute() {
				t.Errorf("taking time on index %d wrong: expected %02d:%02d, got %02d:%02d", i, expTime.Time.Hour(), expTime.Time.Minute(), actTime.Time.Hour(), actTime.Time.Minute())
			}
		}
	}

	tests := []struct {
		name      string
		frequency int
		expected  []entities.TakingTime
		wantErr   bool
	}{
		{
			name:      "Single",
			frequency: 1,
			expected: []entities.TakingTime{
				{Time: createTime(15, 0)},
			},
		},
		{
			name:      "Twice",
			frequency: 2,
			expected: []entities.TakingTime{
				{Time: createTime(8, 0)},
				{Time: createTime(22, 0)},
			},
		},
		{
			name:      "Thrice",
			frequency: 3,
			expected: []entities.TakingTime{
				{Time: createTime(8, 0)},
				{Time: createTime(15, 0)},
				{Time: createTime(22, 0)},
			},
		},
		{
			name:      "Four times",
			frequency: 4,
			expected: []entities.TakingTime{
				{Time: createTime(8, 0)},
				{Time: createTime(12, 45)},
				{Time: createTime(17, 15)},
				{Time: createTime(22, 0)},
			},
		},
		{
			name:      "Maximum",
			frequency: 15,
			expected: []entities.TakingTime{
				{Time: createTime(8, 0)},
				{Time: createTime(9, 0)},
				{Time: createTime(10, 0)},
				{Time: createTime(11, 0)},
				{Time: createTime(12, 0)},
				{Time: createTime(13, 0)},
				{Time: createTime(14, 0)},
				{Time: createTime(15, 0)},
				{Time: createTime(16, 0)},
				{Time: createTime(17, 0)},
				{Time: createTime(18, 0)},
				{Time: createTime(19, 0)},
				{Time: createTime(20, 0)},
				{Time: createTime(21, 0)},
				{Time: createTime(22, 0)},
			},
		},

		{
			name:      "Zero frequency",
			frequency: 0,
			wantErr:   true,
		},
		{
			name:      "Negative frequency",
			frequency: -1,
			wantErr:   true,
		},
		{
			name:      "Above maximum frequency",
			frequency: 16,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entities.CalculateTakingTimes(tt.frequency)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateTakingTimes(%d) expected an error but got nil", tt.frequency)
					return
				}
			} else {
				if err != nil {
					t.Errorf("CalculateTakingTimes(%d) got unexpected error: %v", tt.frequency, err)
				}
			}

			assertTimesEqual(t, tt.expected, got)
		})
	}
}
