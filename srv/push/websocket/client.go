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
	mtx    *sync.Mutex
	cycle  time.Duration
}

func newClient(svc *Service, conn *websocket.Conn) *client {
	c := &client{
		svc:    svc,
		conn:   conn,
		id:     NotLoggedIn,
		writeq: make([]*Response, 0, 100),
		mtx:    &sync.Mutex{},
		cycle:  ClientCycle,
	}
	return c
}

func (c *client) close() {
	c.conn.Close()
	if c.id > 0 {
		c.svc.hb.unregister(c)
	}
}

func (c *client) write(rsp *Response) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.writeq = append(c.writeq, rsp)
}

func (c *client) swap() []*Response {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	ret := c.writeq
	c.writeq = make([]*Response, 0, 100)
	return ret
}

func (c *client) readPump() {
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
			var rsp Response
			switch req.Cmd {
			case "login":
				close(loginDone)
				if id, err := c.svc.loginHandler(req, &rsp); err != nil {
					// 发生错误关闭连接
					log.Error(err)
					return
				} else {
					// 登陆成功
					c.id = id
				}
			default:
				if c.id <= 0 {
					// 未登录
					log.Error("Hasnot logged in yet")
					return
				}

				if handler, ok := c.svc.mux[req.Cmd]; ok {
					if err := handler(req, &rsp); err != nil {
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
			c.write(&rsp)
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(c.cycle)
	defer func() {
		ticker.Stop()
		c.close()
		log.Debug("writePump complete")
	}()

	for {
		select {
		case <-ticker.C:
			rsps := c.swap()
			if err := c.conn.WriteJSON(rsps); err != nil {
				return
			}
		}
	}
}
