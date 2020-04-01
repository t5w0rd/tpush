package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"net/http"
	"time"
	"tpush/internal"
	"tpush/internal/tchatroom"
	push "tpush/srv/push/proto/push"
	"tpush/web/route/plugins"
)

var (
	wrapper = plugins.NewClientWrapper()
)

type Handler struct {
	Etcd *clientv3.Client
}

type SendMsgRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (h *Handler) SendMsgToUser(w http.ResponseWriter, r *http.Request) {
	var req push.SendToUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	cli := grpc.NewClient(
		client.Wrap(wrapper),
	)

	pushCli := push.NewPushService("tpush.srv.push", cli)
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*1000)

	keys := make([]string, len(req.Uids))
	for i, uid := range req.Uids {
		keys[i] = fmt.Sprintf(tchatroom.RegUserKeyFmt, uid)
	}

	nodes := internal.GetDistributeNodes(h.Etcd, keys, time.Millisecond*1000)
	for id, _ := range nodes {
		newCtx := context.WithValue(ctx, plugins.SelectNodeKey{}, id)
		go func(ctx context.Context) {
			pushCli.SendToUser(ctx, &req)
		}(newCtx)
	}

	if err := json.NewEncoder(w).Encode(&SendMsgRsp{}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
