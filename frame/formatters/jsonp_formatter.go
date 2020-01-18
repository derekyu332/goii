package formatters

import "github.com/gin-gonic/gin"

type JsonPResponseFormatter struct {
}

func (this *JsonPResponseFormatter) Format(c *gin.Context, code int, obj interface{}) {
	c.JSONP(code, obj)
}
