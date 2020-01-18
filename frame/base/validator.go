package base

type IValidator interface {
	Validate(values map[string]string) error
}
