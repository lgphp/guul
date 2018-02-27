package errorcode

const (
	EUREKACONFIGNOTFOUND = 7000001
	SERVICENOTFOUND  = 8000001
	SERVICENOTFETCHBASEURL = 8000003
	SERVICEFETCHFAILURE = 8000004
	SERVICEANYERROR = 8000009
	UNKNOWERROR = 8000010
)

var errorCodeToMsg map[int64]string
type EurekaErrorCode struct {
}

func init()  {
	errorCodeToMsg = map[int64]string{
		EUREKACONFIGNOTFOUND: "EUREKA配置文件没找到",
		SERVICENOTFOUND:"服务没找到，可能没有注册",
		SERVICENOTFETCHBASEURL:"服务baseUrl获取不到,服务可能没有注册",
		SERVICEFETCHFAILURE : "服务请求失败",
		SERVICEANYERROR:"服务发生其他未知错误",
		UNKNOWERROR:"GUUL:网关发生未知错误",
	}
}
func (this EurekaErrorCode) ErrMessage( errCode int64 )string  {

	  return errorCodeToMsg[errCode]

}
