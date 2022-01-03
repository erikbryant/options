package date

import (
	"testing"
	"time"
)

func TestInTradingHours(t *testing.T) {
	testCases := []struct {
		t        time.Time
		expected bool
	}{
		// Sunday
		//   14:29 UTC is 9:29 Eastern
		{time.Date(2022, time.Month(1), 2, 14, 29, 0, 0, time.UTC), false},
		//   14:30 UTC is 9:30 Eastern
		{time.Date(2022, time.Month(1), 2, 14, 30, 0, 0, time.UTC), false},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 2, 20, 59, 0, 0, time.UTC), false},
		//   22:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 2, 22, 0, 0, 0, time.UTC), false},

		// Monday
		//   14:29 UTC is 9:29 Eastern
		{time.Date(2022, time.Month(1), 3, 14, 29, 0, 0, time.UTC), false},
		//   14:30 UTC is 9:30 Eastern
		{time.Date(2022, time.Month(1), 3, 14, 30, 0, 0, time.UTC), true},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 3, 20, 59, 0, 0, time.UTC), true},
		//   22:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 3, 22, 0, 0, 0, time.UTC), false},

		// If Monday and Friday work we will assume Tue-Thu also work

		// Friday
		//   14:29 UTC is 9:29 Eastern
		{time.Date(2022, time.Month(1), 7, 14, 29, 0, 0, time.UTC), false},
		//   14:30 UTC is 9:30 Eastern
		{time.Date(2022, time.Month(1), 7, 14, 30, 0, 0, time.UTC), true},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 7, 20, 59, 0, 0, time.UTC), true},
		//   22:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 7, 22, 0, 0, 0, time.UTC), false},

		// Saturday
		//   14:29 UTC is 9:29 Eastern
		{time.Date(2022, time.Month(1), 8, 14, 29, 0, 0, time.UTC), false},
		//   14:30 UTC is 9:30 Eastern
		{time.Date(2022, time.Month(1), 8, 14, 30, 0, 0, time.UTC), false},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 8, 20, 59, 0, 0, time.UTC), false},
		//   22:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 8, 22, 0, 0, 0, time.UTC), false},
	}

	for _, testCase := range testCases {
		answer := InTradingHours(testCase.t)
		if answer != testCase.expected {
			t.Errorf("For %v expected %t, got %t", testCase.t, testCase.expected, answer)
		}
	}
}

func TestTimeSinceMidnight(t *testing.T) {
	testCases := []struct {
		t        time.Time
		expected time.Duration
	}{
		{time.Time{}, 0},
		{time.Date(2021, time.Month(2), 21, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(2021, time.Month(2), 21, 0, 0, 1, 0, time.UTC), 1 * time.Second},
		{time.Date(2022, time.Month(1), 30, 23, 0, 0, 0, time.UTC), 23 * time.Hour},
		{time.Date(2022, time.Month(1), 30, 23, 59, 59, 0, time.UTC), 23*time.Hour + 59*time.Minute + 59*time.Second},
	}

	for _, testCase := range testCases {
		answer := timeSinceMidnight(testCase.t)
		if answer != testCase.expected {
			t.Errorf("For %v expected %v, got %v", testCase.t, testCase.expected, answer)
		}
	}
}

func TestTimeSinceClose(t *testing.T) {
	testCases := []struct {
		t        time.Time
		expected time.Duration
	}{
		// Monday && hour < 16
		//   5:00 UTC is 0:00 Eastern
		{time.Date(2022, time.Month(1), 3, 5, 0, 0, 0, time.UTC), 56 * time.Hour},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 3, 20, 59, 0, 0, time.UTC), 71*time.Hour + 59*time.Minute},

		// Monday && hour >= 16
		//   21:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 3, 21, 0, 0, 0, time.UTC), 0},
		//   Tuesday 4:59 UTC is Monday 23:59 Eastern
		{time.Date(2022, time.Month(1), 4, 4, 59, 0, 0, time.UTC), 7*time.Hour + 59*time.Minute},

		// Sunday
		//   5:00 UTC is 0:00 Eastern
		{time.Date(2022, time.Month(1), 2, 5, 0, 0, 0, time.UTC), 32 * time.Hour},
		//   9:00 UTC is 4:00 Eastern
		{time.Date(2022, time.Month(1), 2, 9, 0, 0, 0, time.UTC), 36 * time.Hour},
		//   Monday 4:59 UTC is Sunday 23:59 Eastern
		{time.Date(2022, time.Month(1), 3, 4, 59, 0, 0, time.UTC), 55*time.Hour + 59*time.Minute},

		// Saturday
		//   5:00 UTC is 0:00 Eastern
		{time.Date(2022, time.Month(1), 1, 5, 0, 0, 0, time.UTC), 8 * time.Hour},
		//   9:00 UTC is 4:00 Eastern
		{time.Date(2022, time.Month(1), 1, 9, 0, 0, 0, time.UTC), 12 * time.Hour},
		//   Sunday 4:59 UTC is Saturday 23:59 Eastern
		{time.Date(2022, time.Month(1), 2, 4, 59, 0, 0, time.UTC), 31*time.Hour + 59*time.Minute},

		// !Monday && !Sunday && !Saturday && hour < 16
		//   5:00 UTC is 00:00 Eastern
		{time.Date(2022, time.Month(1), 4, 5, 0, 0, 0, time.UTC), 8 * time.Hour},
		//   20:59 UTC is 15:59 Eastern
		{time.Date(2022, time.Month(1), 4, 20, 59, 59, 0, time.UTC), 23*time.Hour + 59*time.Minute + 59*time.Second},

		// !Saturday && !Sunday && hour >= 16
		//   21:00 UTC is 16:00 Eastern
		{time.Date(2022, time.Month(1), 4, 21, 0, 0, 0, time.UTC), 0},
		//   21:00:01 UTC is 16:00:01 Eastern
		{time.Date(2022, time.Month(1), 4, 21, 0, 1, 0, time.UTC), 1 * time.Second},
		//   Wednesday 4:59 UTC is Tuesday 23:59 Eastern
		{time.Date(2022, time.Month(1), 5, 4, 59, 0, 0, time.UTC), 7*time.Hour + 59*time.Minute},
	}

	for _, testCase := range testCases {
		answer := TimeSinceClose(testCase.t)
		if answer != testCase.expected {
			t.Errorf("For %v expected %v, got %v", testCase.t, testCase.expected, answer)
		}
	}
}
