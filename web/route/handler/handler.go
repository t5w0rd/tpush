package handler

import (
	"context"
	"encoding/json"
	"github.com/micro/go-micro/v2/client/grpc"
	log "github.com/micro/go-micro/v2/logger"
	"net/http"
	"time"
	"tpush/internal/twebsocket"
	push "tpush/srv/push/proto/push"
	"tpush/web/route/plugins"
)

func handle(handler func(request *twebsocket.RequestData, response *twebsocket.ResponseData) error, w http.ResponseWriter, r *http.Request) {
	var request twebsocket.RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	response := &twebsocket.ResponseData{
		Cmd: request.Cmd,
		Seq: request.Seq,
	}
	if err := handler(&request, response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

var (
	wrapper = plugins.NewClientWrapper()
)

func SendMsgToUser(w http.ResponseWriter, r *http.Request) {
	var req push.SendToUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	cli := grpc.NewClient(
		//client.Wrap(wrapper),
	)
	services, err := cli.Options().Registry.GetService("tpush.srv.push")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	for i, service := range services {
		log.Infof("service %d: %#v", i, service)
		for j, node := range service.Nodes {
			log.Infof("node %d: %#v", j, node)
		}
	}

	pushCli := push.NewPushService("tpush.srv.push", cli)
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)
	rsp, err := pushCli.SendToUser(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	if err := json.NewEncoder(w).Encode(rsp); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
