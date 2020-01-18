package validators

import (
	"errors"
	"fmt"
	"github.com/derekyu332/goii/helper/extend"
	"strconv"
	"strings"
)

type NumberValidator struct {
	Values []string
	Ranges []string
	Amongs []int
}

func checkRange(value int, rule string) bool {
	range_rule := strings.Split(rule, "~")

	if len(range_rule) == 2 {
		if range_rule[0] != "" {
			min_value, err := strconv.Atoi(range_rule[0])

			if err == nil && min_value > value {
				return false
			}
		}

		if range_rule[1] != "" {
			max_value, err := strconv.Atoi(range_rule[1])

			if err == nil && max_value < value {
				return false
			}
		}
	}

	return true
}

func (this *NumberValidator) Validate(values map[string]string) error {
	for _, param := range this.Values {
		value, ok := values[param]

		if ok {
			int_value, err := strconv.Atoi(value)

			if err != nil {
				return errors.New(param + " is not number")
			}

			for _, rule := range this.Ranges {
				if !checkRange(int_value, rule) {
					return errors.New(param + " value out of range " + rule)
				}
			}

			if len(this.Amongs) > 0 && extend.InIntArray(int_value, this.Amongs) == -1 {
				return errors.New(param + " value need among " + fmt.Sprintf("%v", this.Amongs))
			}
		}
	}

	return nil
}
