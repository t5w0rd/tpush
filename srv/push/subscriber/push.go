package subscriber

import (
	"context"
	log "github.com/micro/go-micro/v2/logger"

	push "tpush/srv/push/proto/push"
)

type Push struct{}

func (e *Push) Handle(ctx context.Context, msg *push.Message) error {
	log.Info("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *push.Message) error {
	log.Info("Function Received message: ", msg.Say)
	return nil
}
