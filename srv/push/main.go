package main

import (
	"github.com/micro/go-micro/v2"
	_ "github.com/micro/go-micro/v2/client/selector"
	log "github.com/micro/go-micro/v2/logger"
	_ "github.com/micro/go-micro/v2/registry"
	_ "github.com/micro/go-plugins/wrapper/select/roundrobin"
	"tpush/srv/push/handler"
	"tpush/srv/push/subscriber"

	push "tpush/srv/push/proto/push"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("tpush.srv.push"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	push.RegisterPushHandler(service.Server(), new(handler.Push))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("tpush.srv.push", service.Server(), new(subscriber.Push))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
