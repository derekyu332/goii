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

func GetTodayZero() int64 {
	t := time.Now()
	now_time := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	return now_time
}

func GetYesterdayZero() int64 {
	t := time.Now()
	now_time := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	return now_time - 86400
}

func GetMondayZero() int64 {
	t := time.Now()
	offset := int(time.Monday - t.Weekday())

	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	return weekStartDate.Unix()
}

func GetMonth1stZero() int64 {
	t := time.Now()
	monthStartDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)

	return monthStartDate.Unix()
}

func GetLastMonth1stZero() int64 {
	t := time.Now()
	monthStartDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, -1, 0)

	return monthStartDate.Unix()
}
