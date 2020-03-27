package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/micro/go-micro/v2/errors"
	log "github.com/micro/go-micro/v2/logger"
	"tpush/internal/tchatroom"
	push "tpush/srv/push/proto/push"
)

type Push struct {
	Room *tchatroom.Room
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

func (h *Push) SendToClient(ctx context.Context, req *push.SendToClientReq, rsp *push.SendToClientRsp) error {
	data := &tchatroom.RecvDataRsp{
		Id:   req.Id,
		Uid:  req.Uid,
		Chan: "",
	}

	var buf *bytes.Buffer
	if req.Data != nil {
		buf = bytes.NewBuffer(req.Data)
	} else {
		buf = bytes.NewBufferString(req.Datastr)
	}
	if err := json.NewDecoder(buf).Decode(&data.Data); err != nil {
		return errors.InternalServerError("push.Push.SendToClient", err.Error())
	}

	if len(req.Ids) == 1 {
		cli, ok := h.Room.Client(req.Ids[0])
		if !ok {
			return errors.InternalServerError("push.Push.SendToClient", "dest client not found")
		}
		go cli.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
	} else {
		cligrp := h.Room.Clients(req.Ids)
		go cligrp.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
	}

	return nil
}

func (h *Push) SendToUser(ctx context.Context, req *push.SendToUserReq, rsp *push.SendToUserRsp) error {
	log.Infof("rpc SendToUser")

	data := &tchatroom.RecvDataRsp{
		Id:   req.Id,
		Uid:  req.Uid,
		Chan: "",
	}

	var buf *bytes.Buffer
	if req.Data != nil {
		buf = bytes.NewBuffer(req.Data)
	} else {
		buf = bytes.NewBufferString(req.Datastr)
	}
	if err := json.NewDecoder(buf).Decode(&data.Data); err != nil {
		return errors.InternalServerError("push.Push.SendToUser", err.Error())
	}

	if len(req.Uids) == 1 {
		cli, ok := h.Room.ClientsOfUser(req.Uids[0])
		if !ok {
			return errors.InternalServerError("push.Push.SendToUser", "dest user not found")
		}
		go cli.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
	} else {
		cligrp := h.Room.ClientsOfUsers(req.Uids)
		go cligrp.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
	}

	return nil
}

func (h *Push) SendToChannel(ctx context.Context, req *push.SendToChannelReq, rsp *push.SendToChannelRsp) error {
	data := &tchatroom.RecvDataRsp{
		Id:  req.Id,
		Uid: req.Uid,
	}

	var buf *bytes.Buffer
	if req.Data != nil {
		buf = bytes.NewBuffer(req.Data)
	} else {
		buf = bytes.NewBufferString(req.Datastr)
	}
	if err := json.NewDecoder(buf).Decode(&data.Data); err != nil {
		return errors.InternalServerError("push.Push.SendToChannel", err.Error())
	}

	if len(req.Chans) == 1 {
		data.Chan = req.Chans[0]
		cli, ok := h.Room.ClientsInChannel(req.Chans[0])
		if !ok {
			return errors.InternalServerError("push.Push.SendToChannel", "dest channel not found")
		}
		go cli.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
	} else {
		go func() {
			for _, ch := range req.Chans {
				cligrp, ok := h.Room.ClientsInChannel(ch)
				if ok {
					data.Chan = ch
					cligrp.Write(tchatroom.CmdRecvData, 0, data, 0, "", false)
				}
			}
		}()
	}

	return nil
}
