package base

import (
	"encoding/json"
	"fmt"
	"github.com/derekyu332/goii/frame/formatters"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/url"
	"strconv"
)

type ActionFunc func(*gin.Context) map[string]interface{}

type Route struct {
	HttpMethod   string
	RelativePath string
	Handler      ActionFunc
}

type Worker struct {
	MsgId    int64
	Request  proto.Message
	Response proto.Message
	Handler  ActionFunc
}

type IIdentity interface {
	GetId() interface{}
	Authenticate(c *gin.Context) error
	Authorize(c *gin.Context) error
	IsAuthorized() bool
}

type IRateLimit interface {
	GetRateLimit(c *gin.Context) (int64, int64)
	LoadAllowance(c *gin.Context) (int64, int64)
	SaveAllowance(c *gin.Context, allowance int64)
}

type IController interface {
	PreparedForUse(*gin.Context)
	GetContext() *gin.Context
	Group() string
	TitleRet() string
	TitleMessage() string
	RoutesMap() []Route
	WorkerMap() []Worker
	Behaviors() []IActionFilter
	GetIdentity() IIdentity
	GetRateLimit() IRateLimit
	GetFormatter(*gin.Context) IResponseFormatter
	ParseServiceRequest(string, interface{}) *http.Request
}

type WebController struct {
	UserIdentity IIdentity
	RequestID    int64
	Context      *gin.Context
}

func (this *WebController) PreparedForUse(c *gin.Context) {
	this.UserIdentity = nil
	this.Context = c
	this.RequestID = c.GetInt64(KEY_REQUEST_ID)
}

func (this *WebController) GetContext() *gin.Context {
	return this.Context
}

func (this *WebController) TitleRet() string {
	return "ret"
}

func (this *WebController) TitleMessage() string {
	return "message"
}

func (this *WebController) ParamExist(c *gin.Context, key string) bool {
	_, exists := c.GetQuery(key)

	if !exists {
		_, exists = c.GetPostForm(key)
	}

	return exists
}

func (this *WebController) GetOrPost(c *gin.Context, key string) string {
	value := c.Query(key)

	if value == "" {
		value = c.PostForm(key)
	}

	return value
}

func (this *WebController) GetOrPostInt(c *gin.Context, key string) int {
	value := this.GetOrPost(c, key)

	if int_value, err := strconv.Atoi(value); err == nil {
		return int_value
	} else {
		return 0
	}
}

func (this *WebController) GetOrPostInt64(c *gin.Context, key string) int64 {
	value := this.GetOrPost(c, key)

	if int_value, err := strconv.ParseInt(value, 10, 64); err == nil {
		return int_value
	} else {
		return 0
	}
}

func (this *WebController) GetOrPostFloat64(c *gin.Context, key string) float64 {
	value := this.GetOrPost(c, key)

	if f_value, err := strconv.ParseFloat(value, 64); err == nil {
		return f_value
	} else {
		return 0
	}
}

func (this *WebController) Group() string {
	return ""
}

func (this *WebController) ParseServiceRequest(methodName string, req interface{}) *http.Request {
	httpRequest := &http.Request{
		Method:     "POST",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: map[string][]string{
			"Accept-Encoding": {"gzip, deflate"},
			"Connection":      {"keep-alive"},
			"Cache-Control":   {"max-age=0"},
			"User-Agent":      {"gRPC Service Transfer"},
			"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			"Accept-Language": {"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		},
		Host:       "127.0.0.1:80",
		RemoteAddr: "127.0.0.1:80",
	}

	httpRequest.URL = &url.URL{
		Path: "/" + this.Group() + "/" + methodName + "/",
	}
	httpRequest.RequestURI = httpRequest.URL.Path
	httpRequest.PostForm = make(map[string][]string)
	reqMap := StructToMap(req, "json")

	for key, value := range reqMap {
		httpRequest.PostForm[key] = []string{fmt.Sprintf("%v", value)}
	}

	return httpRequest
}

func (this *WebController) CallAction(action ActionFunc, result interface{}) error {
	c := this.GetContext()

	if c == nil || c.Request == nil {
		return status.Error(codes.Internal, "Invalid Request")
	}

	response := action(c)

	if response != nil {
		ret, ok := response["ret"]

		if ok {
			c.Set(KEY_ACTION_RET, ret)
			errorno := ret.(int)

			if errorno != 0 {
				message := response["message"].(string)

				return status.Error(codes.Code(errorno), message)
			}
		}
	}

	jsonStr, err := json.Marshal(response)

	if err != nil {
		return status.Error(codes.Internal, "Invalid Response")
	}

	err = json.Unmarshal(jsonStr, result)

	if err != nil {
		return status.Error(codes.Internal, "Response Decode Failed")
	}

	return nil
}

func (this *WebController) RoutesMap() []Route {
	return nil
}

func (this *WebController) WorkerMap() []Worker {
	return nil
}

func (this *WebController) Behaviors() []IActionFilter {
	return nil
}

func (this *WebController) GetIdentity() IIdentity {
	return nil
}

func (this *WebController) GetRateLimit() IRateLimit {
	return nil
}

func (this *WebController) GetFormatter(c *gin.Context) IResponseFormatter {
	callback, exists := c.GetQuery("callback")

	if exists && len(callback) > 0 {
		return &formatters.JsonPResponseFormatter{}
	} else {
		return &formatters.JsonResponseFormatter{}
	}
}
