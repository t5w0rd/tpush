package main

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	"time"
	"tpush/internal/tchatroom"
	"tpush/srv/push/handler"
	push "tpush/srv/push/proto/push"
	"tpush/srv/push/subscriber"
)

func main() {
	if err := log.Init(
		log.WithLevel(log.DebugLevel),
	); err != nil {
		log.Fatal(err)
		return
	}

	// New Service
	service := micro.NewService(
		micro.Name("tpush.srv.push"),
		micro.Version("latest"),
		micro.Flags(
			&cli.StringFlag{
				Name:    "web_server_address",
				Usage:   "Set the web server address",
				EnvVars: []string{"WEB_SERVER_ADDRESS"},
				Value:   tchatroom.Address,
			},
			&cli.Float64Flag{
				Name:    "push_cycle",
				Usage:   "Set the client push cycle(seconds)",
				EnvVars: []string{"PUSH_CYCLE"},
				Value:   float64(tchatroom.ClientCycle / time.Second),
			},
			&cli.StringFlag{
				Name:    "stream_pattern",
				Usage:   "Set the web server stream pattern",
				EnvVars: []string{"STREAM_PATTERN"},
				Value:   tchatroom.StreamPattern,
			},
		),
	)

	// Initialise service
	service.Init(
		micro.Action(func(c *cli.Context) error {
			if f := c.String("web_server_address"); len(f) > 0 {
				tchatroom.Address = f
			}

			if f := c.String("push_cycle"); len(f) > 0 {
				tchatroom.ClientCycle = time.Duration(float64(time.Second) * c.Float64("push_cycle"))
			}

			if f := c.String("stream_pattern"); len(f) > 0 {
				tchatroom.StreamPattern = f
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
	service2 := tchatroom.NewService()

	service2Done := make(chan struct{})

	go func() {
		// 启动web服务
		log.Infof("server [web] Listening on %s", tchatroom.Address)
		if err := service2.Run(); err != nil {
			log.Fatal("server [web] Listening err: ", err)
		}
		close(service2Done)
	}()

	select {
	case <-serviceDone:
	case <-service2Done:
	}
}
