package handler

import (
	"sync"
	"tpush/internal/websocket"
)

type clientid = int64
type channelid = string

type clientData struct {
	client   websocket.Client
	id       clientid
	channels sync.Map //map[channelid]bool
}

type channel struct {
	id      channelid
	clients sync.Map //map[clientid]bool
}

type Hub struct {

}
