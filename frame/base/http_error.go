package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type HttpError struct {
	Code    int
	Ret     int
	Message string
}

func (this *HttpError) Error() string {
	return fmt.Sprintf("ret:%d, message:%v", this.Ret, this.Message)
}

func CustomHttpError(c *gin.Context, ret int, message string) *HttpError {
	if message == "" {
		return &HttpError{Code: 200, Ret: ret, Message: E(c, ret)}
	} else {
		return &HttpError{Code: 200, Ret: ret, Message: message}
	}
}

func InvalidParaHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_BAD_REQUEST, message)
}

func BadRequestHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_BAD_REQUEST, message)
}

func NotAuthorizedHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_NOT_AUTHORIZED, message)
}

func MaintainenceHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_MAINTAINANCE, message)
}

func ConnectionLostHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_CONNECTION_LOST, message)
}

func NotAuthenticatedHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_NOT_AUTHENTICATE, message)
}

func ServerUnexpectedHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_UNEXPECTED, message)
}

func TimeoutHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_TIMEOUT, message)
}

func TooManyRequestHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_TOO_MANY_REQUEST, message)
}

func ReloginHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_RELOGIN, message)
}

func RiskControlHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_RISK_CONTROL, message)
}

func LockActionHttpError(c *gin.Context, message string) *HttpError {
	return CustomHttpError(c, ERR_LOCK_FAILED, message)
}
