package formatters

import "github.com/gin-gonic/gin"

type JsonResponseFormatter struct {
}

func (this *JsonResponseFormatter) Format(c *gin.Context, code int, obj interface{}) {
	c.JSON(code, obj)
}
