package base

import "github.com/gin-gonic/gin"

type IResponseFormatter interface {
	Format(c *gin.Context, code int, obj interface{})
}
