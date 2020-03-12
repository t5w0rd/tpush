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
	Ids  []int64 `json:"uids"`
	Data string  `json:"data"`
}

type SendToClientRsp struct {
}

type SendToUserReq struct {
	Uids []int64 `json:"uids"`
	Data string  `json:"data"`
}

type SendToUserRsp struct {
}

type SendToChanReq struct {
	Chans []string `json:"chans"`
	Data  string   `json:"data"`
}

type SendToChanRsp struct {
}

type RecvDataReq struct {
}

type RecvDataRsp struct {
	Id   int64  `json:"id"`
	Uid  int64  `json:"uid"`
	Chan string `json:"chan"`
	Data string `json:"data"`
}
