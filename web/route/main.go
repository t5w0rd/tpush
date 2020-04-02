package main

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/micro/cli/v2"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	"tpush/web/route/handler"
)

func main() {
	// create new web service
	service := web.NewService(
		web.Name("tpush.web.route"),
		web.Version("latest"),
		web.Flags(
			&cli.StringFlag{
				Name:    "log_level",
				Usage:   "Set log level",
				EnvVars: []string{"LOG_LEVEL"},
				Value:   "debug",
			},
		),
	)

	var loglevel log.Level
	// initialise service
	if err := service.Init(
		web.Action(func(c *cli.Context) {
			if f := c.String("log_level"); len(f) > 0 {
				loglevel, _ = log.GetLevel(f)
			}
		}),
	); err != nil {
		log.Fatal(err)
	}

	if err := log.Init(
		log.WithLevel(loglevel),
	); err != nil {
		log.Fatal(err)
		return
	}

	// register call handler
	storeAddress := "10.8.9.100:52379"
	cfg := clientv3.Config{
		Endpoints: []string{storeAddress},
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	h := &handler.Handler{
		Etcd: c,
	}
	service.HandleFunc("/cmd/snd2usr", h.SendMsgToUser)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
