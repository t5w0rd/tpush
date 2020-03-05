package handler

import (
	"context"

	log "github.com/micro/go-micro/v2/logger"

	push "tpush/srv/push/proto/push"
)

type Push struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Push) Call(ctx context.Context, req *push.Request, rsp *push.Response) error {
	log.Info("Received Push.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Push) Stream(ctx context.Context, req *push.StreamingRequest, stream push.Push_StreamStream) error {
	log.Infof("Received Push.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
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
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&push.Pong{Stroke: req.Stroke + 1000}); err != nil {
			return err
		}
	}
}

// Session is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Push) Session(ctx context.Context, stream push.Push_SessionStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		_ = req
	}
}

type MidQueue = chan int64
type UidMidMap = map[int64] MidQueue

type channel struct {
	Name      string
	UserMsgMap UidMidMap
	Freq      float64
}

func NewChannel(name string, freq float64) (ch *channel) {
	ch = &channel{
		Name: name,
		UserMsgMap: make(UidMidMap),
		Freq: freq,
	}
	return ch
}

func (c *channel) pack_pusher() {

}

type Pusher struct {
	chanMap map[string]UidMidMap
}

