package main

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	_ "github.com/micro/go-micro/v2/client/selector"
	log "github.com/micro/go-micro/v2/logger"
	_ "github.com/micro/go-micro/v2/registry"
	_ "github.com/micro/go-plugins/wrapper/select/roundrobin"
	"tpush/srv/push/handler"
	push "tpush/srv/push/proto/push"
	"tpush/srv/push/subscriber"
	"tpush/srv/websocket"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("tpush.srv.push"),
		micro.Version("latest"),
		micro.Flags(
			&cli.StringFlag{
				Name:  "ws_address",
				Usage: "Set the websocket address",
				EnvVars: []string{"MICRO_WS_ADDRESS"},
				Value: websocket.Address,
			},
		),
	)

	// Initialise service
	service.Init(
		micro.Action(func(c *cli.Context) error {
			if f := c.String("ws_address"); len(f) > 0 {
				websocket.Address = f
			}

			return nil
		}),
	)

	// Register Handler
	push.RegisterPushHandler(service.Server(), new(handler.Push))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("tpush.srv.push", service.Server(), new(subscriber.Push))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
