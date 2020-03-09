package config

import (
	"encoding/xml"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"io/ioutil"
	"sync"
)

type XmlConfigure struct {
	Key     string
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

func (this *XmlConfigure) ConfigKey() string {
	return this.Key
}

func (this *XmlConfigure) LoadConfig() error {
	content, err := ioutil.ReadFile(this.Data.FileName())

	if err != nil {
		logger.Error("Read xml %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	this.cfgLock.Lock()
	defer this.cfgLock.Unlock()
	err = xml.Unmarshal(content, this.Data)

	if err != nil {
		logger.Error("Decode xml %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	logger.Warning("Read xml %v success", this.Data.FileName())
	logger.Warning("%v", this.Data)

	return nil
}
