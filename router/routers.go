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
)

const (
	SRVNOTFOUND      = "服务不可用,该服务没有注册"
	SRVINTERNALERROR = "服务内部服务器错误"
	SRVPATHNOTFOUND  = "路由没有找到"
	SRVNOPOST        = "服务禁止访问"
	SRVBAD           = "服务501BAD GETEWAY"
)

var (
	eurekaConf *eureka.EurekaConf
	routers    []map[string]string
	ret        *retMessageBody.RetMessage
 	serviceUrl string
	res *grequests.Response
)

func init() {
	eurekaConf = eurekaConf.GetEurekaConf()
	routers =eurekaConf.GuulRouter
	trigger.On("refresh-router-conf", func() {
		log.Println("刷新路由配置")
		eurekaConf = eurekaConf.GetEurekaConf()
		routers =eurekaConf.GuulRouter
		eurekaConf.ShowAllEurekaConf()
	})
	eurekaConf.ShowAllEurekaConf()
	ret = &retMessageBody.RetMessage{Status: 20001, Result: &retMessageBody.MessageBody{Messsage: "未知错误"}}
}

func RunRouter(app *iris.Application, filter filter.Filter) {

	app.Get("/refresh", func(context iris.Context) {
		    trigger.Fire("refresh-router-conf")
			trigger.Fire("refresh-router")
			context.JSON(iris.Map{"message":"OK"})
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
			newHeaders["Access-Control-Allow-Origin"] ="*"   //允许跨域访问
			path := context.Path()
			for _, srv := range routers {
				if strings.HasPrefix(path, string([]rune(srv["path"])[:len(srv["path"])-1])) {
					serviceUrl = strings.Join([]string{discovery.GetServiceBaseUrl(srv["srvId"]),
						string([]rune(path)[len(srv["path"])-1:])}, "")
				}
			}

			ret.Result.Data = map[string]string{}
			if !strings.HasPrefix(serviceUrl, "http://") {
				ret.Status = 3000211
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

				res, _ = grequests.Get(serviceUrl, ro)

				switch {

				case res.StatusCode == 500:
					ret.Status = 3000211
					ret.Result.Messsage = SRVINTERNALERROR
					context.JSON(ret)
				case res.StatusCode == 404:
					ret.Status = 3000211
					ret.Result.Messsage = SRVPATHNOTFOUND
					context.JSON(ret)
				case res.StatusCode == 501:
					ret.Status = 3000211
					ret.Result.Messsage = SRVBAD
					context.JSON(ret)
				case res.StatusCode == 403:
					ret.Status = 3000211
					ret.Result.Messsage = SRVNOPOST
					context.JSON(ret)
				case res.Ok:
					m := make(map[string]interface{})
					json.Unmarshal(res.Bytes(), &m)
					context.JSON(m)
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
				res, _ = grequests.Post(serviceUrl, ro)
				ret := retMessageBody.RetMessage{Result: &retMessageBody.MessageBody{}}
				ret.Result.Data = map[string]string{}
				switch {
				case len(serviceUrl) < 1:
					ret.Status = 3000211
					ret.Result.Messsage = SRVNOTFOUND
					context.JSON(ret)
				case res.StatusCode == 500:
					ret.Status = 3000211
					ret.Result.Messsage = SRVINTERNALERROR
					context.JSON(ret)
				case res.StatusCode == 404:
					ret.Status = 3000211
					ret.Result.Messsage = SRVPATHNOTFOUND
					context.JSON(ret)
				case res.StatusCode == 501:
					ret.Status = 3000211
					ret.Result.Messsage = SRVBAD
					context.JSON(ret)
				case res.StatusCode == 403:
					ret.Status = 3000211
					ret.Result.Messsage = SRVNOPOST
					context.JSON(ret)
				case res.Ok:
					m := make(map[string]interface{})
					json.Unmarshal(res.Bytes(), &m)
					context.JSON(m)
				}

			default:
				ret := retMessageBody.RetMessage{Result: &retMessageBody.MessageBody{}}
				ret.Status = 3000211
				ret.Result.Messsage = "不支持的请求方法"
				ret.Result.Data = map[string]string{}
				context.JSON(ret)
			}
		})
	}
}
