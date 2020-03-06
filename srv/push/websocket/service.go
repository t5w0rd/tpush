package websocket

import (
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"net/http"
	"time"
)

var (
	Address       string        = "0.0.0.0:8080"
	LoginDeadline time.Duration = time.Second * 1
	ClientCycle                 = time.Second * 1

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Handler func(req *Request, rsp *Response) error
type LoginHandler func(req *Request, rsp *Response) (id clientid, err error)

type Request struct {
	Cmd  string
	Seq  int64
	Data interface{}
}

type Response struct {
	Cmd  string
	Seq  int64
	Code int32
	Data interface{}
}

type Service struct {
	hb           *hub
	mux          map[string]Handler
	loginHandler LoginHandler
}

type loginResult struct {
	err error
	uid int64
}

func NewService() *Service {
	s := &Service{
		hb:  newHub(),
		mux: make(map[string]Handler),
	}
	return s
}

func (s *Service) handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("New client has connected")

	cli := newClient(s, conn)
	s.hb.register(cli)

	go cli.readPump()
	go cli.writePump()
}

func (s *Service) RegisterCommandHandler(cmd string, handler Handler) {
	s.mux[cmd] = handler
}

func (s *Service) RegisterLoginHandler(handler LoginHandler) {
	s.loginHandler = handler
}

func (s *Service) Run() error {
	// 启动hub
	go s.hb.run()

	// 注册web服务处理器
	http.HandleFunc("/push", s.handleStream)

	// 启动web服务
	log.Infof("Server [web] Listening on %s", Address)
	err := http.ListenAndServe(Address, nil)
	if err != nil {
		log.Fatal("ListenAndServe err: ", err)
		return err
	}
	return nil
}
