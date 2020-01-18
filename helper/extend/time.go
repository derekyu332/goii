package extend

import "time"

func IsSameDate(t1 time.Time, t2 time.Time) bool {
	if t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day() {
		return true
	} else {
		return false
	}
}
