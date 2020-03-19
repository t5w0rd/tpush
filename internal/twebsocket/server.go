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
	mux        *ServeMux
	pingPeriod time.Duration
	pongWait   time.Duration
	sendWait   time.Duration

	openHandler  OpenHandler
	closeHandler CloseHandler

	ready chan *client

	mu      sync.Mutex
	clients map[*client]struct{}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("new client has connected")

	cli := newClient(s, conn, s.pongWait, s.sendWait)
	s.mu.Lock()
	s.clients[cli] = struct{}{}
	s.mu.Unlock()
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

func (s *server) pingPump() {
	clients := make([]*client, 0, 10000)
	disconnected := make([]*client, 0, 10000)
	t := time.NewTicker(s.pingPeriod)
	for {
		select {
		case <-t.C:
			clients = clients[:0]
			s.mu.Lock()
			for cli, _ := range s.clients {
				clients = append(clients, cli)
			}
			s.mu.Unlock()

			disconnected = disconnected[:0]
			for _, cli := range clients {
				if err := cli.ping(); err != nil {
					cli.Close()
					disconnected = append(disconnected, cli)
				}
			}

			s.mu.Lock()
			for _, cli := range disconnected {
				delete(s.clients, cli)
			}
			s.mu.Unlock()
		}
	}
}

func (s *server) StartWritePumps(workers int) {
	go s.pingPump()
	for i := 0; i < workers; i++ {
		go s.writePump()
	}
}

func Server(pingPeriod time.Duration, mux *ServeMux, openHandler OpenHandler, closeHandler CloseHandler) WritePumpHttpHandler {
	s := &server{
		mux:        mux,
		pingPeriod: pingPeriod,
		pongWait:   pingPeriod + time.Second*2,
		sendWait:   time.Second * 10,

		openHandler:  openHandler,
		closeHandler: closeHandler,

		ready:   make(chan *client, 100000),
		clients: make(map[*client]struct{}),
	}
	return s
}
