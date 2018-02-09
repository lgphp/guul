package filter

import (
	"github.com/kataras/iris"
	"guul/eureka/discovery"
	"guul/eureka/conf"
	"strings"
	"log"
)
var HasPath = func(path string, prefix []string) (bool) {

	for _, v := range prefix {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	return false
}

type Filter interface {
	PreHandler() func(ctx iris.Context)
}


var (
	eurekaConf *eureka.EurekaConf
)
type AuthHandler struct{}
func (_ *AuthHandler) PreHandler() func(ctx iris.Context) {
	eurekaConf = eurekaConf.GetEurekaConf()
	return func(ctx iris.Context) {
		path := ctx.Path()
		if HasPath(path, eurekaConf.NeedAuthPathPrefix) {
			Header := ctx.Request().Header
			authHeader := make(map[string]string)
			for i, v := range Header {
				authHeader[i] = v[0]
			}
			//ctx.Header("Access-Control-Allow-Origin","*")
			//authHeader["token"] =ctx.GetHeader("authorization")
			tokenParam :=  map[string]string{"token": ctx.GetHeader("authorization")}
			ret := discovery.DoService("POST", "CX-SERVICE-USER",
				"appCommonsUserLogin/token", tokenParam ,nil, authHeader)
			if ret.Status != 0 {
				log.Println(ret.Status, ret.Result.Messsage)
				ctx.StatusCode(403) //鉴权失败
				return
			}
		}
		ctx.Next()
	}
}
