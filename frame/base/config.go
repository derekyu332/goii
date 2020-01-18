package base

type IConfigDoc interface {
	FileName() string
}

type IConfigure interface {
	EnableReload() bool
	LoadConfig() error
}
