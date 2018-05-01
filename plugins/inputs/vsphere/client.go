package vsphere

import (
	"context"
	"github.com/influxdata/telegraf/internal"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"log"
	"net/url"
)

type Client struct {
	Client *govmomi.Client
	Views  *view.Manager
	Root   *view.ContainerView
	Perf   *performance.Manager
}

func NewClient(url *url.URL, vs *VSphere) (*Client, error) {

	tlsCfg, err := internal.GetTLSConfig(vs.SSLCert, vs.SSLKey, vs.SSLCA, vs.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	var c *govmomi.Client
	if tlsCfg != nil && len(tlsCfg.Certificates) > 0 {
		//TODO: remove this log output before final release
		log.Printf("Creating client with Certificate: %s", url.Host)
		c, err = govmomi.NewClientWithCertificate(ctx, url, vs.InsecureSkipVerify, tlsCfg.Certificates[0])
	} else {
		//TODO: remove this log output before final release
		log.Printf("Creating client: %s", url.Host)
		c, err = govmomi.NewClient(ctx, url, vs.InsecureSkipVerify)
	}
	if err != nil {
		return nil, err
	}
	c.Timeout = vs.Timeout.Duration

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
