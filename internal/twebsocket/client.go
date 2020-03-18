package twebsocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"io"
	"sync"
	"time"
)

type Writer interface {
	Write(cmd string, seq int64, data interface{}, code int32, msg string, immed bool)
}

type Client interface {
	Writer
	ContextValue(key interface{}) interface{}
	AddContextValue(key, value interface{})
	Close()
}

type ClientGroup interface {
	Writer
	Clients(output *[]Client)
}

var (
	LeftSB = []byte("[")
	RightSB = []byte("]")
	Comma = []byte(",")
)

type clientGroup struct {
	clients []interface{}
}

func (cg *clientGroup) Clients(output *[]Client) {
	if size := len(cg.clients); cap(*output) < size {
		*output = make([]Client, size)
	} else {
		*output = (*output)[:size]
	}
	for i, o := range cg.clients {
		(*output)[i] = o.(Client)
	}
}

func (cg *clientGroup) Write(cmd string, seq int64, data interface{}, code int32, msg string, immed bool) {
	if len(cg.clients) == 0 {
		return
	}

	rspData := &ResponseData{
		Cmd:  cmd,
		Seq:  seq,
		Code: code,
		Msg:  msg,
		Data: EncodeData(data),
	}

	log.Info("clientgroup begin to encode")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(rspData); err != nil {
		return
	}

	log.Info("clientgroup begin to write")
	for _, c := range cg.clients {
		cli := c.(*client)
		cli.write(buf.Bytes(), immed)
	}
	log.Info("clientgroup write complete")
}

func NewClientGroup(clients []interface{}) ClientGroup {
	cg := &clientGroup{
		clients,
	}
	return cg
}

type client struct {
	svc    *server
	conn   *websocket.Conn
	ctx    context.Context
	//writer bytes.Buffer
	writeq [][]byte
	//writeTimer *time.Timer
	mu          sync.Mutex
	cycle       time.Duration
	closed      bool
	sendMu      sync.Mutex
	immedWriter bytes.Buffer
}

func (c *client) ContextValue(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *client) AddContextValue(key, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
}

func (c *client) Close() {
	_ = c.conn.Close()
}

func (c *client) shutdown() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}

	_ = c.conn.Close()
	//c.writeTimer.Reset(0)
	c.closed = true
	c.mu.Unlock()

	c.svc.closeHandler(c)
}

func (c *client) Write(cmd string, seq int64, data interface{}, code int32, msg string, immed bool) {
	rspData := &ResponseData{
		Cmd:  cmd,
		Seq:  seq,
		Code: code,
		Msg:  msg,
		Data: EncodeData(data),
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(rspData); err != nil {
		return
	}
	c.write(buf.Bytes(), immed)
}

func (c *client) write(json []byte, immed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if immed {
		c.immedWriter.Reset()
		c.immedWriter.WriteByte('[')
		c.immedWriter.Write(json)
		c.immedWriter.WriteByte(']')
		c.send(c.immedWriter.Bytes())
		return
	}

	if len(c.writeq) == 0 {
		c.writeq = append(c.writeq, LeftSB, json)
		c.svc.ready <- c
	} else {
		c.writeq = append(c.writeq, Comma, json)
	}
}

func (c *client) swap(writer io.Writer) (closed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return true
	}

	if len(c.writeq) == 0 {
		return false
	}

	for _, bs := range c.writeq {
		writer.Write(bs)
	}
	writer.Write(RightSB)
	c.writeq = c.writeq[:0]
	return false
}

func (c *client) send(data []byte) error {
	log.Debugf("%09d sent Response: %s", time.Now().UnixNano()%int64(time.Second), data)
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *client) run() error {
	log.Debug("run start")
	defer func() {
		c.shutdown()
		log.Debug("run complete")
	}()

	if err := c.svc.openHandler(c); err != nil {
		return err
	}

	//go c.writePump()

	var buf bytes.Buffer
	en := json.NewEncoder(&buf)
	for {
		var reqDatas []*RequestData
		if err := c.conn.ReadJSON(&reqDatas); err != nil {
			if !websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 意外的错误
				log.Error(err)
			} else {
				log.Debug(err)
			}
			return err
		}

		for _, reqData := range reqDatas {
			if reqData == nil {
				err := errors.New("nil request")
				log.Error(err)
				return err
			}
			log.Debugf("%09d received request: %v", time.Now().UnixNano()%int64(time.Second), reqData)

			handler := c.svc.mux.Handler(reqData.Cmd)
			req := &request{
				data: reqData,
				cli:  c,
			}
			rsp := &response{
				data: &ResponseData{
					Cmd: reqData.Cmd,
					Seq: reqData.Seq,
				},
			}
			if err := handler(req, rsp); err != nil {
				// 发生错误关闭连接
				log.Error(err)
				return err
			}

			buf.Reset()
			if err := en.Encode(rsp.data); err != nil {
				log.Error(err)
				return err
			}
			c.write(buf.Bytes(), req.data.Immed)
		}
	}
}

//func (c *client) writePump() {
//	log.Debug("writePump start")
//	defer func() {
//		c.writeTimer.Stop()
//		c.shutdown()
//		log.Debug("writePump complete")
//	}()
//
//	var buf bytes.Buffer
//	for {
//		select {
//		case <-c.writeTimer.C:
//			buf.Reset()
//			closed := c.swap(&buf)
//			if closed {
//				return
//			}
//
//			if buf.Len() == 0 {
//				log.Debug("empty writeq")
//				continue
//			}
//
//			if err := c.send(buf.Bytes()); err != nil {
//				return
//			}
//		}
//	}
//}

func newClient(svc *server, conn *websocket.Conn, cycle time.Duration) *client {
	c := &client{
		svc:  svc,
		conn: conn,
		ctx:  context.Background(),
		//writeTimer: time.NewTimer(cycle),
		cycle:  cycle,
		closed: false,
	}
	return c
}
