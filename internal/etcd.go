package internal

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	log "github.com/micro/go-micro/v2/logger"
	"time"
)

func GetDistributeNodes(cli *clientv3.Client, keys []string, timeout time.Duration) map[string]string {
	size := len(keys)
	dones := make([]chan struct{}, size)
	for i, _ := range dones {
		dones[i] = make(chan struct{})
	}

	ctxs := make([]context.Context, size)
	cancels := make([]context.CancelFunc, size)
	for i := 0; i < size; i++ {
		ctxs[i], cancels[i] = context.WithTimeout(context.Background(), timeout)
	}

	// multi requests
	ch := make(chan *mvccpb.KeyValue, size)
	for i := 0; i < size; i++ {
		go func(i int) {
			defer close(dones[i])
			log.Infof("PrefixKey: %#v", keys[i])
			getRsp, err := cli.Get(ctxs[i], keys[i], clientv3.WithPrefix())
			if err != nil {
				log.Error(err)
				return
			}

			log.Infof("kvs: %#v", getRsp.Kvs)
			for _, kv := range getRsp.Kvs {
				log.Infof("Key: %s, Value: %s", kv.Key, kv.Value)
				ch <- kv
			}
		}(i)
	}

	go func() {
		// wait all
		for _, done := range dones {
			<-done
		}
		close(ch)
	}()

	// map writer
	ret := make(map[string]string)
	for {
		kv, ok := <-ch
		if !ok {
			break
		}
		ret[string(kv.Value)] = string(kv.Key)
	}

	return ret
}
