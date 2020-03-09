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
	reg        chan *clientData // 注册管道
	unreg      chan clientid    // 注销管道
	clients    sync.Map         //map[clientid]*clientData // 所有登陆的客户端
	channels   sync.Map         //map[channelid]*channel   // 频道
	newChannel *channel
}

func NewHub() *Hub {
	h := &Hub{
		reg:   make(chan *clientData, 100),
		unreg: make(chan clientid, 100),
		//clients:  make(map[clientid]*clientData),
		//channels: make(map[channelid]*channel),
	}
	return h
}

func (h *Hub) Register(id clientid, cli websocket.Client) {
	//h.reg <- &clientData{id, cli}
	h.clients.Store(id, &clientData{
		id:     id,
		client: cli,
	})
}

func (h *Hub) Unregister(id clientid) {
	//h.unreg <- id
	if cli_, ok := h.clients.Load(id); ok {
		h.clients.Delete(id)
		cli := cli_.(*clientData)
		_ = cli //cli.channels.Range()
	}
}

func (h *Hub) EnterChannel(id clientid, cid channelid) bool {
	if cli_, ok := h.clients.Load(id); ok {
		cli := cli_.(*clientData)
		if ch_, ok := h.channels.Load(cid); ok {
			ch := ch_.(*channel)
			ch.clients.Store(id, cli)
			return true
		} else {
			// create channel
			ch := &channel{
				id: cid,
			}
			h.channels.Store(cid, ch)
			return true
		}
	} else {
		return false
	}
}

func (h *Hub) ExitChannel(id clientid, cid channelid) {
	if ch_, ok := h.channels.Load(cid); ok {
		ch := ch_.(*channel)
		ch.clients.Delete(id)
	}
}

func (h *Hub) ChannelClients(cid channelid) (clis []websocket.Client) {
	if ch_, ok := h.channels.Load(cid); ok {
		ch := ch_.(*channel)

		ch.clients.Range(func(key, value interface{}) bool {
			cli := value.(*clientData)
			clis = append(clis, cli.client)
			return true
		})
	}
	return clis
}

func (h *Hub) Run() {
	for {
		select {
		//case cliData := <-h.reg:
		//h.clients[cliData.id] = cliData

		//case id := <-h.unreg:
		//delete(h.clients, id)
		}
	}
}
