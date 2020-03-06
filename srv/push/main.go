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
	"tpush/srv/push/websocket"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("tpush.srv.push"),
		micro.Version("latest"),
		micro.Flags(
			&cli.StringFlag{
				Name:    "tpush_web_address",
				Usage:   "Set the web server address",
				EnvVars: []string{"TPUSH_WEB_ADDRESS"},
				Value:   websocket.Address,
			},
		),
	)

	// Initialise service
	service.Init(
		micro.Action(func(c *cli.Context) error {
			if f := c.String("tpush_web_address"); len(f) > 0 {
				websocket.Address = f
			}

			return nil
		}),
	)

	h := &handler.Push{}

	// Register Handler
	push.RegisterPushHandler(service.Server(), h)

	// Register Struct as Subscriber
	micro.RegisterSubscriber("tpush.srv.push", service.Server(), new(subscriber.Push))

	serviceDone := make(chan struct{})

	// Run service
	go func() {
		if err := service.Run(); err != nil {
			log.Fatal(err)
		}
		close(serviceDone)
	}()

	// websocket service
	service2 := websocket.NewService()
	service2.RegisterLoginHandler(h.Login)
	service2Done := make(chan struct{})

	go func() {
		if err := service2.Run(); err != nil {
			log.Fatal(err)
		}
		close(service2Done)
	}()

	select {
	case <-serviceDone:
	case <-service2Done:
	}
}
