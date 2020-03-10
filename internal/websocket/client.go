package websocket

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
	rspData := &ResponseData{
		Cmd: cmd,
		Seq: seq,
		Code: code,
		Msg: msg,
		Data: data,
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(rspData)

	for _, c := range cg.clients {
		cli := c.(*client)
		cli.write(buf.Bytes(), immed)
	}
}

func NewClientGroup(clients []interface{}) ClientGroup {
	cg := &clientGroup{
		clients,
	}
	return cg
}

type client struct {
	svc        *server
	conn       *websocket.Conn
	ctx        context.Context
	writeq     []*ResponseData
	writer     bytes.Buffer
	writeTimer *time.Timer
	mtx        sync.Mutex
	cycle      time.Duration
	closed     bool
}

func (c *client) ContextValue(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *client) AddContextValue(key, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
}

func (c *client) Close() {
	c.conn.Close()
}

func (c *client) shutdown() {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return
	}

	c.conn.Close()
	c.writeTimer.Reset(0)
	c.closed = true
	c.mtx.Unlock()

	c.svc.closeHandler(c)
}

func (c *client) Write(cmd string, seq int64, data interface{}, code int32, msg string, immed bool) {
	rspData := &ResponseData{
		Cmd: cmd,
		Seq: seq,
		Code: code,
		Msg: msg,
		Data: data,
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(rspData)

	c.write(buf.Bytes(), immed)
}

func (c *client) write(json []byte, immed bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	
	if c.writer.Len() == 0 {
		c.writer.WriteByte('[')
		c.writer.Write(json)
		if immed {
			c.writeTimer.Reset(0)
		} else {
			c.writeTimer.Reset(c.cycle)
		}
	} else {
		c.writer.WriteByte(',')
		c.writer.Write(json)
	}
}

func (c *client) swap(writer io.Writer) (closed bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	
	if c.closed {
		return true
	}
	
	if c.writer.Len() == 0 {
		return false
	} else {
		c.writer.WriteByte(']')
	}
	
	c.writer.WriteTo(writer)  // write all and reset
	return false
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

	go c.writePump()

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
			log.Debugf("%09d received request: %v", time.Now().UnixNano() % int64(time.Second), reqData)

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
			en.Encode(rsp.data)
			c.write(buf.Bytes(), req.data.Immed)
		}
	}
}

func (c *client) writePump() {
	log.Debug("writePump start")
	defer func() {
		c.writeTimer.Stop()
		c.shutdown()
		log.Debug("writePump complete")
	}()

	var buf bytes.Buffer
	for {
		select {
		case <-c.writeTimer.C:
			buf.Reset()
			closed := c.swap(&buf)
			if closed {
				return
			}

			if buf.Len() == 0 {
				log.Debug("empty writeq")
				continue
			}

			log.Debugf("%09d sent Response: %s", time.Now().UnixNano() % int64(time.Second), buf.Bytes())
			if err := c.conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
				return
			}
		}
	}
}

func newClient(svc *server, conn *websocket.Conn, cycle time.Duration) *client {
	c := &client{
		svc:        svc,
		conn:       conn,
		ctx:        context.Background(),
		writeq:     make([]*ResponseData, 0, 100),
		writeTimer: time.NewTimer(cycle),
		cycle:      cycle,
		closed:     false,
	}
	return c
}
