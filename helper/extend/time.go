package extend

import "time"

func IsSameDate(t1 time.Time, t2 time.Time) bool {
	if t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day() {
		return true
	} else {
		return false
	}
}

func UnixTime2TimeStr(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}
