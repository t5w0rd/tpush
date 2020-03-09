package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"sync"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Request interface {
	Command() string
	Sequence() int64
	DecodeData(data interface{}) error
	//ContextValue(key interface{}) interface{}
	Client() Client
}

type Response interface {
	EncodeData(data interface{}, code int32, msg string)
}

type RequestData struct {
	Cmd   string      `json:"cmd"`
	Seq   int64       `json:"seq"`
	Immed bool        `json:"immed,omitempty"`
	Data  interface{} `json:"data"`
}

type ResponseData struct {
	Cmd  string      `json:"cmd"`
	Seq  int64       `json:"seq"`
	Code int32       `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type request struct {
	data *RequestData
	cli  *client
}

func (req *request) Client() Client {
	return req.cli
}

func (req *request) Command() string {
	return req.data.Cmd
}

func (req *request) Sequence() int64 {
	return req.data.Seq
}

func (req *request) DecodeData(data interface{}) error {
	return mapstructure.Decode(req.data.Data, data)
}

//func (req *request) ContextValue(key interface{}) interface{} {
//	return req.cli.ctx.Value(key)
//}

type response struct {
	data *ResponseData
}

func (rsp *response) EncodeData(data interface{}, code int32, msg string) {
	rsp.data.Code = code
	rsp.data.Msg = msg
	rsp.data.Data = data
}

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
	mtx sync.RWMutex
	m   map[string]HandlerFunc
}

func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func (mux *ServeMux) HandleFunc(cmd string, handler HandlerFunc) {
	mux.mtx.Lock()
	defer mux.mtx.Unlock()

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
	mux.mtx.RLock()
	defer mux.mtx.RUnlock()

	if h, exist := mux.m[cmd]; exist {
		return h
	} else {
		return UnsupportedCommandHandler()
	}
}

type server struct {
	mux          *ServeMux
	clientCycle  time.Duration
	openHandler  OpenHandler
	closeHandler CloseHandler
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("new client has connected")

	cli := newClient(s, conn, s.clientCycle)
	go cli.run()
}

func Server(clientCycle time.Duration, mux *ServeMux, openHandler OpenHandler, closeHandler CloseHandler) http.Handler {
	s := &server{
		mux:          mux,
		clientCycle:  clientCycle,
		openHandler:  openHandler,
		closeHandler: closeHandler,
	}
	return s
}
