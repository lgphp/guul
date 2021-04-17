package filter

import (
	"github.com/kataras/iris"
	"guul/eureka/discovery"
	"guul/eureka/conf"
	"strings"
	"guul/util"
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
			authHeader := util.GetHeaders(ctx)
			tokenParam := map[string]string{"token": ctx.GetHeader("authorization")}
			ret := discovery.DoService("POST", "CX-SERVICE-USER",
				"appCommonsUserLogin/token", tokenParam, nil, authHeader)
			if ret.Status != 0 {
				ret.Status = iris.StatusForbidden
				ret.Result.Messsage = strings.Join([]string{"GUUL:PreHandler 过滤器-> 令牌无效，鉴权失败   |    API==>", path}, "")
				ret.Result.Data = map[string]string{}
				ctx.StatusCode(iris.StatusForbidden)
				ctx.JSON(ret)//鉴权失败
				return
			}
		}
		ctx.Next()
	}
}
