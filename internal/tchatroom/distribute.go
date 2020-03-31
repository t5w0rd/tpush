package tchatroom

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

const (
	RegClientKeyFmt  = "/ids/%d"
	RegUserKeyFmt    = "/uids/%d"
	RegChannelKeyFmt = "/chans/%s"

	etcdClientTimeout = time.Millisecond * 200
	refreshTtlPeriod  = time.Second * 10
	regChanBufSize    = 1000
)

type Distribute interface {
	Register(key string)
	Unregister(key string)
	Run() (stopFunc func())
}

type etcd struct {
	nodeName string
	store    *clientv3.Client
	ttl      time.Duration

	regCh   chan string
	unregCh chan string
}

func (d *etcd) Register(key string) {
	d.regCh <- key
}

func (d *etcd) Unregister(key string) {
	d.unregCh <- key
}

func (d *etcd) Run() (stopFunc func()) {
	stopCh := make(chan struct{})

	stopFunc = func() {
		close(stopCh)
	}

	go func() {
		ttl := int64(d.ttl / time.Second)
		registry := make(map[string]clientv3.LeaseID)

		t := time.NewTicker(refreshTtlPeriod)

		for {
			select {
			case <-stopCh:
				return

			case key := <-d.regCh:
				ctx, _ := context.WithTimeout(context.Background(), etcdClientTimeout)
				leaseRsp, err := d.store.Grant(ctx, ttl)
				if err != nil {
					break
				}
				registry[key] = leaseRsp.ID

				ctx, _ = context.WithTimeout(context.Background(), etcdClientTimeout)
				k := fmt.Sprintf("%s/%s", key, d.nodeName)
				_, err = d.store.Put(ctx, k, d.nodeName, clientv3.WithLease(leaseRsp.ID))
				if err != nil {
					break
				}

			case key := <-d.unregCh:
				leaseID, ok := registry[key]
				if !ok {
					break
				}
				delete(registry, key)

				ctx, _ := context.WithTimeout(context.Background(), etcdClientTimeout)
				_, err := d.store.Revoke(ctx, leaseID)
				if err != nil {
					break
				}

			case <-t.C:
				for _, leaseID := range registry {
					ctx, _ := context.WithTimeout(context.Background(), etcdClientTimeout)
					_, err := d.store.KeepAliveOnce(ctx, leaseID)
					if err != nil {
					}
				}
			}
		}
	}()
	return stopFunc
}

func NewEtcdDistribute(nodeName string, store *clientv3.Client, ttl time.Duration) Distribute {
	if ttl <= refreshTtlPeriod {
		panic("ttl is too small")
	}
	d := &etcd{
		nodeName: nodeName,
		store:    store,
		ttl:      ttl,

		regCh:   make(chan string, regChanBufSize),
		unregCh: make(chan string, regChanBufSize),
	}
	return d
}
