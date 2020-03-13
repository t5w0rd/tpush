package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"tpush/internal/tchatroom"
	"tpush/internal/twebsocket"

	"github.com/micro/go-micro/v2/client"
	push "tpush/srv/push/proto/push"
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

func SendMsgToClient(w http.ResponseWriter, r *http.Request) {
	handle(func(request *twebsocket.RequestData, response *twebsocket.ResponseData) error {
		var req tchatroom.SendToClientReq
		if err := twebsocket.DecodeData(request, &req); err != nil {
			return err
		}

		// call the backend service
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(req.Data); err != nil {
			return err
		}
		cliReq := &push.SendToClientReq{
			Ids:  req.Ids,
			Data: buf.Bytes(),
		}

		pushCli := push.NewPushService("tpush.srv.push", client.DefaultClient)
		if _, err := pushCli.SendToClient(context.TODO(), cliReq); err != nil {
			return err
		}

		rsp := &tchatroom.SendToClientRsp{}
		response.Data = twebsocket.EncodeData(rsp)

		return nil
	}, w, r)

	// decode the incoming request as json
	var request twebsocket.RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
