package plugins

import (
	log "github.com/micro/go-micro/v2/logger"
	"sync"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"

	"context"
)

type route struct {
	sync.Mutex
	rr map[string]int
	client.Client
}

func (c *route) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	nOpts := append(opts, client.WithSelectOption(
		// create a selector strategy
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			// flatten
			var nodes []*registry.Node
			for i, service := range services {
				log.Infof("req.Service: %s", req.Service())
				log.Infof("service %d: %+v", i, service)
				for j, node := range service.Nodes {
					log.Infof("node %d: %+v", j, node)
				}
				nodes = append(nodes, service.Nodes...)
			}

			// create the next func that always returns our node
			return func() (*registry.Node, error) {
				if len(nodes) == 0 {
					return nil, selector.ErrNoneAvailable
				}
				c.Lock()
				// get counter
				rr := c.rr[req.Service()]
				// get node
				node := nodes[rr%len(nodes)]
				// increment
				rr++
				// save
				c.rr[req.Service()] = rr
				c.Unlock()

				return node, nil
			}
		}),
	))

	return c.Client.Call(ctx, req, rsp, nOpts...)
}

// NewClientWrapper is a wrapper which roundrobins requests
func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &route{
			rr:     make(map[string]int),
			Client: c,
		}
	}
}
