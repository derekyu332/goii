package logger

import (
	"github.com/op/go-logging"
	"os"
	"sync"
	"time"
)

var logMain *logging.Logger
var logBill *logging.Logger
var logProfile *logging.Logger
var fileMain *os.File
var fileBill *os.File
var fileProfile *os.File
var normalBackend logging.LeveledBackend
var logTime time.Time
var RELEASE string
var normalLevel logging.Level
var gMapLock sync.RWMutex

const (
	MAIN_MODULE    = "NORMAL"
	BILL_MODULE    = "BILL"
	PROFILE_MODULE = "PROFILE"
)

func init() {
	logMain = logging.MustGetLogger(MAIN_MODULE)
	logBill = logging.MustGetLogger(BILL_MODULE)
	logProfile = logging.MustGetLogger(PROFILE_MODULE)

	if RELEASE == "true" {
		normalLevel = logging.WARNING
		configLogger(true)
	} else {
		normalLevel = logging.INFO
		configLogger(false)
	}

}

func configLogger(release bool) {
	gMapLock.Lock()

	if fileMain != nil {
		fileMain.Close()
	}

	if fileBill != nil {
		fileBill.Close()
	}

	if fileProfile != nil {
		fileProfile.Close()
	}

	logTime = time.Now()
	var normalOutput *logging.LogBackend
	var billOutput *logging.LogBackend
	var profileOutput *logging.LogBackend
	var err error

	if !release {
		normalOutput = logging.NewLogBackend(os.Stdout, "", 0)
		billOutput = logging.NewLogBackend(os.Stdout, "", 0)
		profileOutput = logging.NewLogBackend(os.Stdout, "", 0)
	} else {
		fileMain, err = os.OpenFile("../logs/app_"+logTime.Format("20060102")+".log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			panic(err.Error())
		}

		fileBill, err = os.OpenFile("../bills/bill_"+logTime.Format("20060102")+".log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			panic(err.Error())
		}

		fileProfile, err = os.OpenFile("../bills/profile_"+logTime.Format("20060102")+".log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			panic(err.Error())
		}

		normalOutput = logging.NewLogBackend(fileMain, "", 0)
		billOutput = logging.NewLogBackend(fileBill, "", 0)
		profileOutput = logging.NewLogBackend(fileProfile, "", 0)
	}

	logFormat := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.4s} ▶ %{shortfile} <%{shortfunc}>  %{message}`,
	)
	normalFormatter := logging.NewBackendFormatter(normalOutput, logFormat)
	normalBackend = logging.AddModuleLevel(normalFormatter)
	billFormat := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05}|%{message}`,
	)
	billFormatter := logging.NewBackendFormatter(billOutput, billFormat)
	billBackend := logging.AddModuleLevel(billFormatter)
	profileFormat := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05}|%{message}`,
	)
	profileFormatter := logging.NewBackendFormatter(profileOutput, profileFormat)
	profileBackend := logging.AddModuleLevel(profileFormatter)
	normalBackend.SetLevel(normalLevel, MAIN_MODULE)
	billBackend.SetLevel(logging.INFO, BILL_MODULE)
	profileBackend.SetLevel(logging.INFO, PROFILE_MODULE)
	logMain.SetBackend(normalBackend)
	logMain.ExtraCalldepth = 1
	logBill.SetBackend(billBackend)
	logBill.ExtraCalldepth = 1
	logProfile.SetBackend(profileBackend)
	logProfile.ExtraCalldepth = 1
	gMapLock.Unlock()
}

func Info(format string, args ...interface{}) {
	if logTime.Format("20060102") != time.Now().Format("20060102") {
		if RELEASE == "true" {
			configLogger(true)
		} else {
			configLogger(false)
		}
	}

	logMain.Infof(format, args...)
}

func Warning(format string, args ...interface{}) {
	if logTime.Format("20060102") != time.Now().Format("20060102") {
		if RELEASE == "true" {
			configLogger(true)
		} else {
			configLogger(false)
		}
	}

	logMain.Warningf(format, args...)
}

func Error(format string, args ...interface{}) {
	if logTime.Format("20060102") != time.Now().Format("20060102") {
		if RELEASE == "true" {
			configLogger(true)
		} else {
			configLogger(false)
		}
	}

	logMain.Errorf(format, args...)
}

func SetLevel(level int) {
	normalLevel = logging.Level(level)
	normalBackend.SetLevel(normalLevel, MAIN_MODULE)
	logMain.SetBackend(normalBackend)
}

func Bill(format string, args ...interface{}) {
	if logTime.Format("20060102") != time.Now().Format("20060102") {
		if RELEASE == "true" {
			configLogger(true)
		} else {
			configLogger(false)
		}
	}

	logBill.Infof(format, args...)
}

func Profile(format string, args ...interface{}) {
	if logTime.Format("20060102") != time.Now().Format("20060102") {
		if RELEASE == "true" {
			configLogger(true)
		} else {
			configLogger(false)
		}
	}

	logProfile.Infof(format, args...)
}
