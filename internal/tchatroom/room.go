package tchatroom

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

type room struct {
	clients *internal.BiMap  // id <-> Client
	where *internal.BIndex  // client <-> chan set
	who *internal.Index  // uid <-> client set
}

func (r *room) AddClient(cli websocket.Client) int64 {
	id := genid()
	r.clients.AddPair(id, cli)
	return id
}

func (r *room) Login(cli websocket.Client, uid int64) {
	r.who.AddUserTag(uid, cli)
}

func (r *room) ClientsOfUser(uid int64) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := r.who.Tags(uid, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (r *room) RemoveClient(cli websocket.Client) {
	r.clients.RemoveByValue(cli)
	r.where.RemoveUser(cli)
	r.who.RemoveTag(cli)
}

func (r *room) ClientEnterChannel(cli websocket.Client, chs ...string) {
	chs_ := make([]interface{}, len(chs))
	for i, ch := range chs {
		chs_[i] = ch
	}
	r.where.AddUserTag(cli, chs_...)
}

func (r *room) ClientExitChannel(cli websocket.Client, chs ...string) {
	chs_ := make([]interface{}, len(chs))
	for i, ch := range chs {
		chs_[i] = ch
	}
	r.where.RemoveUserTag(cli, chs_...)
}

func (r *room) ClientsInChannel(ch string) (websocket.ClientGroup, bool) {
	var out []interface{}
	if ok := r.where.Users(ch, &out); !ok {
		return nil, false
	}
	return websocket.NewClientGroup(out), true
}

func (r *room) Client(id int64) (websocket.Client, bool) {
	if cli_, ok := r.clients.Value(id); ok {
		return cli_.(websocket.Client), true
	} else {
		return nil, false
	}
}

func (r *room) ClientId(cli websocket.Client) (int64, bool) {
	if id_, ok := r.clients.Key(cli); ok {
		return id_.(int64), true
	} else {
		return 0, false
	}
}

func (r *room) User(cli websocket.Client) (int64, bool) {
	if uid_, ok := r.who.User(cli); ok {
		return uid_.(int64), true
	} else {
		return 0, false
	}
}

func newRoom() *room {
	r := &room{
		clients: internal.NewBiMap(),
		where: internal.NewBIndex(),
		who: internal.NewIndex(true),
	}
	return r
}
