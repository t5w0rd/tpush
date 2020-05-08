package main

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/micro/cli/v2"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	"net/http/pprof"
	"tpush/options"
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

	// initialise service
	if err := service.Init(
		web.Action(func(c *cli.Context) {
			if f := c.String("log_level"); len(f) > 0 {
				loglevel, _ = log.GetLevel(f)
			}

			if f := c.String("enable_distribute"); len(f) > 0 {
				enable_distribute = c.Bool("enable_distribute")
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
	var c *clientv3.Client = nil
	if enable_distribute {
		storeAddress := options.EtcdAddress
		cfg := clientv3.Config{
			Endpoints: []string{storeAddress},
		}

		var err error
		c, err = clientv3.New(cfg)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	h := &handler.Handler{
		Etcd: c,
	}
	service.HandleFunc("/cmd/snd2usr", h.SendToUser)
	service.HandleFunc("/cmd/snd2chan", h.SendToChannel)

	service.HandleFunc("/debug/pprof/", pprof.Index)
	service.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	service.HandleFunc("/debug/pprof/profile", pprof.Profile)
	service.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	service.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
