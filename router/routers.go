package router

import (
	"github.com/kataras/iris"
	"github.com/levigross/grequests"
	"io/ioutil"
	"encoding/json"
	"time"
	"guul/eureka/retMessageBody"
	"guul/filter"
	"strings"
	"guul/eureka/discovery"
	"guul/eureka/conf"
	"github.com/sadlil/go-trigger"
	"log"
	"github.com/kataras/iris/middleware/basicauth"
	"os"
)

const (
	SRVNOTFOUND      = "GUUL: 调用的服务不可用,该服务没有注册"
	SRVINTERNALERROR = "GUUL: 调用的服务内部服务器错误"
	SRVPATHNOTFOUND  = "GUUL: 调用的路由没有找到"
	SRVForbidden        = "GUUL: 调用的服务禁止访问"
	SRVBAD           = "GUUL: 调用的服务503BAD GETEWAY"
	NOTSUPPORTMETHOD ="GUUL: 不支持的请求动作，目前只支持[GET,POST]"
)

var (
	eurekaConf *eureka.EurekaConf
	routers    []map[string]string
	ret        *retMessageBody.RetMessage
	serviceUrl string

)

func init() {
	eurekaConf = eurekaConf.GetEurekaConf()
	routers = eurekaConf.GuulRouter
	trigger.On("refresh-router-conf", func() {
		log.Println("刷新路由配置")
		eurekaConf = eurekaConf.GetEurekaConf()
		routers = eurekaConf.GuulRouter
		eurekaConf.ShowAllEurekaConf()
	})
	eurekaConf.ShowAllEurekaConf()
	ret = &retMessageBody.RetMessage{Status: 20001, Result: &retMessageBody.MessageBody{Messsage: "未知错误"}}
}

