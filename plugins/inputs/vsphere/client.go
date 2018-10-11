package vsphere

import (
	"context"
	"crypto/tls"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
)

// ClientFactory is used to obtain Clients to be used throughout the plugin. Typically,
// a single Client is reused across all functions and goroutines, but the client
// is periodically recycled to avoid authentication expiration issues.
type ClientFactory struct {
	client *Client
	mux    sync.Mutex
	url    *url.URL
	parent *VSphere
}

// Client represents a connection to vSphere and is backed by a govmoni connection
type Client struct {
	Client    *govmomi.Client
	Views     *view.Manager
	Root      *view.ContainerView
	Perf      *performance.Manager
	Valid     bool
	Timeout   time.Duration
	closeGate sync.Once
}

// NewClientFactory creates a new ClientFactory and prepares it for use.
func NewClientFactory(ctx context.Context, url *url.URL, parent *VSphere) *ClientFactory {
	return &ClientFactory{
		client: nil,
		parent: parent,
		url:    url,
	}
}

// GetClient returns a client. The caller is responsible for calling Release()
// on the client once it's done using it.
func (cf *ClientFactory) GetClient(ctx context.Context) (*Client, error) {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	if cf.client == nil {
		var err error
		if cf.client, err = NewClient(ctx, cf.url, cf.parent); err != nil {
			return nil, err
		}
	}

	// Execute a dummy call against the server to make sure the client is
	// still functional. If not, try to log back in. If that doesn't work,
	// we give up.
	ctx1, cancel1 := context.WithTimeout(ctx, cf.parent.Timeout.Duration)
	defer cancel1()
	if _, err := methods.GetCurrentTime(ctx1, cf.client.Client); err != nil {
		log.Printf("I! [input.vsphere]: Client session seems to have time out. Reauthenticating!")
		ctx2, cancel2 := context.WithTimeout(ctx, cf.parent.Timeout.Duration)
		defer cancel2()
		if cf.client.Client.SessionManager.Login(ctx2, url.UserPassword(cf.parent.Username, cf.parent.Password)) != nil {
			return nil, err
		}
	}

	return cf.client, nil
}

// NewClient creates a new vSphere client based on the url and setting passed as parameters.
func NewClient(ctx context.Context, u *url.URL, vs *VSphere) (*Client, error) {
	sw := NewStopwatch("connect", u.Host)
	tlsCfg, err := vs.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	// Use a default TLS config if it's missing
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	if vs.Username != "" {
		u.User = url.UserPassword(vs.Username, vs.Password)
	}

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

	ctx1, cancel1 := context.WithTimeout(ctx, vs.Timeout.Duration)
	defer cancel1()
	vimClient, err := vim25.NewClient(ctx1, soapClient)
	if err != nil {
		return nil, err
	}
	sm := session.NewManager(vimClient)

	// If TSLKey is specified, try to log in as an extension using a cert.
	if vs.TLSKey != "" {
		ctx2, cancel2 := context.WithTimeout(ctx, vs.Timeout.Duration)
		defer cancel2()
		if err := sm.LoginExtensionByCertificate(ctx2, vs.TLSKey); err != nil {
			return nil, err
		}
	}

	// Create the govmomi client.
	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: sm,
	}

	// Only login if the URL contains user information.
	if u.User != nil {
		if err := c.Login(ctx, u.User); err != nil {
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
		Client:  c,
		Views:   m,
		Root:    v,
		Perf:    p,
		Valid:   true,
		Timeout: vs.Timeout.Duration,
	}, nil
}

// Close shuts down a ClientFactory and releases any resources associated with it.
func (cf *ClientFactory) Close() {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	if cf.client != nil {
		cf.client.close()
	}
}

func (c *Client) close() {

	// Use a Once to prevent us from panics stemming from trying
	// to close it multiple times.
	c.closeGate.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()
		if c.Client != nil {
			if err := c.Client.Logout(ctx); err != nil {
				log.Printf("E! [input.vsphere]: Error during logout: %s", err)
			}
		}
	})
}

// GetServerTime returns the time at the vCenter server
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	t, err := methods.GetCurrentTime(ctx, c.Client)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}
