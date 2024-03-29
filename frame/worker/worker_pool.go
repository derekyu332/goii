package worker

import (
	"github.com/gomodule/redigo/redis"
		"github.com/derekyu332/goii/helper/logger"
	"github.com/gocraft/work"
	"time"
	"github.com/derekyu332/goii/frame/base"
	)

var (
	gWorkerPool *work.WorkerPool
	gServiceName string
	gBaseModule base.IModule
)

const (
	REDIS_MAX_IDLE         = 32
	REDIS_CONNECT_TIME_OUT = 3
	REDIS_READ_TIME_OUT    = 10
	REDIS_WRITE_TIME_OUT   = 5
	REDIS_IDLE_TIME_OUT    = 240
)

type WorketContext struct {

}

func (c *WorketContext) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	logger.Info("Starting job: %v, ID = %v", job.Name, job.ID)
	return next()
}

func InitPool(url string, passowrd string, concurrency int, namespace string) {
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url, redis.DialConnectTimeout(REDIS_CONNECT_TIME_OUT*time.Second),
				redis.DialReadTimeout(REDIS_READ_TIME_OUT*time.Second),
				redis.DialWriteTimeout(REDIS_WRITE_TIME_OUT*time.Second))

			if err != nil {
				logger.Error("DialURL %v error %v", url, err)
				return nil, err
			}

			if _, err := c.Do("AUTH", passowrd); err != nil {
				logger.Error("AUTH error %v", err)
				return nil, err
			}

			return c, err
		},
		MaxActive:   concurrency,
		MaxIdle:     REDIS_MAX_IDLE,
		Wait: true,
	}
	gWorkerPool = work.NewWorkerPoolWithOptions(WorketContext{}, uint(concurrency), namespace,
		redisPool, work.WorkerPoolOptions{SleepBackoffs: []int64{0, 10, 100, 200, 400, 500, 1000}})
	gWorkerPool.Middleware((*WorketContext).Log)
	logger.Warning("WorkerPool Init Success")
}

func Run(serviceName string, module base.IModule) {
	controllers := module.GetControllers()

	for _, controller := range controllers {
		if controller.SupportWorker() {
			var job_pre string

			if serviceName != "" {
				job_pre = "/" + serviceName
			}

			job_pre += "/" + controller.Group() + "/"
			routes := controller.RoutesMap()

			for _, route := range routes {
				job_name := job_pre + route.RelativePath
				logger.Warning("Register Worker(%v): Priv = %v, Retry = %v", job_name, controller.Priority(),
					controller.Retry())
				gWorkerPool.JobWithOptions(job_name, work.JobOptions{
					Priority: uint(controller.Priority()),
					MaxFails: uint(controller.Retry() + 1),
				}, module.RunWorker(serviceName))
			}
		}
	}

	gWorkerPool.Start()
	logger.Warning("WorkerPool Start Success")
}

func StopPool() {
	if gWorkerPool != nil {
		gWorkerPool.Stop()
	}
}
