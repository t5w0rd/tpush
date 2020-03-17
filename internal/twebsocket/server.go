package twebsocket

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"net/http"
	"sync"
	"time"
)

type WritePumpHttpHandler interface {
	http.Handler
	StartWritePumps(workers int)
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type OpenHandler func(cli Client) error
type CloseHandler func(cli Client)
type HandlerFunc func(req Request, rsp Response) error

func Error(rsp Response, code int32, msg string, closeConnection bool) error {
	rsp.EncodeData(nil, code, msg)
	if closeConnection {
		err := fmt.Errorf("%s(%d)", msg, code)
		return err
	}
	return nil
}

func Fatal(rsp Response, err error) error {
	return err
}

func UnsupportedCommand(req Request, rsp Response) error {
	return Fatal(rsp, fmt.Errorf("unsupported command:%s", req.Command()))
}

func UnsupportedCommandHandler() HandlerFunc {
	return HandlerFunc(UnsupportedCommand)
}

type ServeMux struct {
	mu sync.RWMutex
	m  map[string]HandlerFunc
}

func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func (mux *ServeMux) HandleFunc(cmd string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if cmd == "" {
		panic("http: invalid pattern")
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if _, exist := mux.m[cmd]; exist {
		panic("http: multiple registrations for " + cmd)
	}

	if mux.m == nil {
		mux.m = make(map[string]HandlerFunc)
	}
	mux.m[cmd] = handler
}

func (mux *ServeMux) Handler(cmd string) (h HandlerFunc) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	if h, exist := mux.m[cmd]; exist {
		return h
	} else {
		return UnsupportedCommandHandler()
	}
}

type server struct {
	mux          *ServeMux
	PushCycle    time.Duration
	openHandler  OpenHandler
	closeHandler CloseHandler

	ready chan *client
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("new client has connected")

	cli := newClient(s, conn, s.PushCycle)
	go cli.run()
}

func (s *server) writePump() {
	buf := new(bytes.Buffer)
	for {
		select {
		case cli, ok := <-s.ready:
			if !ok {
				return
			}
			buf.Reset()
			cli.swap(buf)
			cli.send(buf.Bytes())
		}
	}
}

func (s *server) StartWritePumps(workers int) {
	for i:=0; i<workers; i++ {
		go s.writePump()
	}
}

func Server(pushCycle time.Duration, mux *ServeMux, openHandler OpenHandler, closeHandler CloseHandler) WritePumpHttpHandler {
	s := &server{
		mux:          mux,
		PushCycle:    pushCycle,
		openHandler:  openHandler,
		closeHandler: closeHandler,

		ready:        make(chan *client, 100000),
	}
	return s
}
