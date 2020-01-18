package validators

import "errors"

type RequiredValidator struct {
	Values []string
}

func (this *RequiredValidator) Validate(values map[string]string) error {
	for _, param := range this.Values {
		v, ok := values[param]

		if !ok || v == "" {
			return errors.New(param + " required")
		}
	}

	return nil
}
