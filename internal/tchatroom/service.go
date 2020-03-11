package tchatroom

import (
	"net/http"
	"time"
	"tpush/internal/websocket"
)

const (
	CmdLogin           = "login"
	CmdEnter           = "enter"
	CmdExit            = "exit"
	CmdSendMsgToClient = "snd2cli"
	CmdSendMsgToUser   = "snd2usr"
	CmdSendMsgToChan   = "snd2chan"
	CmdRecvMsg         = "rcvmsg"

	ErrNotLogin       = -11
	ErrLoginFailed    = -12
	ErrUnsupportedCmd = -21
	ErrWrongCmd       = -22
	ErrClientNotFound = -41
	ErrUserNotFound   = -42
	ErrChanNotFound   = -43
)

var (
	Address       = "0.0.0.0:8080"
	ClientCycle   = time.Second * 1
	StreamPattern = "/stream"
)

type Service struct {
	mux  *websocket.ServeMux
	Room *Room
}

func (s *Service) Run() error {
	return http.ListenAndServe(Address, nil)
}

func NewService() *Service {
	r := NewRoom()
	h := &handler{
		room: r,
	}
	mux := websocket.NewServeMux()
	mux.HandleFunc(CmdLogin, h.Login)
	mux.HandleFunc(CmdEnter, h.EnterChan)
	mux.HandleFunc(CmdExit, h.ExitChan)
	mux.HandleFunc(CmdSendMsgToClient, h.SendMsgToClient)
	mux.HandleFunc(CmdSendMsgToUser, h.SendMsgToUser)
	mux.HandleFunc(CmdSendMsgToChan, h.SendMsgToChan)
	mux.HandleFunc(CmdRecvMsg, h.RecvMsg)
	ws := websocket.Server(
		ClientCycle,
		mux,
		h.OnOpen,
		h.OnClose,
	)

	// 注册web服务处理器
	http.Handle(StreamPattern, ws)

	http.Handle("/", http.FileServer(http.Dir("html")))

	s := &Service{
		mux:  mux,
		Room: r,
	}
	return s
}
