package discovery

import (
	"guul/eureka/conf"
	"strings"
	"github.com/parnurzeal/gorequest"
	"time"
	"guul/eureka/retMessageBody"
	"fmt"
	"encoding/json"
	rand2 "crypto/rand"
	"math/big"
	"github.com/levigross/grequests"
	"strconv"
	"guul/eureka/errorcode"
	"log"
)

type Any interface{}

var (
	eurekaConf    *eureka.EurekaConf
	ret           *retMessageBody.RetMessage
	eurekaErrCode errorcode.EurekaErrorCode
)

func init() {
	/*
		获取要发现的服务
	 */

	eurekaConf = eurekaConf.GetEurekaConf()
	ret = &retMessageBody.RetMessage{Status: 1, Result: &retMessageBody.MessageBody{Messsage: "未知错误"}}
}

func getServiceUrl(serviceName string) {
	instanceUrl := strings.Join([]string{eurekaConf.GetEurekaUrl(), "apps", strings.ToUpper(serviceName)}, "/")
	req := gorequest.New().Timeout(5 * time.Second) //5秒超时
	resp, body, errs := req.Get(instanceUrl).Set("Accept", "application/json").End()
	ret.MU.Lock()
	if errs != nil {
		ret.Status = errorcode.SERVICENOTFOUND
		ret.Result.Messsage = strings.Join([]string{serviceName, eurekaErrCode.ErrMessage(errorcode.SERVICENOTFOUND), fmt.Sprint(errs)}, "")
	} else {
		if resp.StatusCode != 200 {
			ret.Status = errorcode.SERVICENOTFOUND
			ret.Result.Messsage = strings.Join([]string{serviceName,
				eurekaErrCode.ErrMessage(errorcode.SERVICENOTFOUND), ":返回状态码为:" + strconv.Itoa(resp.StatusCode), fmt.Sprint(errs)}, "")
		} else {
			ret.Status = 0
			ret.Result.Messsage = map[string]string{}
			ret.Result.Data = []byte(body)
		}
	}
	defer ret.MU.Unlock()
}

func GetServiceBaseUrl(serviceName string) string {
	getServiceUrl(serviceName)
	ret.MU.Lock()
	defer ret.MU.Unlock()
	var InstanceData = make(map[string]interface{})
	if ret.Status == 0 {
		json.Unmarshal(ret.Result.Data.([]byte), &InstanceData)
		instances := InstanceData["application"].(map[string]interface{})["instance"].([]interface{})
		serviceBaseUrls := make([]string, len(instances))
		for i, v := range instances {
			k := v.(map[string]interface{})
			serviceBaseUrls[i] = "http://" + k["hostName"].(string) + ":" + fmt.Sprint(k["port"].(map[string]interface{})["$"]) + "/"
		}
		iUrl, _ := rand2.Int(rand2.Reader, big.NewInt(int64(len(serviceBaseUrls))))
		serviceBaseUrl := serviceBaseUrls[iUrl.Int64()]
		return serviceBaseUrl
	} else {
		log.Println(eurekaErrCode.ErrMessage(errorcode.SERVICENOTFETCHBASEURL))
		return "" //没有获取到BaseUrl
	}

}

func DoService(verb, serviceName, routerPath string, formData map[string]string, jsonData interface{}, headers map[string]string) *retMessageBody.RetMessage {
	var method = "GET"
	if verb != "" {
		method = strings.ToUpper(verb)
	}
	getServiceUrl(serviceName)
	ret.MU.Lock()
	defer ret.MU.Unlock()
	var InstanceData = make(map[string]interface{})
	if ret.Status == 0 {
		json.Unmarshal(ret.Result.Data.([]byte), &InstanceData)
		instances := InstanceData["application"].(map[string]interface{})["instance"].([]interface{})
		serviceBaseUrls := make([]string, len(instances))
		for i, v := range instances {
			k := v.(map[string]interface{})
			serviceBaseUrls[i] = "http://" + k["hostName"].(string) + ":" + fmt.Sprint(k["port"].(map[string]interface{})["$"]) + "/"
		}
		iUrl, _ := rand2.Int(rand2.Reader, big.NewInt(int64(len(serviceBaseUrls))))
		doServiceUrl := serviceBaseUrls[iUrl.Int64()]
		resp, errs := grequests.Req(method, doServiceUrl+routerPath,
			&grequests.RequestOptions{Data: formData, JSON: jsonData, RequestTimeout: 5 * time.Second, Headers: headers})
		if errs != nil {
			ret.Status = errorcode.SERVICEFETCHFAILURE
			ret.Result.Messsage = strings.Join([]string{doServiceUrl + routerPath,
				eurekaErrCode.ErrMessage(errorcode.SERVICEFETCHFAILURE), fmt.Sprint(errs)}, "")
			ret.Result.Data = map[string]string{}
		} else {
			if resp.StatusCode != 200 {
				ret.Status = errorcode.SERVICEFETCHFAILURE
				ret.Result.Messsage = strings.Join([]string{doServiceUrl + routerPath,
					eurekaErrCode.ErrMessage(errorcode.SERVICEFETCHFAILURE), "返回状态码：",
					strconv.Itoa(resp.StatusCode)}, "")
				ret.Result.Data = make(map[string]interface{})
			} else {
				ret.Status = 0
				ret.Result.Messsage = map[string]string{}
				ret.Result.Data = resp.Bytes()
			}
		}
	} else {
		//没有获取到BaseUrl
		ret.Status = errorcode.SERVICENOTFETCHBASEURL
		ret.Result.Messsage = strings.Join([]string{routerPath, "请求失败,没有发现服务",
			eurekaErrCode.ErrMessage(errorcode.SERVICENOTFETCHBASEURL), serviceName}, "")
		ret.Result.Data = map[string]string{}
	}
	return ret

}
