package util

import (
	"github.com/kataras/iris"
)

func GetHeaders(ctx iris.Context) (newHeaders map[string]string) {
	newHeaders = make(map[string]string)
	for i, v := range ctx.Request().Header {
		newHeaders[i] = v[0]
	}
	return
}

func GetFormValues(ctx iris.Context) (newFormValues map[string]string) {
	newFormValues = make(map[string]string)
	for u, c := range ctx.FormValues() {
		newFormValues[u] = c[0]
	}
	return
}
