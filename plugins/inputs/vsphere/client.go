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

// ClientFactory is used to obtain Clients to be used throughout the plugin. Typically,
// a single Client is reused across all functions and goroutines, but the client
// is periodically recycled to avoid authentication expiration issues.
type ClientFactory struct {
	client     *Client
	mux        sync.Mutex
	vSphereURL *url.URL
	parent     *VSphere
}

// Client represents a connection to vSphere and is backed by a govmomi connection
type Client struct {
	Client    *govmomi.Client
	Views     *view.Manager
	Root      *view.ContainerView
	Perf      *performance.Manager
	Valid     bool
	Timeout   time.Duration
	closeGate sync.Once
	log       telegraf.Logger
}

// NewClientFactory creates a new ClientFactory and prepares it for use.
func NewClientFactory(vSphereURL *url.URL, parent *VSphere) *ClientFactory {
	return &ClientFactory{
		client:     nil,
		parent:     parent,
		vSphereURL: vSphereURL,
	}
}

// GetClient returns a client. The caller is responsible for calling Release()
// on the client once it's done using it.
func (cf *ClientFactory) GetClient(ctx context.Context) (*Client, error) {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	retrying := false
	for {
		if cf.client == nil {
			var err error
			if cf.client, err = NewClient(ctx, cf.vSphereURL, cf.parent); err != nil {
				return nil, err
			}
		}

		// Execute a dummy call against the server to make sure the client is
		// still functional. If not, try to log back in. If that doesn't work,
		// we give up.
		ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(cf.parent.Timeout))
		defer cancel1()
		if _, err := methods.GetCurrentTime(ctx1, cf.client.Client); err != nil {
			cf.parent.Log.Info("Client session seems to have time out. Reauthenticating!")
			ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(cf.parent.Timeout))
			defer cancel2()
			if err := cf.client.Client.SessionManager.Login(ctx2, url.UserPassword(cf.parent.Username, cf.parent.Password)); err != nil {
				if !retrying {
					// The client went stale. Probably because someone rebooted vCenter. Clear it to
					// force us to create a fresh one. We only get one chance at this. If we fail a second time
					// we will simply skip this collection round and hope things have stabilized for the next one.
					retrying = true
					cf.client = nil
					continue
				}
				return nil, fmt.Errorf("renewing authentication failed: %s", err.Error())
			}
		}

		return cf.client, nil
	}
}

// NewClient creates a new vSphere client based on the url and setting passed as parameters.
func NewClient(ctx context.Context, vSphereURL *url.URL, vs *VSphere) (*Client, error) {
	sw := NewStopwatch("connect", vSphereURL.Host)
	defer sw.Stop()

	tlsCfg, err := vs.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	// Use a default TLS config if it's missing
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	if vs.Username != "" {
		vSphereURL.User = url.UserPassword(vs.Username, vs.Password)
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

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{}, true)
	if err != nil {
		return nil, err
	}

	p := performance.NewManager(c.Client)

	client := &Client{
		log:     vs.Log,
		Client:  c,
		Views:   m,
		Root:    v,
		Perf:    p,
		Valid:   true,
		Timeout: time.Duration(vs.Timeout),
	}
	// Adjust max query size if needed
	ctx3, cancel3 := context.WithTimeout(ctx, time.Duration(vs.Timeout))
	defer cancel3()
	n, err := client.GetMaxQueryMetrics(ctx3)
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
				c.log.Errorf("Logout: %s", err.Error())
			}
		}
	})
}

// GetServerTime returns the time at the vCenter server
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	t, err := methods.GetCurrentTime(ctx, c.Client)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}

// GetMaxQueryMetrics returns the max_query_metrics setting as configured in vCenter
func (c *Client) GetMaxQueryMetrics(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	om := object.NewOptionManager(c.Client.Client, *c.Client.Client.ServiceContent.Setting)
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
	ver := c.Client.Client.ServiceContent.About.Version
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

// QueryMetrics wraps performance.Query to give it proper timeouts
func (c *Client) QueryMetrics(ctx context.Context, pqs []types.PerfQuerySpec) ([]performance.EntityMetric, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	metrics, err := c.Perf.Query(ctx1, pqs)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(ctx, c.Timeout)
	defer cancel2()
	return c.Perf.ToMetricSeries(ctx2, metrics)
}

// CounterInfoByName wraps performance.CounterInfoByName to give it proper timeouts
func (c *Client) CounterInfoByName(ctx context.Context) (map[string]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return c.Perf.CounterInfoByName(ctx1)
}

// CounterInfoByKey wraps performance.CounterInfoByKey to give it proper timeouts
func (c *Client) CounterInfoByKey(ctx context.Context) (map[int32]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return c.Perf.CounterInfoByKey(ctx1)
}

// ListResources wraps property.Collector.Retrieve to give it proper timeouts
func (c *Client) ListResources(ctx context.Context, root *view.ContainerView, kind []string, ps []string, dst interface{}) error {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return root.Retrieve(ctx1, kind, ps, dst)
}

func (c *Client) GetCustomFields(ctx context.Context) (map[int32]string, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	cfm := object.NewCustomFieldsManager(c.Client.Client)
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
