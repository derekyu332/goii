package base

import "github.com/gin-gonic/gin"

type IActionFilter interface {
	BeforeAction(*gin.Context) error
	AfterAction(*gin.Context) error
}
