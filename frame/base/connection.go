package base

type IConnection interface {
	FindOne(IActiveRecord)
}
