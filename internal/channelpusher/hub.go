package channelpusher

type clientid = int64
type channelid = string

type clientData struct {
	id     clientid
	client interface{}
}

type channel struct {
	name    string
	clients map[clientid]bool
}

func newChannel(name string) *channel {
	c := &channel{
		name:    name,
		clients: make(map[clientid]bool),
	}
	return c
}

type Hub struct {
	reg      chan *clientData         // 注册管道
	unreg    chan clientid            // 注销管道
	clients  map[clientid]*clientData // 所有登陆的客户端
	channels map[channelid]*channel   // 频道
}

func NewHub() *Hub {
	h := &Hub{
		reg:      make(chan *clientData, 100),
		unreg:    make(chan clientid, 100),
		clients:  make(map[clientid]*clientData),
		channels: make(map[channelid]*channel),
	}
	return h
}

func (h *Hub) Register(id clientid, cli interface{}) {
	h.reg <- &clientData{id, cli}
}

func (h *Hub) Unregister(id clientid) {
	h.unreg <- id
}

func (h *Hub) Run() {
	for {
		select {
		case cliData := <-h.reg:
			h.clients[cliData.id] = cliData

		case id := <-h.unreg:
			delete(h.clients, id)
		}
	}
}
