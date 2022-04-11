package base

import (
	"github.com/derekyu332/goii/frame/ether"
	"github.com/derekyu332/goii/frame/kafka"
		"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

const KEY_REQUEST_ID = "KEY_REQUEST_ID"
const KEY_ACTION_RET = "KEY_ACTION_RET"
const KEY_RESPONSE = "KEY_RESPONSE"
const KEY_IDENTITY = "KEY_IDENTITY"

type IModule interface {
	SetEngine(eg *gin.Engine)
	GetControllers() []IController
	SetControllers(controllers []IController)
	RunService() grpc.UnaryServerInterceptor
	RunWorker(serviceName string) interface{}
	RunPoll() kafka.KafkaHandler
	RunEther() ether.EtherHandler
	RunAction(regController IController, relativePath string) func(*gin.Context)
}
