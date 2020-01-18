package frame

import (
	"context"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/frame/i18n"
	"github.com/derekyu332/goii/frame/mongo"
	"github.com/derekyu332/goii/frame/rabbit"
	"github.com/derekyu332/goii/frame/redis"
	"github.com/derekyu332/goii/frame/sql"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SqlConfig struct {
	Driver       string
	Uri          string
	MaxIdleConns int
}

type MongoConfig struct {
	Url    string
	DbName string
}

type RedisConfig struct {
	Url      string
	Password string
}

type WebServerConfig struct {
	Address string
}

type RpcServiceConfig struct {
	Address string
}

type RabbitConfig struct {
	AmqpURI        string
	OpenRpc        bool
	RpcQueue       string
	RpcRoutineKey  string
	OpenWorker     bool
	WorkQueue      string
	WorkRoutineKey string
}

type App struct {
	WebInit       *WebServerConfig
	ServiceInit   *RpcServiceConfig
	SqlInit       *SqlConfig
	MongoInit     *MongoConfig
	RedisInit     *RedisConfig
	RabbitInit    *RabbitConfig
	MessageSource []string
	LogLevel      logging.Level
	Components    map[string]base.IComponent
	engine        *gin.Engine
	module        base.IModule
	server        *grpc.Server
}

var (
	gApp       *App
	appConfigs []base.IConfigure
)

func Application() *App {
	return gApp
}

func (this *App) PrepareToRun() error {
	rand.Seed(time.Now().UnixNano())
	gApp = this
	this.engine = gin.New()
	this.engine.Use(gin.Recovery(), limits.RequestSizeLimiter(20<<20))
	m := base.PluginAllocator("module.so")
	this.module, _ = m.(base.IModule)
	this.module.SetEngine(this.engine)
	logger.SetLevel((int)(this.LogLevel))

	if this.SqlInit != nil {
		sql.InitEngine(this.SqlInit.Driver, this.SqlInit.Uri, this.SqlInit.MaxIdleConns)
	}

	if this.MongoInit != nil {
		mongoConfig, err := mgo.ParseURL(this.MongoInit.Url)

		if err != nil {
			panic(err)
		}

		mongoConfig.PoolTimeout = 3000 * time.Millisecond
		mongoConfig.ReadTimeout = 10000 * time.Millisecond
		mongoConfig.WriteTimeout = 5000 * time.Millisecond
		mongoConfig.Safe = mgo.Safe{
			W:        1,
			WTimeout: 5000,
		}
		mongo.InitConnection(mongoConfig, this.MongoInit.DbName)
	}

	if this.MessageSource != nil {
		i18n.InitBundle(this.MessageSource)
	}

	if this.RedisInit != nil {
		redis.InitConnection(this.RedisInit.Url, this.RedisInit.Password)
	}

	if this.RabbitInit != nil {
		rabbit.InitProducer(this.RabbitInit.AmqpURI)

		if this.RabbitInit.OpenRpc {
			rabbit.InitConsumer(this.RabbitInit.AmqpURI, this.RabbitInit.RpcQueue, this.RabbitInit.RpcRoutineKey)
		}

		if this.RabbitInit.OpenWorker {
			rabbit.InitWorker(this.RabbitInit.AmqpURI, this.RabbitInit.WorkQueue, this.RabbitInit.WorkRoutineKey,
				this.module.RunWorker())
		}
	}

	if this.Components != nil {
		for _, com := range this.Components {
			if err := com.Initialize(); err != nil {
				panic(err)
			}
		}
	}

	if this.ServiceInit != nil {
		customInterceptor := this.module.RunService()
		this.server = grpc.NewServer(grpc.UnaryInterceptor(customInterceptor))
	}

	return nil
}

func (this *App) GetServer() *grpc.Server {
	return this.server
}

func (this *App) GetEngine() *gin.Engine {
	return this.engine
}

func (this *App) GetComponent(comId string) base.IComponent {
	com, ok := this.Components[comId]

	if ok {
		return com
	} else {
		return nil
	}
}

func (this *App) RegisterControllers(controllers []base.IController) {
	this.module.SetControllers(controllers)

	for _, controller := range controllers {
		name := controller.Group()
		routes := controller.RoutesMap()

		if name != "" {
			group := this.engine.Group(name)

			for _, route := range routes {
				logger.Warning("Register Route(%v) = %v", name, route)
				group.Handle(route.HttpMethod, route.RelativePath,
					this.module.RunAction(controller, route.RelativePath))
			}
		} else {
			for _, route := range routes {
				logger.Warning("Register Route = %v", route)
				this.engine.Handle(route.HttpMethod, route.RelativePath,
					this.module.RunAction(controller, route.RelativePath))
			}
		}
	}
}

func LoadAllConfigs(configs []base.IConfigure) {
	for _, config := range configs {
		if err := config.LoadConfig(); err != nil {
			panic(err)
		}

		appConfigs = append(appConfigs, config)
	}
}

func ReloadAllConfigs() {
	for _, config := range appConfigs {
		if config.EnableReload() {
			if err := config.LoadConfig(); err != nil {
				logger.Warning("Reload config failed %v", err.Error())
			}
		}
	}
}

func (this *App) Run() {
	var http_server *http.Server

	if this.WebInit != nil {
		http_server = &http.Server{Addr: this.WebInit.Address, Handler: this.engine}

		go func() {
			if err := http_server.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
	}

	if this.ServiceInit != nil {
		lis, err := net.Listen("tcp", this.ServiceInit.Address)

		if err != nil {
			panic(err)
		}

		go func() {
			if err := this.server.Serve(lis); err != nil {
				panic(err)
			}
		}()
	}

	logger.Warning("Application Start Success")
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2)

	for s := range ch {
		switch s {
		case syscall.SIGUSR1:
			{
				ReloadAllConfigs()
			}
		case syscall.SIGUSR2:
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if http_server != nil {
					if err := http_server.Shutdown(ctx); err != nil {
						log.Fatal("Server Shutdown:", err)
					}
				}

				logger.Warning("Application Shutdown Success")
				os.Exit(0)
			}
		}
	}
}
