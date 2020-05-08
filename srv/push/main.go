package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	_ "net/http/pprof"
	"time"
	"tpush/internal/tchatroom"
	"tpush/options"
	"tpush/srv/push/handler"
	push "tpush/srv/push/proto/push"
	"tpush/srv/push/subscriber"
)

func main() {
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
				Name:    "recv_timeout",
				Usage:   "Set the client recv timeout(seconds)",
				EnvVars: []string{"RECV_TIMEOUT"},
				Value:   float64(tchatroom.RecvTimeout / time.Second),
			},
			&cli.Float64Flag{
				Name:    "login_timeout",
				Usage:   "Set login timeout(seconds)",
				EnvVars: []string{"LOGIN_TIMEOUT"},
				Value:   float64(tchatroom.LoginTimeout / time.Second),
			},
			&cli.StringFlag{
				Name:    "stream_pattern",
				Usage:   "Set the web server stream pattern",
				EnvVars: []string{"STREAM_PATTERN"},
				Value:   tchatroom.StreamPattern,
			},
			&cli.StringFlag{
				Name:    "log_level",
				Usage:   "Set log level",
				EnvVars: []string{"LOG_LEVEL"},
				Value:   "debug",
			},
			&cli.BoolFlag{
				Name:    "enable_distribute",
				Usage:   "enable distribute",
				EnvVars: []string{"ENABLE_DISTRIBUTE"},
				Value:   false,
			},
		),
	)

	var loglevel log.Level
	var enable_distribute bool
	// Initialise service
	service.Init(
		micro.Action(func(c *cli.Context) error {
			if f := c.String("web_server_address"); len(f) > 0 {
				tchatroom.Address = f
			}

			if f := c.String("recv_timeout"); len(f) > 0 {
				tchatroom.RecvTimeout = time.Duration(float64(time.Second) * c.Float64("recv_timeout"))
			}

			if f := c.String("login_timeout"); len(f) > 0 {
				tchatroom.LoginTimeout = time.Duration(float64(time.Second) * c.Float64("login_timeout"))
			}

			if f := c.String("stream_pattern"); len(f) > 0 {
				tchatroom.StreamPattern = f
			}

			if f := c.String("log_level"); len(f) > 0 {
				loglevel, _ = log.GetLevel(f)
			}

			if f := c.String("enable_distribute"); len(f) > 0 {
				enable_distribute = c.Bool("enable_distribute")
			}

			return nil
		}),
	)

	if err := log.Init(
		log.WithLevel(loglevel),
	); err != nil {
		log.Fatal(err)
		return
	}

	// websocket service
	opts := make([]tchatroom.Option, 0)
	if enable_distribute {
		storeAddress := options.EtcdAddress
		cfg := clientv3.Config{
			Endpoints: []string{storeAddress},
		}
		c, err := clientv3.New(cfg)
		if err != nil {
			log.Fatal(err)
			return
		}

		o := service.Server().Options()
		nodeId := fmt.Sprintf("%s-%s", o.Name, o.Id)
		d := tchatroom.NewEtcdDistribute(nodeId, c, time.Second*30)
		d.Run()

		opts = append(opts, tchatroom.WithDistribute(d))
	}

	service2 := tchatroom.NewService(opts...)

	service2Done := make(chan struct{})

	go func() {
		// 启动web服务
		log.Infof("Server [web] Listening on %s", tchatroom.Address)
		if err := service2.Run(); err != nil {
			log.Fatal("Server [web] Listening err: ", err)
		}
		close(service2Done)
	}()

	h := &handler.Push{
		Room: service2.Room,
	}

	// Register Handler
	push.RegisterPushHandler(service.Server(), h)

	// Register Struct as Subscriber
	sub := &subscriber.Push{}
	micro.RegisterSubscriber("tpush.srv.push", service.Server(), sub)

	serviceDone := make(chan struct{})

	// Run service
	log.Infof(service.Server().Options().Id)
	go func() {
		if err := service.Run(); err != nil {
			log.Fatal(err)
		}
		close(serviceDone)
	}()

	select {
	case <-serviceDone:
	case <-service2Done:
	}
}
