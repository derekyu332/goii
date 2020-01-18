package base

type IReporter interface {
	Report(title string, content string)
}
