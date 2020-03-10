package handler

import (
	"context"
	"time"
	"tpush/internal/websocket"
	log "github.com/micro/go-micro/v2/logger"
	push "tpush/srv/push/proto/push"
)

type Push struct {
	hub* Hub
}

func NewPush() *Push {
	h := &Push{
		hub: NewHub(),
	}
	return h
}

// Call is a single request handler called via client.Call or the generated client code
func (h *Push) Call(ctx context.Context, req *push.Request, rsp *push.Response) error {
	log.Info("received Push.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (h *Push) Stream(ctx context.Context, req *push.StreamingRequest, stream push.Push_StreamStream) error {
	log.Infof("received Push.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("responding: %d", i)
		if err := stream.Send(&push.StreamingResponse{
			Count: int64(i + 1000),
		}); err != nil {
			return err
		}
	}
	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (h *Push) PingPong(ctx context.Context, stream push.Push_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("got ping %v", req.Stroke)
		if err := stream.Send(&push.Pong{Stroke: req.Stroke + 1000}); err != nil {
			return err
		}
	}
}

const (
	CmdLogin = "login"
	CmdEnter = "enter"
	CmdExit = "exit"
	CmdSendMsgToClient = "snd2cli"
	CmdSendMsgToUser = "snd2usr"
	CmdSendMsgToChan = "snd2chan"
	CmdRecvMsg = "rcvmsg"
)

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

type SendMsgToClientReq struct {
	Id int64 `json:"uid"`
	Msg string `json:"msg"`
}

type SendMsgToClientRsp struct {
}

type SendMsgToUserReq struct {
	Uid int `json:"uid"`
	Msg string `json:"msg"`
}

type SendMsgToUserRsp struct {
}

type SendMsgToChanReq struct {
	Chan string `json:"chan"`
	Msg string `json:"msg"`
}

type SendMsgToChanRsp struct {
}

type RecvMsgReq struct {
}

type RecvMsgRsp struct {
	Id  int64 `json:"id"`
	Uid int64 `json:"uid"`
	Chan string `json:"chan"`
	Msg string `json:"msg"`
}

type HelloReq struct {
	Name string `json:"name"`
}

type HelloRsp struct {
	Say string `json:"say"`
}

type loginDoneKey struct{}

type clientDataKey struct{}

type clientData struct {
	id int64
}

func (h *Push) OnOpen(cli websocket.Client) error {
	loginDone := make(chan int64)
	cli.AddContextValue(loginDoneKey{}, loginDone)
	cli.AddContextValue(clientDataKey{}, &clientData{
		id: h.hub.AddClient(cli),
	})

	go func() {
		defer log.Debug("waitLogin complete")
		select {
		case uid := <-loginDone:
			h.hub.Register(cli, uid)
			log.Debugf("client logged in succ, uid: %v", uid)
			return
		case <-time.After(time.Second * 2):
			log.Error("client hasnot logged in for a long time")
			cli.Close()
			return
		}
	}()
	return nil
}

func (h *Push) OnClose(cli websocket.Client) {
	h.hub.RemoveClient(cli)
}

func (h *Push) Login(req websocket.Request, rsp websocket.Response) error {
	var loginReq LoginReq
	if err := req.DecodeData(&loginReq); err != nil {
		return err
	}
	uid := loginReq.Uid

	loginDone := req.Client().ContextValue(loginDoneKey{}).(chan int64)
	loginDone <- uid

	clientData := req.Client().ContextValue(clientDataKey{}).(*clientData)

	rsp.EncodeData(&LoginRsp{
		Id: clientData.id,
	}, 0, "")

	return nil
}

func (h *Push) EnterChan(req websocket.Request, rsp websocket.Response) error {
	var enterChanReq EnterChanReq
	if err := req.DecodeData(&enterChanReq); err != nil {
		return err
	}

	h.hub.ClientEnterChannel(req.Client(), enterChanReq.Chans...)

	rsp.EncodeData(&EnterChanRsp{}, 0, "")

	return nil
}

func (h *Push) ExitChan(req websocket.Request, rsp websocket.Response) error {
	var exitChanReq ExitChanReq
	if err := req.DecodeData(&exitChanReq); err != nil {
		return err
	}

	h.hub.ClientExitChannel(req.Client(), exitChanReq.Chans...)

	rsp.EncodeData(&ExitChanRsp{}, 0, "")

	return nil
}

func (h *Push) SendMsgToClient(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToClientReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	cli, ok := h.hub.Client(request.Id)
	if !ok {
		return websocket.Error(rsp, -1, "dest client not found", false)
	}
	id, ok := h.hub.ClientId(req.Client())
	if !ok {
		return websocket.Error(rsp, -1, "client has no id", true)
	}
	uid, ok := h.hub.User(req.Client())

	data := &RecvMsgRsp{
		Id: id,
		Uid: uid,
		Chan: "",
		Msg: request.Msg,
	}
	cli.(websocket.Client).Write(CmdRecvMsg, 0, data, 0, "", false)

	rsp.EncodeData(&SendMsgToClientRsp{}, 0, "")
	return nil
}

func (h *Push) SendMsgToUser(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToUserReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	rsp.EncodeData(&SendMsgToUserRsp{}, 0, "")
	return nil
}

func (h *Push) SendMsgToChan(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToChanReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	rsp.EncodeData(&SendMsgToChanRsp{}, 0, "")
	return nil
}

func (h *Push) RecvMsg(req websocket.Request, rsp websocket.Response) error {
	return websocket.Error(rsp, -1, "wrong cmd", false)
}

func (h *Push) Hello(req websocket.Request, rsp websocket.Response) error {
	var helloReq HelloReq
	if err := req.DecodeData(&helloReq); err != nil {
		return err
	}

	rsp.EncodeData(&HelloRsp{"Hello, " + helloReq.Name}, 0, "")

	return nil
}
