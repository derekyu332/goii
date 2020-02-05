package extend

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
