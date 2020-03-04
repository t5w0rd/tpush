package handler

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"time"

	"github.com/micro/go-micro/v2/client"
	log "github.com/micro/go-micro/v2/logger"
	push "tpush/srv/push/proto/push"
)

func PushCall(w http.ResponseWriter, r *http.Request) {
	// decode the incoming request as json
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// call the backend service
	pushCli := push.NewPushService("tpush.srv.push", client.DefaultClient)
	rsp, err := pushCli.Call(context.TODO(), &push.Request{
		Name: request["name"].(string),
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// we want to augment the response
	response := map[string]interface{}{
		"msg": rsp.Msg,
		"ref": time.Now().UnixNano(),
	}

	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// stream
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func streamLoop(pushCli push.PushService, conn *websocket.Conn) error {
	// Read initial request from websocket
	var req push.StreamingRequest
	err := conn.ReadJSON(&req)
	if err != nil {
		return err
	}

	// Even if we aren't expecting further requests from the websocket, we still need to read from it to ensure we
	// get close signals
	// 客户端每次调用websocket.send()，在服务端都会对应生成一个Reader，这里样例代码只需要一个请求，多余的丢弃了。
	go func() {
		for {
			if mt, r, err := conn.NextReader(); err != nil {
				break
			} else {
				_ = r
				log.Infof("messageType: %v", mt)
			}
		}
	}()

	log.Infof("Received Request: %v", req)

	// Send request to strStream server
	strStream, err := pushCli.Stream(context.Background(), &req)
	if err != nil {
		return err
	}
	defer strStream.Close()

	// Read from the strStream server and pass responses on to websocket
	for {
		// Read from strStream, end request once the strStream is closed
		rsp, err := strStream.Recv()
		if err != nil {
			if err != io.EOF {
				return err
			}

			break
		}

		// Write server response to the websocket
		log.Infof("Send Response: %v", rsp)
		err = conn.WriteJSON(rsp)
		if err != nil {
			// End request if socket is closed
			if isExpectedClose(err) {
				log.Info("Expected Close on socket", err)
				break
			} else {
				return err
			}
		}
	}

	return nil
}

func isExpectedClose(err error) bool {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
		log.Info("Unexpected websocket close: ", err)
		return false
	}

	return true
}

func PushStream(w http.ResponseWriter, r *http.Request) {
	// Upgrade request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Upgrade: ", err)
		return
	}
	defer conn.Close()

	// Handle websocket request
	pushCli := push.NewPushService("tpush.srv.push", client.DefaultClient)
	if err := streamLoop(pushCli, conn); err != nil {
		log.Fatal("Echo: ", err)
		return
	}
	log.Info("Stream complete")
}

func clientPump(strPingPong push.Push_PingPongService, conn *websocket.Conn) error {
	defer func() {
		conn.Close()
		strPingPong.Close()
		log.Infof("clientPump complete")
	}()

	var req push.Ping
	for {
		err := conn.ReadJSON(&req)
		if err != nil {
			return err
		}

		err = strPingPong.Send(&req)
		if err != nil {
			return err
		}
	}
}

func serverPump(strPingPong push.Push_PingPongService, conn *websocket.Conn, ret chan error) {
	defer func() {
		strPingPong.Close()
		conn.Close()
		log.Infof("serverPump complete")
	}()

	for {
		rsp, err := strPingPong.Recv()
		if err != nil {
			ret <- err
			return
		}

		err = conn.WriteJSON(rsp)
		if err != nil {
			// End request if socket is closed
			if isExpectedClose(err) {
				log.Info("Expected Close on socket", err)
				ret <- nil
				return
			} else {
				ret <- err
				return
			}
		}
	}
}

func pingPongLoop(pushCli push.PushService, conn *websocket.Conn) error {
	// Read initial request from websocket
	strPingPong, err := pushCli.PingPong(context.Background())
	if err != nil {
		return err
	}

	ret := make(chan error)
	go serverPump(strPingPong, conn, ret)
	clientPump(strPingPong, conn)

	<-ret

	return nil
}

func PushPingPong(w http.ResponseWriter, r *http.Request) {
	// Upgrade request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Upgrade: ", err)
		return
	}
	defer conn.Close()

	// Handle websocket request
	pushCli := push.NewPushService("tpush.srv.push", client.DefaultClient)
	if err := pingPongLoop(pushCli, conn); err != nil {
		log.Fatal("Echo: ", err)
		return
	}
	log.Info("PingPong complete")
}
