package date

import (
	"time"
)

// InTradingHours returns true if the market is open at time t (ignoring holidays).
func InTradingHours(t time.Time) bool {
	// The market runs on Eastern time.
	eastern, _ := time.LoadLocation("America/New_York")
	et := t.In(eastern)

	// The market is not open on weekends.
	weekday := et.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// The market is open from 9:30am to 4pm.
	hour := et.Hour()
	min := et.Minute()
	if hour < 9 || (hour == 9 && min < 30) || hour >= 16 {
		return false
	}

	// TODO: Check for market holiday.

	return true
}

// timeSinceMidnight returns the number of seconds that have elapsed since midnight of the given day.
func timeSinceMidnight(t time.Time) time.Duration {
	year, month, day := t.Date()
	t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return t.Sub(t2)
}

// TimeSinceClose returns the time between the last market close and the given time (ignoring holidays).
func TimeSinceClose(t time.Time) time.Duration {
	// The market runs on Eastern time.
	eastern, _ := time.LoadLocation("America/New_York")
	et := t.In(eastern)

	hour := et.Hour()
	weekday := et.Weekday()

	if weekday == time.Monday && hour < 16 {
		// Use Friday's close.
		// 4pm - midnight == 8 hours.
		return 8*time.Hour + 24*time.Hour + 24*time.Hour + timeSinceMidnight(et)
	}

	if weekday == time.Sunday {
		// Use Friday's close.
		return 8*time.Hour + 24*time.Hour + timeSinceMidnight(et)
	}

	if weekday == time.Saturday || hour < 16 {
		// Use yesterday's close.
		return 8*time.Hour + timeSinceMidnight(et)
	}

	if hour >= 16 {
		// The market has closed today. Use 4pm today.
		return timeSinceMidnight(et) - 16*time.Hour
	}

	// TODO: Check for market holiday.

	return 0
}
