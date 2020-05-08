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
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	defaultServeMux    = newDefaultServeMux()
	defaultSendTimeout = time.Second * 10
)

type UpgradeHandler func(req *http.Request) error
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

func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func pingHandler(req Request, rsp Response) error {
	return nil
}

func newDefaultServeMux() *ServeMux {
	mux := NewServeMux()
	mux.HandleFunc("ping", pingHandler)
	return mux
}

type server struct {
	opt *Options

	ready chan *client

	mu      sync.Mutex
	clients map[*client]struct{}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.opt.upgradeHandler != nil {
		if err := s.opt.upgradeHandler(r); err != nil {
			log.Error(err)
			return
		}
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("new client has connected", r.Header)

	cli := newClient(s, conn, s.opt.recvTimeout, s.opt.sendTimeout, false)
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

func (s *server) StartWritePumps(workers int) {
	for i := 0; i < workers; i++ {
		go s.writePump()
	}
}

func Server(opts ...Option) WritePumpHttpHandler {
	opt := new(Options)

	for _, o := range opts {
		o(opt)
	}

	if opt.mux == nil {
		opt.mux = defaultServeMux
	}
	if opt.sendTimeout == 0 {
		opt.sendTimeout = defaultSendTimeout
	}

	s := &server{
		opt:     opt,
		ready:   make(chan *client, 100000),
		clients: make(map[*client]struct{}),
	}
	return s
}
