package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"tpush/internal/tchatroom"
	"tpush/internal/websocket"

	"github.com/micro/go-micro/v2/client"
	push "tpush/srv/push/proto/push"
)

func handle(handler func (request *websocket.RequestData, response *websocket.ResponseData) error, w http.ResponseWriter, r *http.Request) {
	var request websocket.RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	response := &websocket.ResponseData{
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
	handle(func(request *websocket.RequestData, response *websocket.ResponseData) error {
		var req tchatroom.SendMsgToClientReq
		if err := request.DecodeData(&req); err != nil {
			return err
		}

		// call the backend service
		pushCli := push.NewPushService("tpush.srv.push", client.DefaultClient)
		if _, err := pushCli.SendMsgToClient(context.TODO(), &push.SendMsgToClientReq{
			Id: req.Id,
			Msg: req.Msg,
		}); err != nil {
			return err
		}

		rsp := &tchatroom.SendMsgToClientRsp{}
		response.EncodeData(rsp)

		return nil
	}, w, r)

	// decode the incoming request as json
	var request websocket.RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
