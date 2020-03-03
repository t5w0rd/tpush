package handler

import (
	"context"
	"fmt"
	log "github.com/micro/go-micro/v2/logger"

	"github.com/micro/go-micro/v2/errors"
	gateway "tpush/srv/gateway/proto/gateway"
	push "tpush/srv/push/proto/push"
)

type Gateway struct {
	PushCli push.PushService
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Gateway) Call(ctx context.Context, req *gateway.Request, rsp *gateway.Response) error {
	log.Info("Received Gateway.Call request")
	pushRsp, err := e.PushCli.Call(ctx, &push.Request{
		Name: fmt.Sprintf("Mr. %s", req.Name),
	})
	if err != nil {
		return errors.InternalServerError("tpush.srv.push.push.call", err.Error())
	}

	rsp.Msg = pushRsp.Msg
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Gateway) Stream(ctx context.Context, req *gateway.StreamingRequest, stream gateway.Gateway_StreamStream) error {
	log.Infof("Received Gateway.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&gateway.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Gateway) PingPong(ctx context.Context, stream gateway.Gateway_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&gateway.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
