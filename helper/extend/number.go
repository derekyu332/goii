package extend

import "strings"

func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	case float32, float64, complex64, complex128:
		return true
	case string:
		str := val.(string)
		if str == "" {
			return false
		}
		// Trim any whitespace
		str = strings.Trim(str, " \\t\\n\\r\\v\\f")
		if str[0] == '-' || str[0] == '+' {
			if len(str) == 1 {
				return false
			}
			str = str[1:]
		}
		// hex
		if len(str) > 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
			for _, h := range str[2:] {
				if !((h >= '0' && h <= '9') || (h >= 'a' && h <= 'f') || (h >= 'A' && h <= 'F')) {
					return false
				}
			}
			return true
		}
		// 0-9,Point,Scientific
		p, s, l := 0, 0, len(str)
		for i, v := range str {
			if v == '.' { // Point
				if p > 0 || s > 0 || i+1 == l {
					return false
				}
				p = i
			} else if v == 'e' || v == 'E' { // Scientific
				if i == 0 || s > 0 || i+1 == l {
					return false
				}
				s = i
			} else if v < '0' || v > '9' {
				return false
			}
		}
		return true
	}

	return false
}

func InterfaceToInt(inter interface{}) (int, bool) {
	switch inter.(type) {
	case int:
		{
			return inter.(int), true
		}
	case int64:
		{
			return int(inter.(int64)), true
		}
	case float64:
		{
			return int(inter.(float64)), true
		}
	case float32:
		{
			return int(inter.(float32)), true
		}
	}

	return 0, false
}

func InterfaceToInt64(inter interface{}) (int64, bool) {
	switch inter.(type) {
	case int:
		{
			return int64(inter.(int)), true
		}
	case int64:
		{
			return inter.(int64), true
		}
	case float64:
		{
			return int64(inter.(float64)), true
		}
	case float32:
		{
			return int64(inter.(float32)), true
		}
	}

	return 0, false
}

func InterfaceToFloat64(inter interface{}) (float64, bool) {
	switch inter.(type) {
	case int:
		{
			return float64(inter.(int)), true
		}
	case int64:
		{
			return float64(inter.(int64)), true
		}
	case float64:
		{
			return inter.(float64), true
		}
	case float32:
		{
			return float64(inter.(float32)), true
		}
	}

	return 0, false
}
