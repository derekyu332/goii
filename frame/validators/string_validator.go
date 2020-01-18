package validators

import (
	"errors"
	"strconv"
)

type StringValidator struct {
	Values []string
	Max    int
	Min    int
	Length int
}

func (this *StringValidator) Validate(values map[string]string) error {
	for _, param := range this.Values {
		value, ok := values[param]

		if ok {
			if this.Length != 0 && len(value) != this.Length {
				return errors.New(param + " length must be " + strconv.Itoa(this.Length))
			} else if this.Max != 0 && len(value) > this.Max {
				return errors.New(param + " max-length = " + strconv.Itoa(this.Max))
			} else if this.Min != 0 && len(value) < this.Min {
				return errors.New(param + " min-length = " + strconv.Itoa(this.Min))
			}
		}
	}

	return nil
}
