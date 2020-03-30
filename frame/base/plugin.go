package base

import (
	"plugin"
)

func PluginAllocator(path string, name string) interface{} {
	p, err := plugin.Open(path + name)

	if err != nil {
		panic(err)
	}

	n, err := p.Lookup("New")

	if err != nil {
		panic(err)
	}

	fn, ok := n.(func() interface{})

	if ok == false {
		panic("New Func type error")
	}

	return fn()
}
