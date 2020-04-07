package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	log "github.com/micro/go-micro/v2/logger"
	"net/http"
	"time"
	"tpush/internal"
	"tpush/internal/tchatroom"
	push "tpush/srv/push/proto/push"
	route "tpush/web/route/proto"
	"tpush/web/route/wrapper"
)

var (
	clientWrapper = wrapper.NewClientWrapper()
)

type Handler struct {
	Etcd    *clientv3.Client
	PushCli push.PushService
}

func (h *Handler) SendToUser(w http.ResponseWriter, r *http.Request) {
	var req route.SendToUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if h.PushCli == nil {
		opts := make([]client.Option, 0)
		if h.Etcd != nil {
			opts = append(opts, client.Wrap(clientWrapper))
		}
		cli := grpc.NewClient(opts...)
		h.PushCli = push.NewPushService("tpush.srv.push", cli)
	}

	if h.Etcd == nil {
		data, err := json.Marshal(req.Data)
		if err != nil {
			log.Error(err)
			return
		}

		pushReq := &push.SendToUserReq{
			Uids: req.Uids,
			Data: data,
			Id:   req.Id,
			Uid:  req.Uid,
		}
		log.Info("SendMsgToUser")
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)
		_, err = h.PushCli.SendToUser(ctx, pushReq)
		if err != nil {
			log.Error(err)
		}
	} else {
		keys := make([]string, len(req.Uids))
		for i, uid := range req.Uids {
			keys[i] = fmt.Sprintf(tchatroom.RegUserKeyFmt, uid)
		}
		nodes := internal.GetDistributeNodes(h.Etcd, keys, time.Millisecond*1000)
		for id, _ := range nodes {
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)
			ctx = context.WithValue(ctx, wrapper.SelectNodeKey{}, id)
			go func(ctx context.Context) {
				data, err := json.Marshal(req.Data)
				if err != nil {
					log.Error(err)
					return
				}

				pushReq := &push.SendToUserReq{
					Uids: req.Uids,
					Data: data,
					Id:   req.Id,
					Uid:  req.Uid,
				}
				log.Info("SendToUser")
				_, err = h.PushCli.SendToUser(ctx, pushReq)
				if err != nil {
					log.Error(err)
				}
			}(ctx)
		}
	}

	if err := json.NewEncoder(w).Encode(&route.SendToRsp{}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) SendToChannel(w http.ResponseWriter, r *http.Request) {
	var req route.SendToChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	opts := make([]client.Option, 0)
	if h.Etcd != nil {
		opts = append(opts, client.Wrap(clientWrapper))
	}
	cli := grpc.NewClient(opts...)

	if h.PushCli == nil {
		h.PushCli = push.NewPushService("tpush.srv.push", cli)
	}

	if h.Etcd == nil {
		data, err := json.Marshal(req.Data)
		if err != nil {
			log.Error(err)
			return
		}

		pushReq := &push.SendToChannelReq{
			Chans: req.Chans,
			Data: data,
			Id:   req.Id,
			Uid:  req.Uid,
		}
		log.Info("SendToChannel")
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)
		_, err = h.PushCli.SendToChannel(ctx, pushReq)
		if err != nil {
			log.Error(err)
		}
	} else {
		keys := make([]string, len(req.Chans))
		for i, uid := range req.Chans {
			keys[i] = fmt.Sprintf(tchatroom.RegChannelKeyFmt, uid)
		}
		nodes := internal.GetDistributeNodes(h.Etcd, keys, time.Millisecond*1000)
		for id, _ := range nodes {
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)
			ctx = context.WithValue(ctx, wrapper.SelectNodeKey{}, id)
			go func(ctx context.Context) {
				data, err := json.Marshal(req.Data)
				if err != nil {
					log.Error(err)
					return
				}

				pushReq := &push.SendToChannelReq{
					Chans: req.Chans,
					Data: data,
					Id:   req.Id,
					Uid:  req.Uid,
				}
				log.Info("SendMsgToChannel")
				_, err = h.PushCli.SendToChannel(ctx, pushReq)
				if err != nil {
					log.Error(err)
				}
			}(ctx)
		}
	}

	if err := json.NewEncoder(w).Encode(&route.SendToRsp{}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
