package main

import (
	"github.com/kataras/iris"
	"guul/router"
	"guul/filter"
	"log"
	"guul/eureka/register"
	"github.com/sadlil/go-trigger"
	eurekaConf "guul/eureka/conf"
	"strconv"
	"runtime"
)


func main()  {
	runtime.GOMAXPROCS(runtime.NumCPU()) //  使用所有的机器核心
	app := iris.New()
	trigger.On("refresh-router" , func() {
		log.Println("刷新路由")
		router.RunRouter(app,new(filter.AuthHandler))
	})
	router.RunRouter(app,new(filter.AuthHandler))
	eurekaRegSrv :=new (eureka.RegisterEureka)
	regmessage, _ := eurekaRegSrv.DoRegisterService()
	log.Println(regmessage)
	go eurekaRegSrv.SendHeartBeat() //发送心跳
	serverPort := new(eurekaConf.EurekaConf).GetEurekaConf().GetHostIPPort()
	app.Run(iris.Addr(":"+strconv.Itoa(serverPort)) ,iris.WithoutServerError(iris.ErrServerClosed))
}
