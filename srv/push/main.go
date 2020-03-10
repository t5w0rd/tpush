package main

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	"net/http"
	"time"
	"tpush/internal/websocket"
	"tpush/srv/push/handler"
	push "tpush/srv/push/proto/push"
	"tpush/srv/push/subscriber"
)

var (
	Address     = "0.0.0.0:8080"
	ClientCycle = time.Second * 1
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
				Value:   Address,
			},
			&cli.Float64Flag{
				Name:    "push_cycle",
				Usage:   "Set the client push cycle(seconds)",
				EnvVars: []string{"PUSH_CYCLE"},
				Value:   float64(ClientCycle / time.Second),
			},
		),
	)

	// Initialise service
	service.Init(
		micro.Action(func(c *cli.Context) error {
			if f := c.String("web_server_address"); len(f) > 0 {
				Address = f
			}

			if f := c.String("push_cycle"); len(f) > 0 {
				ClientCycle = time.Duration(float64(time.Second) * c.Float64("push_cycle"))
			}

			return nil
		}),
	)

	h := handler.NewPush()

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
	mux := websocket.NewServeMux()
	mux.HandleFunc("login", h.Login)
	mux.HandleFunc("enter", h.EnterChan)
	mux.HandleFunc("exit", h.ExitChan)
	mux.HandleFunc("hello", h.Hello)
	ws := websocket.Server(
		ClientCycle,
		mux,
		h.OnOpen,
		h.OnClose,
	)

	service2Done := make(chan struct{})

	go func() {
		// 注册web服务处理器
		http.Handle("/push", ws)

		http.Handle("/", http.FileServer(http.Dir("html")))

		// 启动web服务
		log.Infof("server [web] Listening on %s", Address)

		if err := http.ListenAndServe(Address, nil); err != nil {
			log.Fatal("server [web] Listening err: ", err)
		}
		close(service2Done)
	}()

	select {
	case <-serviceDone:
	case <-service2Done:
	}
}
