package base

import (
	"github.com/derekyu332/goii/frame/kafka"
	"github.com/derekyu332/goii/frame/rabbit"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

const KEY_REQUEST_ID = "KEY_REQUEST_ID"
const KEY_ACTION_RET = "KEY_ACTION_RET"
const KEY_RESPONSE = "KEY_RESPONSE"
const KEY_IDENTITY = "KEY_IDENTITY"

type IModule interface {
	SetEngine(eg *gin.Engine)
	SetControllers(controllers []IController)
	RunService() grpc.UnaryServerInterceptor
	RunWorker() rabbit.RabbitHandler
	RunPoll() kafka.KafkaHandler
	RunAction(regController IController, relativePath string) func(*gin.Context)
}
