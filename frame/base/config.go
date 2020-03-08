package base

type IConfigDoc interface {
	FileName() string
}

type IConfigure interface {
	EnableReload() bool
	ConfigDoc() IConfigDoc
	LoadConfig() error
}
