package handler

import (
	"sync/atomic"
	"tpush/internal"
	"tpush/internal/websocket"
)

var (
	cliId int64 = 0
)

func genid() int64 {
	return atomic.AddInt64(&cliId, 1)
}

type Hub struct {
	clients *internal.BiMap  // id <-> Client
	where *internal.BIndex  // client <-> chan set
	who *internal.Index  // uid <-> client set
}

func (h *Hub) AddClient(cli websocket.Client) int64 {
	id := genid()
	h.clients.AddPair(id, cli)
	return id
}

func (h *Hub) Register(cli websocket.Client, uid int64) {
	h.who.AddUserTag(uid, cli)
}

func (h *Hub) ClientsOfUser(uid int64) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := h.who.Tags(uid, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (h *Hub) RemoveClient(cli websocket.Client) {
	h.clients.RemoveByValue(cli)
	h.where.RemoveUser(cli)
	h.who.RemoveTag(cli)
}

func (h *Hub) ClientEnterChannel(cli websocket.Client, chs ...string) {
	h.where.AddUserTag(cli, chs)
}

func (h *Hub) ClientExitChannel(cli websocket.Client, chs ...string) {
	h.where.RemoveUserTag(cli, chs)
}

func (h *Hub) ClientsInChannel(ch string) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := h.where.Users(ch, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (h *Hub) Client(id int64) (websocket.Client, bool) {
	if cli_, ok := h.clients.Value(id); ok {
		return cli_.(websocket.Client), true
	} else {
		return nil, false
	}
}

func (h *Hub) ClientId(cli websocket.Client) (int64, bool) {
	if id_, ok := h.clients.Key(cli); ok {
		return id_.(int64), true
	} else {
		return 0, false
	}
}

func (h *Hub) User(cli websocket.Client) (int64, bool) {
	if uid_, ok := h.who.User(cli); ok {
		return uid_.(int64), true
	} else {
		return 0, false
	}
}

func NewHub() *Hub {
	hub := &Hub{
		clients: internal.NewBiMap(),
		where: internal.NewBIndex(),
		who: internal.NewIndex(true),
	}
	return hub
}
