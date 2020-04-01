package wrapper

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	//log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
)

type route struct {
	client.Client
}

type SelectNodeKey struct{}

func (c *route) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	id := ctx.Value(SelectNodeKey{}).(string)
	nOpts := append(opts, client.WithSelectOption(
		// create a selector strategy
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			// flatten
			for i, service := range services {
				//log.Infof("req.Service: %s", req.Service())
				//log.Infof("service %d: %#v", i, service)
				for _, node := range service.Nodes {
					if node.Id == id {
						return func() (*registry.Node, error) {
							return node, nil
						}
					}
				}
			}

			return func() (*registry.Node, error) {
				return nil, selector.ErrNoneAvailable
			}
		}),
	))

	return c.Client.Call(ctx, req, rsp, nOpts...)
}

// NewClientWrapper is a wrapper which roundrobins requests
func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &route{
			Client: c,
		}
	}
}
