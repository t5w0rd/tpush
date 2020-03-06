package websocket

type clientid = int64
type channelid = string

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

type hub struct {
	reg      chan *client           // 注册管道
	unreg    chan *client           // 注销管道
	clients  map[clientid]*client   // 所有登陆的客户端
	channels map[channelid]*channel // 频道
}

func newHub() *hub {
	h := &hub{
		reg:      make(chan *client, 100),
		unreg:    make(chan *client, 100),
		clients:  make(map[clientid]*client),
		channels: make(map[channelid]*channel),
	}
	return h
}

func (h *hub) register(cli *client) {
	h.reg <- cli
}

func (h *hub) unregister(cli *client) {
	h.unreg <- cli
}

func (h *hub) run() {
	for {
		select {
		case cli := <-h.reg:
			h.clients[cli.id] = cli
		case cli := <-h.unreg:
			delete(h.clients, cli.id)
		}
	}
}
