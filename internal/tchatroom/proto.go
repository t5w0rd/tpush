package tchatroom

type LoginReq struct {
	Uid int64 `json:"uid"`
}

type LoginRsp struct {
	Id int64 `json:"id"`
}

type EnterChanReq struct {
	Chans []string `json:"chans"`
}

type EnterChanRsp struct {
}

type ExitChanReq struct {
	Chans []string `json:"chans"`
}

type ExitChanRsp struct {
}

type SendToClientReq struct {
	Ids  []int64     `json:"ids"`
	Data interface{} `json:"data,omitempty"`
}

type SendToClientRsp struct {
}

type SendToUserReq struct {
	Uids []int64     `json:"uids"`
	Data interface{} `json:"data,omitempty"`
}

type SendToUserRsp struct {
}

type SendToChanReq struct {
	Chans []string    `json:"chans"`
	Data  interface{} `json:"data,omitempty"`
}

type SendToChanRsp struct {
}

type RecvDataReq struct {
}

type RecvDataRsp struct {
	Id   int64       `json:"id"`
	Uid  int64       `json:"uid"`
	Chan string      `json:"chan"`
	Data interface{} `json:"data,omitempty"`
}
