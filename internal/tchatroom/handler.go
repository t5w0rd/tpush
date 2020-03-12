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
	var request LoginReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}
	uid := request.Uid

	loginDone := req.Client().ContextValue(loginDoneKey{}).(chan int64)
	loginDone <- uid

	clientData := req.Client().ContextValue(clientDataKey{}).(*clientData)

	rsp.EncodeData(&LoginRsp{
		Id: clientData.id,
	}, 0, "")

	return nil
}

func (h *handler) EnterChan(req websocket.Request, rsp websocket.Response) error {
	var request EnterChanReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	h.room.ClientEnterChannel(req.Client(), request.Chans...)

	rsp.EncodeData(&EnterChanRsp{}, 0, "")

	return nil
}

func (h *handler) ExitChan(req websocket.Request, rsp websocket.Response) error {
	var request ExitChanReq
	if err := req.DecodeData(&request); err != nil {
		return err
	}

	h.room.ClientExitChannel(req.Client(), request.Chans...)

	rsp.EncodeData(&ExitChanRsp{}, 0, "")

	return nil
}

func (h *handler) SendToClient(req websocket.Request, rsp websocket.Response) error {
	var request SendToClientReq
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

	data := &RecvDataRsp{
		Id:   id,
		Uid:  uid,
		Chan: "",
		Data: request.Data,
	}

	if len(request.Ids) == 1 {
		cli, ok := h.room.Client(request.Ids[0])
		if !ok {
			return websocket.Error(rsp, ErrClientNotFound, "dest client not found", false)
		}
		go cli.Write(CmdRecvData, 0, data, 0, "", false)
	} else {
		cligrp := h.room.Clients(request.Ids)
		go cligrp.Write(CmdRecvData, 0, data, 0, "", false)
	}

	rsp.EncodeData(&SendToClientRsp{}, 0, "")
	return nil
}

func (h *handler) SendToUser(req websocket.Request, rsp websocket.Response) error {
	var request SendToUserReq
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

	data := &RecvDataRsp{
		Id:   id,
		Uid:  uid,
		Chan: "",
		Data: request.Data,
	}
	if len(request.Uids) == 1 {
		cligrp, ok := h.room.ClientsOfUser(request.Uids[0])
		if !ok {
			return websocket.Error(rsp, ErrUserNotFound, "dest user not found", false)
		}
		go cligrp.Write(CmdRecvData, 0, data, 0, "", false)
	} else {
		cligrp := h.room.ClientsOfUsers(request.Uids)
		go cligrp.Write(CmdRecvData, 0, data, 0, "", false)
	}

	rsp.EncodeData(&SendToUserRsp{}, 0, "")
	return nil
}

func (h *handler) SendToChan(req websocket.Request, rsp websocket.Response) error {
	var request SendToChanReq
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

	data := &RecvDataRsp{
		Id:   id,
		Uid:  uid,
		Data: request.Data,
	}

	if len(request.Chans) == 1 {
		cligrp, ok := h.room.ClientsInChannel(request.Chans[0])
		if !ok {
			return websocket.Error(rsp, ErrChanNotFound, "dest chan not found", false)
		}
		data.Chan = request.Chans[0]
		go cligrp.Write(CmdRecvData, 0, data, 0, "", false)
	} else {
		go func() {
			for _, ch := range request.Chans {
				cligrp, ok := h.room.ClientsInChannel(ch)
				if ok {
					data.Chan = ch
					cligrp.Write(CmdRecvData, 0, data, 0, "", false)
				}
			}
		}()
	}

	rsp.EncodeData(&SendToChanRsp{}, 0, "")
	return nil
}

func (h *handler) RecvData(req websocket.Request, rsp websocket.Response) error {
	return websocket.Error(rsp, ErrWrongCmd, "wrong cmd", false)
}
