package behaviors

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/extend"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
)

type CookieFilter struct {
	Values []string
	Except []string
}

func (this *CookieFilter) BeforeAction(c *gin.Context) error {
	if this.Except != nil && extend.InStringArray(c.Request.URL.Path, this.Except) >= 0 {
		return nil
	}

	for _, param := range this.Values {
		cookie, err := c.Request.Cookie(param)

		if err != nil || cookie.Value == "" {
			return base.InvalidParaHttpError(c, err.Error())
		} else {
			logger.Info("[%v] Cookie[%v] = %v", c.GetInt64(base.KEY_REQUEST_ID), param, cookie.Value)
		}
	}

	return nil
}

func (this *CookieFilter) AfterAction(c *gin.Context) error {
	return nil
}
