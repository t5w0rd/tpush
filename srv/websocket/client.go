package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func (c *Client) readPump() {

}

func (c *Client) writePump() {

}
