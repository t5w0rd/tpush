package main

import (
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	"tpush/srv/gateway/handler"
	"tpush/srv/gateway/plugins"
	"tpush/srv/gateway/subscriber"

	gateway "tpush/srv/gateway/proto/gateway"
	push "tpush/srv/push/proto/push"
)

func main() {
	wrapper := plugins.NewClientWrapper()

	// New Service
	service := micro.NewService(
		micro.Name("tpush.srv.gateway"),
		micro.Version("latest"),
		micro.WrapClient(wrapper),
	)

	// Initialise service
	service.Init()

	// Register Handler
	h := &handler.Gateway{
		PushCli: push.NewPushService("tpush.srv.push", service.Client()),
	}
	gateway.RegisterGatewayHandler(service.Server(), h)

	// Register Struct as Subscriber
	micro.RegisterSubscriber("tpush.srv.gateway", service.Server(), new(subscriber.Gateway))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
