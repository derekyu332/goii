package behaviors

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
)

type IpFilter struct {
	WhiteList []string
	BlackList []string
}

func (this *IpFilter) BeforeAction(c *gin.Context) error {
	clientIp := c.ClientIP()

	if len(this.WhiteList) > 0 {
		for _, ip := range this.WhiteList {
			if clientIp == ip {
				return nil
			}
		}
		logger.Warning("[%v] IP %v not int white list", c.GetInt64(base.KEY_REQUEST_ID), clientIp)
		return base.NotAuthorizedHttpError(c, "IP not in white list")
	} else if len(this.BlackList) > 0 {
		for _, ip := range this.BlackList {
			if clientIp == ip {
				logger.Warning("[%v] IP %v in black list", c.GetInt64(base.KEY_REQUEST_ID), clientIp)
				return base.NotAuthorizedHttpError(c, "IP in black list")
			}
		}

		return nil
	} else {
		return nil
	}
}

func (this *IpFilter) AfterAction(c *gin.Context) error {
	return nil
}
