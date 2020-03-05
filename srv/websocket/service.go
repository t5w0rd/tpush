package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	log "github.com/micro/go-micro/v2/logger"
	"time"
)

var (
	Address string = "0.0.0.0:8080"
	LoginDeadline time.Duration = time.Second * 1

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Request struct {
	Cmd string
	Seq int64
	Data interface{}
}

type Response struct {
	Cmd string
	Seq int64
	Code int32
	Data interface{}
}

type Service struct {
	hb *hub
	sendCycle time.Duration
}

type loginResult struct {
	err error
	uid int64
}

func NewService() *Service {
	s := &Service{
		hb: newHub(),
	}
	return s
}

func (s *Service) handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	result := make(chan *loginResult)
	defer close(result)

	go s.waitLogin(conn, result)

	var id int64
	ticker := time.NewTicker(LoginDeadline)
	select {
	case res := <- result:
		if res.err != nil {
			log.Error(res.err)
			conn.Close()
			return
		}
		id = res.uid
	case <- ticker.C:
		log.Info("Client hasnot logged in for a long time")
		conn.Close()
		<-result
		return
	}

	cli := newClient(conn, id)
	s.hb.register(cli)

	go cli.readPump()
	go cli.writePump()
}

// 等待客户端登陆
func (s *Service) waitLogin(conn *websocket.Conn, result chan *loginResult) {
	res := &loginResult{}

	defer func() {
		result <- res
	}()

	var req Request
	if res.err = conn.ReadJSON(&req); res.err != nil {
		return
	}

	rsp := &Response{
		Cmd: req.Cmd,
		Seq: req.Seq,
	}
	if res.err = conn.WriteJSON(rsp); res.err != nil {
		return
	}

	res.uid = req.Data.(map[string]int64)["uid"]
}

func (s *Service) Run() error {
	// 启动hub
	go s.hb.run()

	// 注册web服务处理器
	http.HandleFunc("/stream", s.handleStream)

	// 启动web服务
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe err: ", err)
		return err
	}
	return nil
}
