package tchatroom

import (
	"errors"
	log "github.com/micro/go-micro/v2/logger"
	"time"
	"tpush/internal/websocket"
)

type handler struct {
	room *Room
}

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
	Id  int64  `json:"uid"`
	Msg string `json:"msg"`
}

type SendMsgToClientRsp struct {
}

type SendMsgToUserReq struct {
	Uid int64  `json:"uid"`
	Msg string `json:"msg"`
}

type SendMsgToUserRsp struct {
}

type SendMsgToChanReq struct {
	Chan string `json:"chan"`
	Msg  string `json:"msg"`
}

type SendMsgToChanRsp struct {
}

type RecvMsgReq struct {
}

type RecvMsgRsp struct {
	Id   int64  `json:"id"`
	Uid  int64  `json:"uid"`
	Chan string `json:"chan"`
	Msg  string `json:"msg"`
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

func (h *handler) OnOpen(cli websocket.Client) error {
	loginDone := make(chan int64)
	cli.AddContextValue(loginDoneKey{}, loginDone)
	cli.AddContextValue(clientDataKey{}, &clientData{
		id: h.room.AddClient(cli),
	})

	go func() {
		defer log.Debug("waitLogin complete")
		select {
		case uid := <-loginDone:
			h.room.Login(cli, uid)
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

func (h *handler) OnClose(cli websocket.Client) {
	h.room.RemoveClient(cli)
}

func (h *handler) Login(req websocket.Request, rsp websocket.Response) error {
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

func (h *handler) EnterChan(req websocket.Request, rsp websocket.Response) error {
	var enterChanReq EnterChanReq
	if err := req.DecodeData(&enterChanReq); err != nil {
		return err
	}

	h.room.ClientEnterChannel(req.Client(), enterChanReq.Chans...)

	rsp.EncodeData(&EnterChanRsp{}, 0, "")

	return nil
}

func (h *handler) ExitChan(req websocket.Request, rsp websocket.Response) error {
	var exitChanReq ExitChanReq
	if err := req.DecodeData(&exitChanReq); err != nil {
		return err
	}

	h.room.ClientExitChannel(req.Client(), exitChanReq.Chans...)

	rsp.EncodeData(&ExitChanRsp{}, 0, "")

	return nil
}

func (h *handler) SendMsgToClient(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToClientReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	uid, ok := h.room.User(req.Client())
	if !ok {
		return websocket.Error(rsp, ErrNotLogin, "client hasnot logged in", true)
	}

	id, ok := h.room.ClientId(req.Client())
	if !ok {
		return websocket.Fatal(rsp, errors.New("client has no id"))
	}

	cli, ok := h.room.Client(request.Id)
	if !ok {
		return websocket.Error(rsp, ErrClientNotFound, "dest client not found", false)
	}

	data := &RecvMsgRsp{
		Id:   id,
		Uid:  uid,
		Chan: "",
		Msg:  request.Msg,
	}
	go cli.Write(CmdRecvMsg, 0, data, 0, "", false)

	rsp.EncodeData(&SendMsgToClientRsp{}, 0, "")
	return nil
}

func (h *handler) SendMsgToUser(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToUserReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	uid, ok := h.room.User(req.Client())
	if !ok {
		return websocket.Error(rsp, ErrNotLogin, "client hasnot logged in", true)
	}

	id, ok := h.room.ClientId(req.Client())
	if !ok {
		return websocket.Fatal(rsp, errors.New("client has no id"))
	}

	cligrp, ok := h.room.ClientsOfUser(request.Uid)
	if !ok {
		return websocket.Error(rsp, ErrUserNotFound, "dest user not found", false)
	}

	data := &RecvMsgRsp{
		Id:   id,
		Uid:  uid,
		Chan: "",
		Msg:  request.Msg,
	}
	go cligrp.Write(CmdRecvMsg, 0, data, 0, "", false)

	rsp.EncodeData(&SendMsgToUserRsp{}, 0, "")
	return nil
}

func (h *handler) SendMsgToChan(req websocket.Request, rsp websocket.Response) error {
	var request SendMsgToChanReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	uid, ok := h.room.User(req.Client())
	if !ok {
		return websocket.Error(rsp, ErrNotLogin, "client hasnot logged in", true)
	}

	id, ok := h.room.ClientId(req.Client())
	if !ok {
		return websocket.Fatal(rsp, errors.New("client has no id"))
	}

	cligrp, ok := h.room.ClientsInChannel(request.Chan)
	if !ok {
		return websocket.Error(rsp, ErrChanNotFound, "dest chan not found", false)
	}

	data := &RecvMsgRsp{
		Id:   id,
		Uid:  uid,
		Chan: request.Chan,
		Msg:  request.Msg,
	}
	go cligrp.Write(CmdRecvMsg, 0, data, 0, "", false)

	rsp.EncodeData(&SendMsgToChanRsp{}, 0, "")
	return nil
}

func (h *handler) RecvMsg(req websocket.Request, rsp websocket.Response) error {
	return websocket.Error(rsp, ErrWrongCmd, "wrong cmd", false)
}
