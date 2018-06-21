package misc

import (
	"time"
)

// return tomorrow 00:00 am
func NextDay00() time.Time {
	n := time.Now()
	//	n, _ = time.Parse(time.RFC3339, "2018-12-31T12:34:56+07:00")
	return time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, n.Location())
}

// return next monday 00:00 am
func NextWeek00() time.Time {
	n := time.Now()
	for i := 1; i <= 7; i++ {
		day00 := time.Date(n.Year(), n.Month(), n.Day()+i, 0, 0, 0, 0, n.Location())
		if day00.Weekday() == time.Monday {
			return day00
		}
	}
	// unreachable code
	return n
}

// return 1st day of the next month
func NextMonth00() time.Time {
	n := time.Now()
	//	n, _ = time.Parse(time.RFC3339, "2018-12-25T12:34:56+07:00")
	for i := 1; i <= 31; i++ {
		day00 := time.Date(n.Year(), n.Month(), n.Day()+i, 0, 0, 0, 0, n.Location())
		if day00.Day() == 1 {
			return day00
		}
	}
	// unreachable code
	return n
}
