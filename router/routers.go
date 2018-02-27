package router

import (
	"github.com/kataras/iris"
	"github.com/levigross/grequests"
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
	"guul/eureka/errorcode"
	"guul/util"
	"io/ioutil"
)

const (
	SRVNOTFOUND      = "GUUL: 调用的服务不可用,该服务没有注册"
	SRVINTERNALERROR = "GUUL: 调用的服务内部服务器错误"
	SRVPATHNOTFOUND  = "GUUL: 调用的路由没有找到"
	SRVForbidden     = "GUUL: 调用的服务禁止访问"
	SRVBAD           = "GUUL: 调用的服务503BAD GETEWAY"
	NOTSUPPORTMETHOD = "GUUL: 不支持的请求动作，目前只支持[GET,POST]"
	REQUESTTIMEOUT = 20 * time.Second
)

var (
	eurekaConf *eureka.EurekaConf
	routers    []map[string]string
	ret        *retMessageBody.RetMessage
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
	ret = &retMessageBody.RetMessage{Status: errorcode.UNKNOWERROR,
	Result: &retMessageBody.MessageBody{Messsage: errorcode.EurekaErrorCode{}.ErrMessage(errorcode.UNKNOWERROR)}}
}

func RunRouter(app *iris.Application, filter filter.Filter) {

	app.Use(func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*") //网关允许跨域访问
		ctx.Next()
	})

	manageUser := os.Getenv("MANAGE-USER")
	if manageUser == "" {
		manageUser = "lgphp"
	}
	managePass := os.Getenv("MANAGE-PASS")
	if managePass == "" {
		managePass = "52cx.comqazxc"
	}
	authConfig := basicauth.Config{
		Users:   map[string]string{manageUser: managePass},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(30) * time.Minute,
	}
	authentication := basicauth.New(authConfig)

	app.Get("/refresh", authentication, func(context iris.Context) {

		trigger.Fire("refresh-router-conf")
		trigger.Fire("reg-service")
		trigger.Fire("send-beat")
		app.RefreshRouter() //刷新路由
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

		context.JSON(iris.Map{"ServiceId": eurekaConf.GetServiceName(),
			"ServiceName": "GUUL-APP网关",
			"Author": "luogang@52cx.com",
			"Date": "2018-02-10",
			"Language": "Golang1.9.2(Base of iris)"})
	})

	app.Get("/health", func(context iris.Context) {
		context.JSON(iris.Map{"status": "UP"})
	})

	for _, v := range routers {
		app.Any(v["path"], filter.PreHandler(), func(ctx iris.Context) {
			method := ctx.Method()
			newHeaders := util.GetHeaders(ctx)
			path := ctx.Path()
			serviceUrl := "" //服务的URL
			serviceId := ""  //转发服务的id
			for _, k := range routers {
				if strings.HasPrefix(path, string([]rune(k["path"])[:len(k["path"])-1])) {
					serviceUrl = strings.Join([]string{discovery.GetServiceBaseUrl(k["srvId"]), string([]rune(path)[len(k["path"])-1:])}, "")
					serviceId = k["srvId"]
				}
			}

			ret.Result.Data = map[string]string{}
			if serviceUrl == "" {
				ret.Status = iris.StatusNotFound
				ret.Result.Messsage = "服务:" + serviceId + "   " + SRVNOTFOUND
				ctx.JSON(ret)
				return
			}

			//转发请求

			params := ctx.URLParams()
			ro := &grequests.RequestOptions{
				Params:         params,
				Headers:        newHeaders,
				RequestTimeout: REQUESTTIMEOUT,
			}
			switch strings.ToUpper(method) {
			case "GET":
			case "POST":
				postdata := util.GetFormValues(ctx)

				ro = &grequests.RequestOptions{
					Params:         params,
					Headers:        newHeaders,
					Data:           postdata,
					//RequestBody:    ctx.Request().Body,
					RequestTimeout: REQUESTTIMEOUT,
				}
				rawJson, _ := ioutil.ReadAll(ctx.Request().Body)
				if len(string(rawJson)) > 0 {
					ro.JSON = string(rawJson)
				}
			default:
				ret := retMessageBody.RetMessage{Result: &retMessageBody.MessageBody{}}
				ret.Status = iris.StatusMethodNotAllowed
				ret.Result.Messsage = NOTSUPPORTMETHOD
				ret.Result.Data = map[string]string{}
				ctx.JSON(ret)
				return
			}

			res, err := grequests.Req(method, serviceUrl, ro)
			switch {
			case err != nil:
				ret.Status = iris.StatusRequestTimeout
				ret.Result.Messsage = "请求超时或者其他错误" + err.Error()
				ctx.JSON(ret)
			case res.StatusCode == iris.StatusInternalServerError:
				ret.Status = iris.StatusInternalServerError
				ret.Result.Messsage = SRVINTERNALERROR
				ctx.JSON(ret)
			case res.StatusCode == iris.StatusNotFound:
				ret.Status = iris.StatusNotFound
				ret.Result.Messsage = SRVPATHNOTFOUND
				ctx.JSON(ret)
			case res.StatusCode == iris.StatusBadGateway:
				ret.Status = iris.StatusBadGateway
				ret.Result.Messsage = SRVBAD
				ctx.JSON(ret)
			case res.StatusCode == iris.StatusForbidden:
				ret.Status = iris.StatusForbidden
				ret.Result.Messsage = SRVForbidden
				ctx.JSON(ret)
			case res.Ok:
				ctx.Header("Content-Type", res.Header.Get("Content-Type"))
				ctx.Write(res.Bytes())
			default:
				ret.Status = errorcode.SERVICEANYERROR
				ret.Result.Messsage = errorcode.EurekaErrorCode{}.ErrMessage(errorcode.SERVICEANYERROR)
				ctx.JSON(ret)
			}
		})
	}

}
