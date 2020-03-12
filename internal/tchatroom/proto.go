package tchatroom

import "tpush/internal/twebsocket"

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
	Ids []int64 `json:"uids"`
	twebsocket.Payload
}

type SendToClientRsp struct {
}

type SendToUserReq struct {
	Uids []int64 `json:"uids"`
	twebsocket.Payload
}

type SendToUserRsp struct {
}

type SendToChanReq struct {
	Chans []string `json:"chans"`
	twebsocket.Payload
}

type SendToChanRsp struct {
}

type RecvDataReq struct {
}

type RecvDataRsp struct {
	Id   int64  `json:"id"`
	Uid  int64  `json:"uid"`
	Chan string `json:"chan"`
	twebsocket.Payload
}
