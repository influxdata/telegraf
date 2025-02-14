package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/influxdata/telegraf"
)

// The highest number of metrics we can query for, no matter what settings
// and server say.
const absoluteMaxMetrics = 10000

// clientFactory is used to obtain Clients to be used throughout the plugin. Typically,
// a single client is reused across all functions and goroutines, but the client
// is periodically recycled to avoid authentication expiration issues.
type clientFactory struct {
	client     *client
	mux        sync.Mutex
	vSphereURL *url.URL
	parent     *VSphere
}

// client represents a connection to vSphere and is backed by a govmomi connection
type client struct {
	client    *govmomi.Client
	views     *view.Manager
	root      *view.ContainerView
	perf      *performance.Manager
	valid     bool
	timeout   time.Duration
	closeGate sync.Once
	log       telegraf.Logger
}

// newClientFactory creates a new clientFactory and prepares it for use.
func newClientFactory(vSphereURL *url.URL, parent *VSphere) *clientFactory {
	return &clientFactory{
		client:     nil,
		parent:     parent,
		vSphereURL: vSphereURL,
	}
}

// getClient returns a client. The caller is responsible for calling Release()
// on the client once it's done using it.
func (cf *clientFactory) getClient(ctx context.Context) (*client, error) {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	retrying := false
	for {
		if cf.client == nil {
			var err error
			if cf.client, err = newClient(ctx, cf.vSphereURL, cf.parent); err != nil {
				return nil, err
			}
		}

		err := cf.testClient(ctx)
		if err != nil {
			if !retrying {
				// The client went stale. Probably because someone rebooted vCenter. Clear it to
				// force us to create a fresh one. We only get one chance at this. If we fail a second time
				// we will simply skip this collection round and hope things have stabilized for the next one.
				retrying = true
				cf.client = nil
				continue
			}
			return nil, err
		}

		return cf.client, nil
	}
}

func (cf *clientFactory) testClient(ctx context.Context) error {
	// Execute a dummy call against the server to make sure the client is
	// still functional. If not, try to log back in. If that doesn't work,
	// we give up.
	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(cf.parent.Timeout))
	defer cancel1()
	if _, err := methods.GetCurrentTime(ctx1, cf.client.client); err != nil {
		cf.parent.Log.Info("Client session seems to have time out. Reauthenticating!")
		ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(cf.parent.Timeout))
		defer cancel2()

		// Resolving the secrets and construct the authentication info
		username, err := cf.parent.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		defer username.Destroy()
		password, err := cf.parent.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		defer password.Destroy()
		auth := url.UserPassword(username.String(), password.String())

		if err := cf.client.client.SessionManager.Login(ctx2, auth); err != nil {
			return fmt.Errorf("renewing authentication failed: %w", err)
		}
	}

	return nil
}

// newClient creates a new vSphere client based on the url and setting passed as parameters.
func newClient(ctx context.Context, vSphereURL *url.URL, vs *VSphere) (*client, error) {
	sw := newStopwatch("connect", vSphereURL.Host)
	defer sw.stop()

	tlsCfg, err := vs.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	// Use a default TLS config if it's missing
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	if !vs.Username.Empty() {
		// Resolving the secrets and construct the authentication info
		username, err := vs.Username.Get()
		if err != nil {
			return nil, fmt.Errorf("getting username failed: %w", err)
		}
		password, err := vs.Password.Get()
		if err != nil {
			username.Destroy()
			return nil, fmt.Errorf("getting password failed: %w", err)
		}
		vSphereURL.User = url.UserPassword(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}

	vs.Log.Debugf("Creating client: %s", vSphereURL.Host)
	soapClient := soap.NewClient(vSphereURL, tlsCfg.InsecureSkipVerify)

	// Add certificate if we have it. Use it to log us in.
	if len(tlsCfg.Certificates) > 0 {
		soapClient.SetCertificate(tlsCfg.Certificates[0])
	}

	// Set up custom CA chain if specified.  We need to do this before we create the vim25 client,
	// since it might fail on missing CA chains otherwise.
	if vs.TLSCA != "" {
		if err := soapClient.SetRootCAs(vs.TLSCA); err != nil {
			return nil, err
		}
	}

	// Set the proxy dependent on the settings
	proxy, err := vs.HTTPProxy.Proxy()
	if err != nil {
		return nil, fmt.Errorf("creating proxy failed: %w", err)
	}
	transport := soapClient.DefaultTransport()
	transport.Proxy = proxy
	soapClient.Client.Transport = transport

	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(vs.Timeout))
	defer cancel1()
	vimClient, err := vim25.NewClient(ctx1, soapClient)
	if err != nil {
		return nil, err
	}
	sm := session.NewManager(vimClient)

	// If TSLKey is specified, try to log in as an extension using a cert.
	if vs.TLSKey != "" {
		ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(vs.Timeout))
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
	if vSphereURL.User != nil {
		if err := c.Login(ctx, vSphereURL.User); err != nil {
			return nil, err
		}
	}

	c.Timeout = time.Duration(vs.Timeout)
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, make([]string, 0), true)
	if err != nil {
		return nil, err
	}

	p := performance.NewManager(c.Client)

	client := &client{
		log:     vs.Log,
		client:  c,
		views:   m,
		root:    v,
		perf:    p,
		valid:   true,
		timeout: time.Duration(vs.Timeout),
	}
	// Adjust max query size if needed
	ctx3, cancel3 := context.WithTimeout(ctx, time.Duration(vs.Timeout))
	defer cancel3()
	n, err := client.getMaxQueryMetrics(ctx3)
	if err != nil {
		return nil, err
	}
	vs.Log.Debugf("vCenter says max_query_metrics should be %d", n)
	if n < vs.MaxQueryMetrics {
		vs.Log.Warnf("Configured max_query_metrics is %d, but server limits it to %d. Reducing.", vs.MaxQueryMetrics, n)
		vs.MaxQueryMetrics = n
	}
	return client, nil
}

