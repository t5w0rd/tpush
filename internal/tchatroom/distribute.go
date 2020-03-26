package tchatroom

import (
	"github.com/go-redis/redis/v7"
	"time"
)

const (
	refreshTtlPeriod = time.Second * 10
	regChanBufSize   = 1000
)

type Distribute interface {
	Register(key string)
	Unregister(key string)
}

type distribute struct {
	nodeName string
	store    *redis.Client
	ttl      time.Duration

	regCh   chan string
	unregCh chan string
}

func (d *distribute) Register(key string) {
	d.regCh <- key
}

func (d *distribute) Unregister(key string) {
	d.unregCh <- key
}

func (d *distribute) Run() (stopFunc func()) {
	stopCh := make(chan struct{})

	stopFunc = func() {
		close(stopCh)
	}

	go func() {
		registry := make(map[string]struct{})
		t := time.NewTicker(refreshTtlPeriod)

		for {
			select {
			case <-stopCh:
				return

			case key := <-d.regCh:
				d.store.Set(key, d.nodeName, d.ttl)

			case key := <-d.unregCh:
				d.store.Del(key)

			case <-t.C:
				for key, _ := range registry {
					d.store.Set(key, d.nodeName, d.ttl)
				}
			}
		}
	}()
	return stopFunc
}

func NewRedisDistribute(nodeName string, store *redis.Client, ttl time.Duration) *distribute {
	if ttl <= time.Second {
		panic("ttl is too small")
	}
	d := &distribute{
		nodeName: nodeName,
		store:    store,
		ttl:      ttl,

		regCh:   make(chan string, regChanBufSize),
		unregCh: make(chan string, regChanBufSize),
	}
	return d
}
