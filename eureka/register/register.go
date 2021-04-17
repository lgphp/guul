package eureka

import (
	"guul/eureka/conf"
	"fmt"
	"time"
	"strconv"
	"strings"
	"log"
	"github.com/sadlil/go-trigger"
	"github.com/levigross/grequests"
	"github.com/kataras/iris/context"

)

type RegisterEureka struct {
}

type Any interface{}

var eurekaConf *eureka.EurekaConf
var homePageUrl, statePageUrl, healthPageUrl string

func init() {
	eurekaConf = eurekaConf.GetEurekaConf()
	homePageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
		strconv.Itoa(eurekaConf.GetHostIPPort())}, "")
	statePageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
		strconv.Itoa(eurekaConf.GetHostIPPort()), "/", "info"}, "")
	healthPageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
		strconv.Itoa(eurekaConf.GetHostIPPort()), "/", "health"}, "")
}

func (this RegisterEureka) SendHeartBeat() {
	/**
	 发送服务心跳
   */
	beatFlag := false
	trigger.On("send-beat", func() {
		log.Println("刷新配置注册心跳")
		beatFlag = true

	})
	beatUrl := strings.Join([]string{eurekaConf.GetEurekaUrl(), "apps", eurekaConf.GetServiceName(),
		eurekaConf.GetInstanceID()}, "/")
	for {
		resp, errs := grequests.Put(beatUrl, &grequests.RequestOptions{RequestTimeout: eureka.REQUESTTIMEOUT})
		if errs != nil {
			log.Println("发送服务心跳失败:", errs)
			msg, _ := this.DoRegisterService() //重试注册服务
			log.Println(msg)
			//break
			//os.Exit(0) //退出系统
		} else {
			if !resp.Ok {
				log.Println("发送服务心跳失败:", resp.StatusCode)
				msg, _ := this.DoRegisterService() //重试注册服务
				log.Println(msg)
			}
		}
		if beatFlag {
			break
		}
		time.Sleep(5 * time.Second)
	}

}

func (this RegisterEureka) DoRegisterService() (message string, err error) {

	/**
		注册服务
	 */
	log.Println(eurekaConf)
	instanceData := this.getInstanceData()
	registerUrl := strings.Join([]string{eurekaConf.GetEurekaUrl(), "apps", eurekaConf.GetServiceName()}, "/")
	fmt.Println(instanceData)
	resp, errs := grequests.Post(registerUrl,
		&grequests.RequestOptions{JSON: instanceData,
			Headers: map[string]string{"Content-Type": context.ContentJSONHeaderValue},
			RequestTimeout: eureka.REQUESTTIMEOUT})
	message = "服务注册成功"
	if errs != nil {

		err = fmt.Errorf("%s", errs)
		message = "服务注册失败" + fmt.Sprint(err)
	} else {

		if !resp.Ok {
			fmt.Println(resp)
			err = fmt.Errorf("%s", resp.StatusCode)
			message = "服务注册失败" + fmt.Sprint(err)
		}

	}

	/**
		热更新eureka 配置
	 */
	trigger.On("reg-service", func() {
		eurekaConf = eurekaConf.GetEurekaConf()
		homePageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
			strconv.Itoa(eurekaConf.GetHostIPPort())}, "")
		statePageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
			strconv.Itoa(eurekaConf.GetHostIPPort()), "/", "info"}, "")
		healthPageUrl = strings.Join([]string{"http://", eurekaConf.GetHostIPAddr(), ":",
			strconv.Itoa(eurekaConf.GetHostIPPort()), "/", "health"}, "")
		msg, _ := this.DoRegisterService()
		log.Println("刷新配置" + msg)
		go this.SendHeartBeat()
	})

	return
}

func (this RegisterEureka) getInstanceData() (instanceData string) {
	dataCenterInfo := `
		{
		"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
			"name": "MyOwn"
		}
		`

	instanceData = `{
	   "instance":
                {
                    "instanceId": "` + eurekaConf.GetInstanceID() + `",
                    "hostName": "` + eurekaConf.GetHostIPAddr() + `",
                    "app": "` + eurekaConf.GetServiceName() + `",
                    "ipAddr": "` + eurekaConf.GetHostIPAddr() + `",
                    "status": "UP",
                    "overriddenstatus": "UNKNOW",
                    "port": {
                        "$": ` + strconv.Itoa(eurekaConf.GetHostIPPort()) + `,
                        "@enabled": "true"
                    },
                    "securePort": {
                        "$": 443,
                        "@enabled": "false"
                    },
                    "countryId": 1,
                    "dataCenterInfo": ` + dataCenterInfo + `,
                    "leaseInfo": {
                        "renewalIntervalInSecs": ` + strconv.Itoa(eureka.EurekaRenewalIntervalInSecs) + `,
                        "durationInSecs": ` + strconv.Itoa(eureka.EurekaDurationInSecs) + `,
                        "registrationTimestamp": ` + fmt.Sprintf("%d", time.Now().UnixNano()/1000000) + `,
                        "lastRenewalTimestamp": ` + fmt.Sprintf("%d", time.Now().UnixNano()/1000000) + `,
                        "evictionTimestamp": 0,
                        "serviceUpTimestamp": ` + fmt.Sprintf("%d", time.Now().UnixNano()/1000000) + `
                    },
                    "metadata": {
                        "@class": "java.util.Collections$EmptyMap"
                    },
                    "homePageUrl": "` + homePageUrl + `",
                    "statusPageUrl": "` + statePageUrl + `",
                    "healthCheckUrl": "` + healthPageUrl + `",
                    "vipAddress": "` + strings.ToLower(eurekaConf.GetServiceName()) + `",
                    "secureVipAddress": "` + strings.ToLower(eurekaConf.GetServiceName()) + `",
                    "isCoordinatingDiscoveryServer": "false",
                    "lastUpdatedTimestamp": ` + fmt.Sprintf("%d", time.Now().UnixNano()/1000000) + `,
                    "lastDirtyTimestamp": ` + fmt.Sprintf("%d", time.Now().UnixNano()/1000000) + `,
                    "actionType": "ADDED"
                }
		}
	`

	return
}
