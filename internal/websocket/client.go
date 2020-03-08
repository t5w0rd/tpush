package websocket

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"sync"
	"time"
)

type client struct {
	svc        *server
	conn       *websocket.Conn
	ctx        context.Context
	writeq     []*ResponseData
	writeTimer *time.Timer
	mtx        sync.Mutex
	cycle      time.Duration
	closed     bool
}

func (c *client) close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return
	}

	c.conn.Close()
	c.writeTimer.Reset(0)
	c.svc.closeHandler(c.ctx)
	c.closed = true
}

func (c *client) write(rsp *ResponseData, immediately bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if len(c.writeq) == 0 {
		if immediately {
			c.writeTimer.Reset(0)
		} else {
			c.writeTimer.Reset(c.cycle)
		}
	}
	c.writeq = append(c.writeq, rsp)
}

func (c *client) swap() (rsps []*ResponseData, closed bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return nil, true
	}
	if len(c.writeq) == 0 {
		return nil, false
	}
	rsps = c.writeq
	c.writeq = make([]*ResponseData, 0, 100)
	return rsps, false
}

func (c *client) run() error {
	log.Debug("run start")
	defer func() {
		c.close()
		log.Debug("run complete")
	}()

	if ctx, err := c.svc.openHandler(c.ctx, func() {
		c.conn.Close()
	}); err != nil {
		return err
	} else {
		c.ctx = ctx
	}

	go c.writePump()

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
			log.Debugf("received request: %v", reqData)

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
			c.write(rsp.data, req.data.Immed)
		}
	}
}

func (c *client) writePump() {
	log.Debug("writePump start")
	defer func() {
		c.writeTimer.Stop()
		c.close()
		log.Debug("writePump complete")
	}()

	for {
		select {
		case <-c.writeTimer.C:
			rspDatas, closed := c.swap()
			if closed {
				return
			}

			if rspDatas == nil {
				log.Debug("empty sendq")
				continue
			}

			log.Debugf("sent Response: %v", rspDatas)
			if err := c.conn.WriteJSON(rspDatas); err != nil {
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
