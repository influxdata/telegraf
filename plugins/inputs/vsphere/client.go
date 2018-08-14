package vsphere

import (
	"context"
	"log"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

// Client represents a connection to vSphere and is backed by a govmoni connection
type Client struct {
	Client *govmomi.Client
	Views  *view.Manager
	Root   *view.ContainerView
	Perf   *performance.Manager
	Valid  bool
}

// NewClient creates a new vSphere client based on the url and setting passed as parameters.
func NewClient(u *url.URL, vs *VSphere) (*Client, error) {
	sw := NewStopwatch("connect", u.Host)
	tlsCfg, err := vs.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	if vs.Username != "" {
		u.User = url.UserPassword(vs.Username, vs.Password)
	}
	ctx := context.Background()

	log.Printf("D! [input.vsphere]: Creating client: %s", u.Host)
	soapClient := soap.NewClient(u, tlsCfg.InsecureSkipVerify)

	// Add certificate if we have it. Use it to log us in.
	if tlsCfg != nil && len(tlsCfg.Certificates) > 0 {
		soapClient.SetCertificate(tlsCfg.Certificates[0])
	}

	// Set up custom CA chain if specified.  We need to do this before we create the vim25 client,
	// since it might fail on missing CA chains otherwise.
	if vs.TLSCA != "" {
		if err := soapClient.SetRootCAs(vs.TLSCA); err != nil {
			return nil, err
		}
	}

	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}
	sm := session.NewManager(vimClient)

	// If TSLKey is specified, try to log in as an extension using a cert.
	if vs.TLSKey != "" {
		sm.LoginExtensionByCertificate(ctx, vs.TLSKey)
	}

	// Create the govmomi client.
	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: sm,
	}

	// Only login if the URL contains user information.
	if u.User != nil {
		err = c.Login(ctx, u.User)
		if err != nil {
			return nil, err
		}
	}

	c.Timeout = vs.Timeout.Duration
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{}, true)
	if err != nil {
		return nil, err
	}

	p := performance.NewManager(c.Client)

	sw.Stop()

	return &Client{
		Client: c,
		Views:  m,
		Root:   v,
		Perf:   p,
		Valid:  true,
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