// close shuts down a clientFactory and releases any resources associated with it.
func (cf *clientFactory) close() {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	if cf.client != nil {
		cf.client.close()
	}
}

func (c *client) close() {
	// Use a Once to prevent us from panics stemming from trying
	// to close it multiple times.
	c.closeGate.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		if c.client != nil {
			if err := c.client.Logout(ctx); err != nil {
				c.log.Errorf("Logout: %s", err.Error())
			}
		}
	})
}

// getServerTime returns the time at the vCenter server
func (c *client) getServerTime(ctx context.Context) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	t, err := methods.GetCurrentTime(ctx, c.client)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}

// getMaxQueryMetrics returns the max_query_metrics setting as configured in vCenter
func (c *client) getMaxQueryMetrics(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	om := object.NewOptionManager(c.client.Client, *c.client.Client.ServiceContent.Setting)
	res, err := om.Query(ctx, "config.vpxd.stats.maxQueryMetrics")
	if err == nil {
		if len(res) > 0 {
			if s, ok := res[0].GetOptionValue().Value.(string); ok {
				v, err := strconv.Atoi(s)
				if err == nil {
					c.log.Debugf("vCenter maxQueryMetrics is defined: %d", v)
					if v == -1 {
						// Whatever the server says, we never ask for more metrics than this.
						return absoluteMaxMetrics, nil
					}
					return v, nil
				}
			}
			// Fall through version-based inference if value isn't usable
		}
	} else {
		c.log.Debug("Option query for maxQueryMetrics failed. Using default")
	}

	// No usable maxQueryMetrics setting. Infer based on version
	ver := c.client.Client.ServiceContent.About.Version
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		c.log.Warnf("vCenter returned an invalid version string: %s. Using default query size=64", ver)
		return 64, nil
	}
	c.log.Debugf("vCenter version is: %s", ver)
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	if major < 6 || major == 6 && parts[1] == "0" {
		return 64, nil
	}
	return 256, nil
}

// queryMetrics wraps performance.Query to give it proper timeouts
func (c *client) queryMetrics(ctx context.Context, pqs []types.PerfQuerySpec) ([]performance.EntityMetric, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.timeout)
	defer cancel1()
	metrics, err := c.perf.Query(ctx1, pqs)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(ctx, c.timeout)
	defer cancel2()
	return c.perf.ToMetricSeries(ctx2, metrics)
}

// counterInfoByName wraps performance.counterInfoByName to give it proper timeouts
func (c *client) counterInfoByName(ctx context.Context) (map[string]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.timeout)
	defer cancel1()
	return c.perf.CounterInfoByName(ctx1)
}

// counterInfoByKey wraps performance.counterInfoByKey to give it proper timeouts
func (c *client) counterInfoByKey(ctx context.Context) (map[int32]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.timeout)
	defer cancel1()
	return c.perf.CounterInfoByKey(ctx1)
}

func (c *client) getCustomFields(ctx context.Context) (map[int32]string, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.timeout)
	defer cancel1()
	cfm := object.NewCustomFieldsManager(c.client.Client)
	fields, err := cfm.Field(ctx1)
	if err != nil {
		return nil, err
	}
	r := make(map[int32]string)
	for _, f := range fields {
		r[f.Key] = f.Name
	}
	return r, nil
}
