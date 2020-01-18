package base

import (
	"fmt"
	"github.com/derekyu332/goii/frame/i18n"
	"github.com/gin-gonic/gin"
)

const (
	ERR_INVALID_PARA     = 100
	ERR_BAD_REQUEST      = 101
	ERR_NOT_AUTHORIZED   = 102
	ERR_MAINTAINANCE     = 103
	ERR_CONNECTION_LOST  = 104
	ERR_NOT_AUTHENTICATE = 105
	ERR_UNEXPECTED       = 106
	ERR_TIMEOUT          = 107
	ERR_TOO_MANY_REQUEST = 108
	ERR_RELOGIN          = 109
	ERR_RISK_CONTROL     = 110
	ERR_LOCK_FAILED      = 111
)

func E(c *gin.Context, errorno int) string {
	msg := i18n.M(c, fmt.Sprintf("%v", errorno))

	if msg == "" {
		msg = fmt.Sprintf("ErrorNo %v", errorno)
	}

	return msg
}
