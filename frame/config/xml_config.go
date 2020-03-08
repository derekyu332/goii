package config

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/micro/go-micro/config"
	"sync"
)

type XmlConfigure struct {
	Data    base.IConfigDoc
	cfgLock sync.RWMutex
}

func (this *XmlConfigure) EnableReload() bool {
	return true
}

func (this *XmlConfigure) ConfigDoc() base.IConfigDoc {
	this.cfgLock.RLock()
	defer this.cfgLock.RUnlock()
	return this.Data
}

func (this *XmlConfigure) LoadConfig() error {
	err := config.Load(this.Data.FileName())

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
