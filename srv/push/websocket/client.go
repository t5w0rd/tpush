package websocket

import (
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"sync"
	"time"
)

const (
	NotLoggedIn int64 = -1
)

type client struct {
	svc    *Service
	conn   *websocket.Conn
	id     int64
	writeq []*Response
	writeTimer *time.Timer
	mtx    *sync.Mutex
	cycle  time.Duration
	closed bool
}

func newClient(svc *Service, conn *websocket.Conn) *client {
	c := &client{
		svc:    svc,
		conn:   conn,
		id:     NotLoggedIn,
		writeq: make([]*Response, 0, 100),
		mtx:    &sync.Mutex{},
		cycle:  ClientCycle,
		closed: false,
	}
	return c
}

func (c *client) close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return
	}

	c.conn.Close()
	if c.id > 0 {
		c.svc.hb.unregister(c)
	}
	if c.writeTimer != nil {
		c.writeTimer.Reset(0)
	}
	c.closed = true
}

func (c *client) write(rsp *Response) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.writeq = append(c.writeq, rsp)
	c.writeTimer.Reset(c.cycle)
}

func (c *client) swap() (rsps []*Response, closed bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return nil, true
	}
	if len(c.writeq) == 0 {
		return nil, false
	}
	rsps = c.writeq
	c.writeq = make([]*Response, 0, 100)
	return rsps, false
}

func (c *client) readPump() {
	log.Debug("readPump start")
	defer func() {
		c.close()
		log.Debug("readPump complete")
	}()

	var reqs []*Request

	loginDone := make(chan struct{})
	go func() {
		defer log.Debug("waitLogin complete")
		select {
		case <-loginDone:
			return
		case <-time.After(LoginDeadline):
			log.Error("Client hasnot logged in for a long time")
			c.conn.Close()
			return
		}
	}()

	for {
		if err := c.conn.ReadJSON(&reqs); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 意外的错误
				log.Error(err)
			}
			return
		}

		for _, req := range reqs {
			if req == nil {
				log.Error("Nil request")
				return
			}
			rsp := &Response{
				Cmd: req.Cmd,
				Seq: req.Seq,
			}
			switch req.Cmd {
			case "login":
				if c.id > 0 {
					log.Error("Duplicate login")
					return
				}
				close(loginDone)
				if id, err := c.svc.loginHandler(req, rsp); err != nil {
					// 发生错误关闭连接
					log.Error(err)
					return
				} else {
					// 登陆成功
					c.id = id
					log.Debug("Client id:", c.id)
				}
			default:
				if c.id <= 0 {
					// 未登录
					log.Error("Hasnot logged in yet")
					return
				}

				if handler, ok := c.svc.mux[req.Cmd]; ok {
					if err := handler(req, rsp); err != nil {
						// 发生错误关闭连接
						log.Error(err)
						return
					}
				} else {
					// 未知命令
					log.Errorf("Unsupported command: %v", req.Cmd)
					return
				}
			}
			c.write(rsp)
		}
	}
}

func (c *client) writePump() {
	log.Debug("writePump start")
	c.writeTimer = time.NewTimer(c.cycle)

	defer func() {
		c.writeTimer.Stop()
		c.close()
		log.Debug("writePump complete")
	}()

	for {
		select {
		case <-c.writeTimer.C:
			rsps, closed := c.swap()
			if closed {
				return
			}

			if rsps == nil {
				log.Debug("Empty sendq")
				continue
			}

			log.Debug("Send Response:", rsps)
			if err := c.conn.WriteJSON(rsps); err != nil {
				return
			}
		}
	}
}
