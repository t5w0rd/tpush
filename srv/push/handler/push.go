package handler

import (
	"context"
	"sync/atomic"
	"time"
	"tpush/internal/websocket"

	log "github.com/micro/go-micro/v2/logger"

	push "tpush/srv/push/proto/push"
)

type Push struct {
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Push) Call(ctx context.Context, req *push.Request, rsp *push.Response) error {
	log.Info("received Push.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Push) Stream(ctx context.Context, req *push.StreamingRequest, stream push.Push_StreamStream) error {
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
func (e *Push) PingPong(ctx context.Context, stream push.Push_PingPongStream) error {
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

type LoginReq struct {
	Uid int64 `json:"uid"`
}

type LoginRsp struct {
	Id int64 `json:"id"`
}

type HelloReq struct {
	Name string `json:"name"`
}

type HelloRsp struct {
	Say string `json:"say"`
}

type loginDoneKey struct{}

type clientDataKey struct{}

var (
	cliId int64 = 0
)

func genid() int64 {
	return atomic.AddInt64(&cliId, 1)
}

func (e *Push) OnOpen(cli websocket.Client) error {
	loginDone := make(chan int64)
	cli.AddContextValue(loginDoneKey{}, loginDone)
	cli.AddContextValue(clientDataKey{}, &clientData{
		id: genid(),
	})

	go func() {
		defer log.Debug("waitLogin complete")
		select {
		case id := <-loginDone:
			log.Debugf("client logged in succ, id: %v", id)
			return
		case <-time.After(time.Second * 2):
			log.Error("client hasnot logged in for a long time")
			cli.Close()
			return
		}
	}()
	return nil
}

func (e *Push) OnClose(cli websocket.Client) {
}

func (e *Push) Login(req websocket.Request, rsp websocket.Response) error {
	var loginReq LoginReq
	if err := req.DecodeData(&loginReq); err != nil {
		return err
	}
	id := loginReq.Uid

	loginDone := req.Client().ContextValue(loginDoneKey{}).(chan int64)
	loginDone <- id

	clientData := req.Client().ContextValue(clientDataKey{}).(*clientData)

	rsp.EncodeData(&LoginRsp{
		Id: clientData.id,
	}, 0, "")

	return nil
}

func (e *Push) Hello(req websocket.Request, rsp websocket.Response) error {
	var helloReq HelloReq
	if err := req.DecodeData(&helloReq); err != nil {
		return err
	}

	rsp.EncodeData(&HelloRsp{"Hello, " + helloReq.Name}, 0, "")

	return nil
}
