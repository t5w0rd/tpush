package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"reflect"
	"testing"
	"time"
	"tpush/internal/tchatroom"
)

func TestBIndex_RemoveUser(t *testing.T) {
	bi := tchatroom.NewBIndex()
	bi.AddUserTag("LJ", "NB")
	bi.AddUserTag("LJ", "SWITCH")
	bi.AddUserTag("CJL", "NB")
	bi.AddUserTag("CJL", "KOF")

	bi.RemoveTag("NB")

	var s []interface{}
	bi.Tags("LJ", &s)
	t.Log(s)

	bi.Users("KOF", &s)
	t.Log(s)

	bi.AllTags(&s)
	t.Log(s)

	bi.AllUsers(&s)
	t.Log(s)
}

func test1(p ...interface{}) {
	fmt.Println(reflect.TypeOf(p))
}

func test2(p ...interface{}) {
	test1(p...)
}

func TestFunc(t *testing.T) {
	test2(1, 2, 3, 4)
	b := []interface{}{1, 2, 3}
	test2(b...)
}

type Json struct {
	Cmd string
	Seq int64
}

func TestStringsBuilder(t *testing.T) {
	var b bytes.Buffer
	b.WriteByte('[')
	j := &Json{"login", 1002}
	json.NewEncoder(&b).Encode(j)
	b.WriteByte(',')
	j.Cmd = "enter"
	json.NewEncoder(&b).Encode(j)
	b.WriteByte(']')

	var b2 bytes.Buffer
	b2.ReadFrom(&b)
	t.Log(b2.String())
}

func TestNewBytes(t *testing.T) {
	d := "我abc"
	e := []rune(d)
	f := []byte(d)
	t.Log(e)
	t.Log(f)
}

func TestEtcd(t *testing.T) {
	cfg := clientv3.Config{
		Endpoints: []string{"etcd-cluster-client"},
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		panic(err)
	}
	kv := clientv3.NewKV(cli)
	kv.Put(nil, "", "", clientv3.WithPrevKV())
	time.Now()
}
