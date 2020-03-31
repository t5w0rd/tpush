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
	"tpush/web/route/plugins"
)

var (
	wrapper = plugins.NewClientWrapper()
)

type Handler struct {
	Etcd *clientv3.Client
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
	// TODO:
	log.Infof("%#v", nodes)

	rsp, err := pushCli.SendToUser(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	if err := json.NewEncoder(w).Encode(rsp); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
