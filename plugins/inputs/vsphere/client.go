package vsphere

import (
	"context"
	"github.com/influxdata/telegraf/internal"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"net/url"
)

type Client struct {
	Client *govmomi.Client
	Views  *view.Manager
	Root   *view.ContainerView
	Perf   *performance.Manager
}

func NewClient(url *url.URL, timeout internal.Duration) (*Client, error) {
	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, url, true)
	if err != nil {
		return nil, err
	}
	c.Timeout = timeout.Duration
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{}, true)
	if err != nil {
		return nil, err
	}
	p := performance.NewManager(c.Client)

	return &Client{
		Client: c,
		Views:  m,
		Root:   v,
		Perf:   p,
	}, nil
}

func (c *Client) Close() {
	ctx := context.Background()
	if c.Views != nil {
		c.Views.Destroy(ctx)

	}
	if c.Client != nil {
		c.Client.Logout(ctx)
	}
}
