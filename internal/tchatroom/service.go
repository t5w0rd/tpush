package tchatroom

import (
	"net/http"
	"runtime"
	"time"
	"tpush/internal/twebsocket"
)

const (
	CmdLogin        = "login"
	CmdEnter        = "enter"
	CmdExit         = "exit"
	CmdSendToClient = "snd2cli"
	CmdSendToUser   = "snd2usr"
	CmdSendToChan   = "snd2chan"
	CmdRecvData     = "rcvdata"

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
	PushCycle     = time.Second * 1
	LoginTimeout  = time.Second * 2
	StreamPattern = "/stream"
)

type Service struct {
	mux  *twebsocket.ServeMux
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
	mux := twebsocket.NewServeMux()
	mux.HandleFunc(CmdLogin, h.Login)
	mux.HandleFunc(CmdEnter, h.EnterChan)
	mux.HandleFunc(CmdExit, h.ExitChan)
	mux.HandleFunc(CmdSendToClient, h.SendToClient)
	mux.HandleFunc(CmdSendToUser, h.SendToUser)
	mux.HandleFunc(CmdSendToChan, h.SendToChan)
	mux.HandleFunc(CmdRecvData, h.RecvData)
	ws := twebsocket.Server(
		PushCycle,
		mux,
		h.OnOpen,
		h.OnClose,
	)
	ws.StartWritePumps(runtime.NumCPU())

	// 注册web服务处理器
	http.Handle(StreamPattern, ws)

	http.Handle("/", http.FileServer(http.Dir("html")))

	s := &Service{
		mux:  mux,
		Room: r,
	}
	return s
}
