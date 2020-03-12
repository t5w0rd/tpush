package tchatroom

import (
	"sync/atomic"
	"tpush/internal/websocket"
)

var (
	cliId int64 = 0
)

func genid() int64 {
	id := atomic.AddInt64(&cliId, 1)
	if id == 0 {
		return atomic.AddInt64(&cliId, 1)
	}
	return id
}

type Room struct {
	clients *BiMap  // id <-> Client
	where   *BIndex // client <-> chan set
	who     *Index  // uid <-> client set
}

func (r *Room) AddClient(cli websocket.Client) int64 {
	id := genid()
	r.clients.AddPair(id, cli)
	return id
}

func (r *Room) Login(cli websocket.Client, uid int64) {
	r.who.AddUserTag(uid, cli)
}

func (r *Room) ClientsOfUser(uid int64) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := r.who.Tags(uid, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (r *Room) ClientsOfUsers(uids []int64) websocket.ClientGroup {
	uids_ := make([]interface{}, len(uids))
	for i, uid := range uids {
		uids_[i] = uid
	}
	var out []interface{}
	r.who.SelectTags(uids_, &out)
	return websocket.NewClientGroup(out)
}

func (r *Room) RemoveClient(cli websocket.Client) {
	r.clients.RemoveByValue(cli)
	r.where.RemoveUser(cli)
	r.who.RemoveTag(cli)
}

func (r *Room) ClientEnterChannel(cli websocket.Client, chs ...string) {
	chs_ := make([]interface{}, len(chs))
	for i, ch := range chs {
		chs_[i] = ch
	}
	r.where.AddUserTag(cli, chs_...)
}

func (r *Room) ClientExitChannel(cli websocket.Client, chs ...string) {
	chs_ := make([]interface{}, len(chs))
	for i, ch := range chs {
		chs_[i] = ch
	}
	r.where.RemoveUserTag(cli, chs_...)
}

func (r *Room) ClientsInChannel(ch string) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := r.where.Users(ch, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (r *Room) ClientsInChannels(chs []string) websocket.ClientGroup {
	chs_ := make([]interface{}, len(chs))
	for i, ch := range chs {
		chs_[i] = ch
	}
	var out []interface{}
	r.where.SelectUsers(chs_, &out)
	return websocket.NewClientGroup(out)
}

func (r *Room) Client(id int64) (websocket.Client, bool) {
	if cli_, ok := r.clients.Value(id); ok {
		return cli_.(websocket.Client), true
	} else {
		return nil, false
	}
}

func (r *Room) Clients(ids []int64) websocket.ClientGroup {
	ids_ := make([]interface{}, len(ids))
	for i, id := range ids {
		ids_[i] = id
	}
	clis_, oks := r.clients.Values(ids_)
	clis := make([]interface{}, 0, len(ids))
	for i, ok := range oks {
		if ok {
			clis = append(clis, clis_[i])
		}
	}
	return websocket.NewClientGroup(clis)
}

func (r *Room) ClientId(cli websocket.Client) (int64, bool) {
	if id_, ok := r.clients.Key(cli); ok {
		return id_.(int64), true
	} else {
		return 0, false
	}
}

func (r *Room) User(cli websocket.Client) (int64, bool) {
	if uid_, ok := r.who.User(cli); ok {
		return uid_.(int64), true
	} else {
		return 0, false
	}
}

func NewRoom() *Room {
	r := &Room{
		clients: NewBiMap(),
		where:   NewBIndex(),
		who:     NewIndex(true),
	}
	return r
}
