package config

import (
	"encoding/xml"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"io/ioutil"
)

type XmlConfigure struct {
	Data base.IConfigDoc
}

func (this *XmlConfigure) EnableReload() bool {
	return true
}

func (this *XmlConfigure) LoadConfig() error {
	content, err := ioutil.ReadFile(this.Data.FileName())

	if err != nil {
		logger.Error("Read xml %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	err = xml.Unmarshal(content, this.Data)

	if err != nil {
		logger.Error("Decode xml %v failed %v", this.Data.FileName(), err.Error())
		return err
	}

	logger.Warning("Read xml %v success", this.Data.FileName())
	logger.Warning("%v", this.Data)

	return nil
}
