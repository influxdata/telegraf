package vsphere

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
)

// Client represents a connection to vSphere and is backed by a govmoni connection
type Client struct {
	Client *govmomi.Client
	Views  *view.Manager
	Root   *view.ContainerView
	Perf   *performance.Manager
}

// NewClient creates a new vSphere client based on the url and setting passed as parameters.
func NewClient(u *url.URL, vs *VSphere) (*Client, error) {
	tlsCfg, err := vs.TLSConfig()
	if err != nil {
		return nil, err
	}

	if vs.Username != "" {
		if vs.Password == "" {
			return nil, fmt.Errorf("vSphere password must be specified")
		}
		log.Printf("D! Logging in using explicit credentials: %s", vs.Username)
		u.User = url.UserPassword(vs.Username, vs.Password)
	}
	ctx := context.Background()
	var c *govmomi.Client
	if tlsCfg != nil && len(tlsCfg.Certificates) > 0 {
		log.Printf("D! Creating client with Certificate: %s", u.Host)
		c, err = govmomi.NewClientWithCertificate(ctx, u, vs.InsecureSkipVerify, tlsCfg.Certificates[0])
	} else {
		log.Printf("D! Creating client: %s", u.Host)
		c, err = govmomi.NewClient(ctx, u, vs.InsecureSkipVerify)
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

// Close disconnects a client from the vSphere backend and releases all assiciated resources.
func (c *Client) Close() {
	ctx := context.Background()
	if c.Views != nil {
		c.Views.Destroy(ctx)

	}
	if c.Client != nil {
		c.Client.Logout(ctx)
	}
}
