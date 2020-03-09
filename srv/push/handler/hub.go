package handler

import (
	"sync/atomic"
	"tpush/internal"
	"tpush/internal/websocket"
)

type clientid = int64
type channelid = string

var (
	cliId clientid = 0
)

func genid() clientid {
	return atomic.AddInt64(&cliId, 1)
}

type Hub struct {
	clients *internal.BiMap  // id <-> Client
	where *internal.BIndex  // clientData <-> channelid
}

func (h *Hub) AddClient(cli websocket.Client) int64 {
	id := genid()
	h.clients.AddPair(id, cli)
	return id
}

func (h *Hub) RemoveClient(cli websocket.Client) {
	h.clients.RemoveByValue(cli)
	h.where.RemoveUser(cli)
}

func (h *Hub) ClientEnterChannel(cli websocket.Client, ch channelid) {
	h.where.AddUserTag(cli, ch)
}

func (h *Hub) ClientExitChannel(cli websocket.Client, ch channelid) {
	h.where.RemoveUserTag(cli, ch)
}

func (h *Hub) Clients(ch channelid, output *[]websocket.Client) bool {
	var out []interface{}
	if ok := h.where.Users(ch, &out); !ok {
		return false
	}

	if size := len(out); cap(*output) < size {
		*output = make([]websocket.Client, size)
	} else {
		*output = (*output)[:size]
	}
	for i, o := range out {
		(*output)[i] = o.(websocket.Client)
	}
	return true
}

func NewHub() *Hub {
	hub := &Hub{
		clients: internal.NewBiMap(),
		where: internal.NewBIndex(),
	}
	return hub
}
