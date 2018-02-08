package retMessageBody

import "sync"

type MessageBody struct {
	Messsage interface{} `json:"message"`
	Data interface{} `json:"data"`
}
type RetMessage struct {
	MU sync.Mutex `json:"-"`
	Status int64 `json:"status"`
	Result *MessageBody `json:"result"`
}