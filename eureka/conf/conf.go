package eureka

import (
	"strconv"
	"strings"
	"log"
	"os"
	"net"
	"github.com/levigross/grequests"
)

const CONFIGURL = "http://172.17.10.95:1505/statics/guul-service/application.json"

type EurekaConf struct {
	HostIPAddr         string              `"主机地址"`
	HostIPPort         int
	ServiceName        string
	InstanceId         string
	EurekaUrl          string              `"eureka的URL"`
	GuulRouter         []map[string]string `"路由配置"`
	NeedAuthPathPrefix []string
}

func (this *EurekaConf) GetEurekaConf() *EurekaConf {
	resp,err :=grequests.Get(CONFIGURL,nil)
	if err != nil {
		log.Println("配置文件没有找到..系统退出", err)
		os.Exit(0)
	}


	if resp.StatusCode != 200 {
		log.Println("配置文件无法访问..系统退出", err)
		os.Exit(1)
	}
	eurekaConfPtr := &EurekaConf{}
	errs := resp.JSON(&eurekaConfPtr)
	//errs := json.Unmarshal(resp.Bytes(), &eurekaConfPtr)
	if errs != nil {
		log.Println("配置文件格式非法..系统退出", err)
		os.Exit(1)
	}

	if envIPAddr:=os.Getenv("SERVER-ADDR"); envIPAddr!=""{
		eurekaConfPtr.HostIPAddr = envIPAddr
		if eurekaConfPtr.HostIPAddr=="" {
			eurekaConfPtr.HostIPAddr = getHostIP()
		}
	}

	if envHostPort:=os.Getenv("SERVER-PORT");envHostPort!=""{
		eurekaConfPtr.HostIPPort , err = strconv.Atoi(envHostPort)
		if err!=nil{
			eurekaConfPtr.HostIPPort = 8000  //默认端口
			log.Printf("主机端口:%s不是一个数字",envHostPort)
		}
	}

	if envEurekaUrl:=os.Getenv("EUREKA-URL");envEurekaUrl!=""{
			eurekaConfPtr.EurekaUrl = envEurekaUrl
	}



	eurekaConfPtr.InstanceId = strings.Join([]string{getHostIP(),
		":", strconv.Itoa(eurekaConfPtr.HostIPPort), ":",
		eurekaConfPtr.ServiceName}, "")



	return eurekaConfPtr
}

func getHostIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		log.Println("无法获取本机IP地址，系统退出")
		os.Exit(1)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return os.Getenv("HOSTNAME")  //取不到本机IP 取主机名
}

func (this *EurekaConf) String() string {

	return "HOST:" + this.GetHostIPAddr() + "   PORT:" +
		strconv.Itoa(this.HostIPPort) + "   EUREKAURL:" + this.EurekaUrl + "   InstanceID" + this.GetInstanceID()
}

func (this *EurekaConf) ShowAllEurekaConf()  {
	for _,v :=range this.GuulRouter {
		log.Println("网关路由配置" , v["path"] , "=>", v["srvId"])
	}

}

func (this *EurekaConf) SetHostIPAddr(ipAddr string) {
	this.HostIPAddr = ipAddr
}

func (this EurekaConf) GetHostIPAddr() string {
	return this.HostIPAddr
}

func (this *EurekaConf) SetHostIPPOrt(port int) {

	this.HostIPPort = port
}

func (this EurekaConf) GetHostIPPort() int {
	return this.HostIPPort
}

func (this *EurekaConf) SetServiceName(srvName string) {
	this.ServiceName = srvName
}
func (this EurekaConf) GetServiceName() string {
	return this.ServiceName
}
func (this *EurekaConf) SetEurekaUrl(eurekaUrl string) {
	this.EurekaUrl = eurekaUrl
}
func (this EurekaConf) GetEurekaUrl() string {
	return this.EurekaUrl
}

func (this EurekaConf) GetInstanceID() string {
	return this.InstanceId
}

func (this EurekaConf) GetGuulRouter() []map[string]string {
	return this.GuulRouter
}

func (this EurekaConf) GetNeedAuthPathPrefix()[]string  {
	return this.NeedAuthPathPrefix
}