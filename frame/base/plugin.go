package base

import (
	"go/build"
	"plugin"
)

func PluginAllocator(name string) interface{} {
	p, err := plugin.Open(build.Default.GOPATH + "/src/github.com/derekyu332/plugins/" + name)

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
