package behaviors

import (
	"fmt"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
	"time"
)

type ActionReporter struct {
	Reporters []base.IReporter
	reqStart  int64
}

func (this *ActionReporter) BeforeAction(c *gin.Context) error {
	this.reqStart = time.Now().UnixNano() / int64(time.Millisecond)
	return nil
}

func (this *ActionReporter) AfterAction(c *gin.Context) error {
	duration := int64(0)

	if this.reqStart > 0 {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		duration = now - this.reqStart
	}

	content := fmt.Sprintf("%v|%v|%v|%v|%v|%v", c.Request.URL.Path, c.GetInt64(base.KEY_REQUEST_ID), duration,
		c.GetInt(base.KEY_ACTION_RET), c.ClientIP(), c.GetString(base.KEY_IDENTITY))
	logger.Profile("ACTION|%v", content)

	for _, reporter := range this.Reporters {
		reporter.Report("ACTION", content)
	}

	return nil
}
