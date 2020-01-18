package extend

import (
	"strconv"
	"strings"
)

func InStringArray(value string, array []string) int {
	for i, v := range array {
		if v == value {
			return i
		}
	}

	return -1
}

func InIntArray(value int, array []int) int {
	for i, v := range array {
		if v == value {
			return i
		}
	}

	return -1
}

func InInt64Array(value int64, array []int64) int {
	for i, v := range array {
		if v == value {
			return i
		}
	}

	return -1
}

func ExplodeInt64(value string, seperator string) []int64 {
	strs := strings.Split(value, seperator)
	var results []int64

	for _, s := range strs {
		int_value, _ := strconv.ParseInt(s, 10, 64)
		results = append(results, int_value)
	}

	return results
}
