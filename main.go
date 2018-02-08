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
)

var reg eureka.RegisterEureka
func main()  {
	app := iris.New()
	trigger.On("refresh-router" , func() {
		log.Println("刷新路由")
		router.RunRouter(app,new(filter.AuthHandler))
	})
	router.RunRouter(app,new(filter.AuthHandler))
	regmessage, _ := reg.DoRegisterService()
	log.Println(regmessage)
	go reg.SendHeartBeat() //发送心跳
	serverPort := new(eurekaConf.EurekaConf).GetEurekaConf().GetHostIPPort()
	app.Run(iris.Addr(":"+strconv.Itoa(serverPort)) ,iris.WithoutServerError(iris.ErrServerClosed))
}
