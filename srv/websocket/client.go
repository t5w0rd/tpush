package websocket

import (
	"github.com/gorilla/websocket"
	log "github.com/micro/go-micro/v2/logger"
	"time"
)

type client struct {
	conn *websocket.Conn
	id int64
	send chan *Response
	cycle time.Duration
}

func newClient(conn *websocket.Conn, id int64) *client {
	c := &client{
		conn: conn,
		id: id,
	}
	return c
}

func (c *client) readPump() {
	var req Response
	for {
		if err := c.conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 意外的错误
				log.Error(err)
			}
			break
		}
		switch req.Cmd {
		default:
			log.Errorf("Unsupported command: %v", req.Cmd)
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(c.cycle)
	for {
		select {
		case <- ticker.C:
			rsps := make([]*Response, 0, len(c.send) * 2)
			for rsp := range c.send {
				rsps = append(rsps, rsp)
			}

			if err := c.conn.WriteJSON(rsps); err != nil {
				break
			}
		}
	}
}
