package base

type IConfigDoc interface {
	FileName() string
}

type IConfigure interface {
	EnableReload() bool
	ConfigKey() string
	ConfigDoc() IConfigDoc
	LoadConfig() error
}
