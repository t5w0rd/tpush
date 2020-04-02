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
	Etcd *clientv3.Client
}

func (h *Handler) SendMsgToUser(w http.ResponseWriter, r *http.Request) {
	var req route.SendToUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	cli := grpc.NewClient(
		client.Wrap(clientWrapper),
	)

	pushCli := push.NewPushService("tpush.srv.push", cli)

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
			pushCli.SendToUser(ctx, pushReq)
		}(ctx)
	}

	if err := json.NewEncoder(w).Encode(&route.SendMsgRsp{}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
