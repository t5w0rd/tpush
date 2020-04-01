package proto

type SendToUserReq struct {
	Uids []int64     `json:"uids"`
	Data interface{} `json:"data,omitempty"`
	Id   int64       `json:"id,omitempty"`
	Uid  int64       `json:"uid,omitempty"`
}

type SendMsgRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
