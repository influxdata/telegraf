package vsphere

import (
	"context"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

// ClientFactory is used to obtain Clients to be used throughout the plugin. Typically,
// a single Client is reused across all functions and goroutines, but the client
// is periodically recycled to avoid authentication expiration issues.
type ClientFactory struct {
	client   *Client
	mux      sync.Mutex
	url      *url.URL
	parent   *VSphere
	recycler *time.Ticker
}

// Client represents a connection to vSphere and is backed by a govmoni connection
type Client struct {
	Client    *govmomi.Client
	Views     *view.Manager
	Root      *view.ContainerView
	Perf      *performance.Manager
	Valid     bool
	refcount  int32
	mux       sync.Mutex
	idle      *sync.Cond
	closeGate sync.Once
}

// NewClientFactory creates a new ClientFactory and prepares it for use.
func NewClientFactory(ctx context.Context, url *url.URL, parent *VSphere) *ClientFactory {
	cf := &ClientFactory{
		client:   nil,
		parent:   parent,
		url:      url,
		recycler: time.NewTicker(30 * time.Minute),
	}

	// Perdiodically recycle clients to make sure they don't expire
	go func() {
		for {
			select {
			case <-cf.recycler.C:
				cf.destroyCurrent()
			case <-ctx.Done():
				cf.destroyCurrent() // Kill the current connection when we're done.
				return
			}
		}
	}()
	return cf
}

// GetClient returns a client. The caller is responsible for calling Release()
// on the client once it's done using it.
func (cf *ClientFactory) GetClient() (*Client, error) {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	if cf.client == nil {
		var err error
		if cf.client, err = NewClient(cf.url, cf.parent); err != nil {
			return nil, err
		}
	}
	cf.client.grab()
	return cf.client, nil
}

func (cf *ClientFactory) destroyCurrent() {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	go func(c *Client) {
		c.closeWhenIdle()
	}(cf.client)
	cf.client = nil
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

	result := &Client{
		Client:   c,
		Views:    m,
		Root:     v,
		Perf:     p,
		Valid:    true,
		refcount: 0,
	}
	result.idle = sync.NewCond(&result.mux)
	return result, nil
}

func (c *Client) close() {

	// Use a Once to prevent us from panics stemming from trying
	// to close it multiple times.
	c.closeGate.Do(func() {
		ctx := context.Background()
		if c.Views != nil {
			c.Views.Destroy(ctx)

		}
		if c.Client != nil {
			c.Client.Logout(ctx)
		}
	})
}

func (c *Client) closeWhenIdle() {
	c.mux.Lock()
	defer c.mux.Unlock()
	log.Printf("D! [input.vsphere]: Waiting to close connection")
	for c.refcount > 0 {
		c.idle.Wait()
	}
	log.Printf("D! [input.vsphere]: Closing connection")
	c.close()
}

// Release indicates that a caller is no longer using the client and it can
// be recycled if needed.
func (c *Client) Release() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.refcount--; c.refcount == 0 {
		c.idle.Broadcast()
	}
	//log.Printf("D! [input.vsphere]: Release. Connection refcount:%d", c.refcount)
}

func (c *Client) grab() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.refcount++
	//log.Printf("D! [input.vsphere]: Grab. Connection refcount:%d", c.refcount)
}
