package config

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/micro/go-micro/config"
	"sync"
)

type TomlConfigure struct {
	Key     string
	Data    base.IConfigDoc
	cfgLock sync.RWMutex
}

func (this *TomlConfigure) EnableReload() bool {
	return true
}

func (this *TomlConfigure) ConfigDoc() base.IConfigDoc {
	this.cfgLock.RLock()
	defer this.cfgLock.RUnlock()
	return this.Data
}

func (this *TomlConfigure) ConfigKey() string {
	return this.Key
}

func (this *TomlConfigure) LoadConfig() error {
	err := config.LoadFile(this.Data.FileName())

	if err != nil {
		logger.Error("config.LoadFile %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	this.cfgLock.Lock()
	defer this.cfgLock.Unlock()
	err = config.Scan(this.Data)

	if err != nil {
		logger.Error("Decode xml %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	logger.Warning("Read xml %v success", this.Data.FileName())
	logger.Warning("%v", this.Data)

	return nil
}