func RunRouter(app *iris.Application, filter filter.Filter) {

	manageUser:=os.Getenv("MANAGE-USER")
		if manageUser==""{
			manageUser="lgphp"
		}
	managePass :=os.Getenv("MANAGE-PASS")
	if managePass==""{
		managePass = "52cx.comqazxc"
	}
	authConfig := basicauth.Config{
		Users:   map[string]string{manageUser: managePass},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(30) * time.Minute,
	}
	authentication := basicauth.New(authConfig)


	app.Get("/refresh", authentication,func(context iris.Context) {
		trigger.Fire("refresh-router-conf")
		trigger.Fire("reg-service")
		trigger.Fire("send-beat")
		app.RefreshRouter()  //刷新路由
		trigger.Fire("refresh-router")
		context.JSON(iris.Map{"message": "OK"})
	})

	app.OnErrorCode(iris.StatusNotFound, func(context iris.Context) {
		ret.Status = iris.StatusNotFound
		ret.Result.Messsage = SRVPATHNOTFOUND
		ret.Result.Data = make(map[string]string)
		context.JSON(ret)
	})
	app.OnErrorCode(iris.StatusInternalServerError, func(context iris.Context) {
		ret.Status = iris.StatusInternalServerError
		ret.Result.Messsage = "GUUL 网关内部服务器错误"
		ret.Result.Data = make(map[string]string)
		context.JSON(ret)
	})

	app.Get("/info", func(context iris.Context) {

		context.JSON(iris.Map{"Description": "GUUL GateWay By GOLang1.92"})
	})

	app.Get("/health", func(context iris.Context) {

		context.JSON(iris.Map{"status": "UP"})
	})

	for _, v := range routers {
		app.Any(v["path"], filter.PreHandler(), func(context iris.Context) {
			method := context.Method()
			Header := context.Request().Header
			newHeaders := make(map[string]string)
			for i, v := range Header {
				newHeaders[i] = v[0]
			}
			newHeaders["Access-Control-Allow-Origin"] = "*" //允许跨域访问
			path := context.Path()
			for _, srv := range routers {
				if strings.HasPrefix(path, string([]rune(srv["path"])[:len(srv["path"])-1])) {
					serviceUrl = strings.Join([]string{discovery.GetServiceBaseUrl(srv["srvId"]),
						string([]rune(path)[len(srv["path"])-1:])}, "")
				}
			}

			ret.Result.Data = map[string]string{}
			if !strings.HasPrefix(serviceUrl, "http://") {
				ret.Status = iris.StatusNotFound
				ret.Result.Messsage = SRVNOTFOUND
				context.JSON(ret)
				return
			}

			switch method {
			case "GET":
				params := context.URLParams()
				ro := &grequests.RequestOptions{
					Params:         params,
					Headers:        newHeaders,
					RequestTimeout: 5 * time.Second,
				}
				//判断如果mimeheader是否是JSON
				res, err := grequests.Get(serviceUrl, ro)
				switch {
				case err !=nil :
					ret.Status = iris.StatusRequestTimeout
					ret.Result.Messsage = "请求超时或者其他错误" + err.Error()
					context.JSON(ret)
				case res.StatusCode == iris.StatusInternalServerError:
					ret.Status = iris.StatusInternalServerError
					ret.Result.Messsage = SRVINTERNALERROR
					context.JSON(ret)
				case res.StatusCode == iris.StatusNotFound:
					ret.Status = iris.StatusNotFound
					ret.Result.Messsage = SRVPATHNOTFOUND
					context.JSON(ret)
				case res.StatusCode == iris.StatusBadGateway:
					ret.Status = iris.StatusBadGateway
					ret.Result.Messsage = SRVBAD
					context.JSON(ret)
				case res.StatusCode == iris.StatusForbidden:
					ret.Status = 3000211
					ret.Result.Messsage = SRVForbidden
					context.JSON(ret)
				case res.Ok:
					m := make(map[string]interface{})
					json.Unmarshal(res.Bytes(), &m)
					if m == nil {
						context.JSON(m)
					} else {
						context.Write(res.Bytes())
					}
				}

			case "POST":
				params := context.URLParams()
				postdata := make(map[string]string)
				for k, v := range context.FormValues() {
					postdata[k] = v[0]
				}
				ro := &grequests.RequestOptions{
					Params:         params,
					Headers:        newHeaders,
					Data:           postdata,
					RequestTimeout: 5 * time.Second,
				}
				rawJson, _ := ioutil.ReadAll(context.Request().Body)
				if len(string(rawJson)) > 0 {
					ro.JSON = string(rawJson)
				}
				res, err := grequests.Post(serviceUrl, ro)
				ret := retMessageBody.RetMessage{Result: &retMessageBody.MessageBody{}}
				ret.Result.Data = map[string]string{}
				switch {
				case err !=nil :
					ret.Status = iris.StatusRequestTimeout
					ret.Result.Messsage = "GUUL:调用请求超时或者其他错误" + err.Error()
					context.JSON(ret)
				case res.StatusCode == iris.StatusInternalServerError:
					ret.Status = iris.StatusInternalServerError
					ret.Result.Messsage = SRVINTERNALERROR
					context.JSON(ret)
				case res.StatusCode == iris.StatusNotFound:
					ret.Status = iris.StatusNotFound
					ret.Result.Messsage = SRVPATHNOTFOUND
					context.JSON(ret)
				case res.StatusCode == iris.StatusBadGateway:
					ret.Status = iris.StatusBadGateway
					ret.Result.Messsage = SRVBAD
					context.JSON(ret)
				case res.StatusCode == iris.StatusForbidden:
					ret.Status = iris.StatusForbidden
					ret.Result.Messsage = SRVForbidden
					context.JSON(ret)
				case res.Ok:
					m := make(map[string]interface{})
					json.Unmarshal(res.Bytes(), &m)
					if m == nil {
						context.JSON(m)
					} else {
						context.Write(res.Bytes())
					}
				}

			default:
				ret := retMessageBody.RetMessage{Result: &retMessageBody.MessageBody{}}
				ret.Status = iris.StatusMethodNotAllowed
				ret.Result.Messsage =  NOTSUPPORTMETHOD
				ret.Result.Data = map[string]string{}
				context.JSON(ret)
			}
		})
	}
}
