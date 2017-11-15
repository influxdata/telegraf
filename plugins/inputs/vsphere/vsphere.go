package vsphere

import (
	"context"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"sync"
	"time"
)

type objectMap map[string]objectRef

type Endpoint struct {
	Parent       *VSphere
	Url          *url.URL
	intervals    []int32
	lastColl     map[string]time.Time
	hostMap      objectMap
	vmMap        objectMap
	clusterMap   objectMap
	nameCache    map[string]string
	bgObjectDisc bool
	collectMux   sync.RWMutex
}

type VSphere struct {
	Vcenters                []string
	VmInterval              internal.Duration
	HostInterval            internal.Duration
	ClusterInterval         internal.Duration
	DatastoreInterval       internal.Duration
	ObjectDiscoveryInterval internal.Duration
	MaxSamples              int32
	MaxQuery                int32
	endpoints               []Endpoint
}

type objectRef struct {
	name      string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
}

type InstanceMetrics map[string]map[string]interface{}

func NewEndpoint(parent *VSphere, url *url.URL) Endpoint {
	hostMap := make(objectMap)
	vmMap := make(objectMap)
	clusterMap := make(objectMap)
	e := Endpoint{
		Url:        url,
		Parent:     parent,
		lastColl:   make(map[string]time.Time),
		hostMap:    hostMap,
		vmMap:      vmMap,
		clusterMap: clusterMap,
		nameCache:  make(map[string]string),
	}
	e.init()
	return e
}

func (e *Endpoint) init() error {
	conn, err := NewConnection(e.Url)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Load interval table
	//
	ctx := context.Background()
	list, err := conn.Perf.HistoricalInterval(ctx)
	if err != nil {
		return err
	}
	e.intervals = make([]int32, len(list))
	for k, i := range list {
		e.intervals[k] = i.SamplingPeriod
	}
	return nil
}

func (e *Endpoint) discover() error {
	conn, err := NewConnection(e.Url)
	if err != nil {
		return err
	}

	defer conn.Close()

	nameCache := make(map[string]string)

	// Discover clusters
	//
	ctx := context.Background()
	clusterMap, err := e.getClusters(ctx, conn.Root)
	if err != nil {
		return err
	}
	for _, cluster := range clusterMap {
		nameCache[cluster.ref.Reference().Value] = cluster.name
	}

	// Discover hosts
	//
	hostMap, err := e.getHosts(ctx, conn.Root)
	if err != nil {
		return err
	}
	for _, host := range hostMap {
		nameCache[host.ref.Reference().Value] = host.name
	}

	// Discover VMs
	//
	vmMap, err := e.getVMs(ctx, conn.Root)
	for _, vm := range vmMap {
		nameCache[vm.ref.Reference().Value] = vm.name
	}

	// Atomically swap maps
	//
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	e.nameCache = nameCache
	e.vmMap = vmMap
	e.hostMap = hostMap
	e.clusterMap = clusterMap
	return nil
}

func (e *Endpoint) collectResourceType(p *performance.Manager, ctx context.Context, alias string, acc telegraf.Accumulator,
	objects objectMap, nameCache map[string]string, intervalDuration internal.Duration) error {

	// Object maps may change, so we need to hold the collect lock
	//
	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	// Interval = 0 means collection for this metric was diabled, so don't even bother.
	//
	interval := int32(intervalDuration.Duration.Seconds())
	if interval <= 0 {
		return nil
	}
	log.Printf("D! Resource type: %s, Interval is %d", alias, interval)

	// Do we have new data yet?
	//
	now := time.Now()
	nIntervals := int32(1)
	latest, hasLatest := e.lastColl[alias]
	if hasLatest {
		elapsed := time.Now().Sub(latest).Seconds()
		if elapsed < float64(interval) {
			// No new data would be available. We're outta here!
			//
			return nil
		}
		nIntervals := int32(elapsed / (float64(interval)))
		if nIntervals > e.Parent.MaxSamples {
			nIntervals = e.Parent.MaxSamples
		}
	}
	e.lastColl[alias] = now
	log.Printf("D! Collecting %d intervals for %s", nIntervals, alias)
	fullAlias := "vsphere." + alias

	start := time.Now()
	log.Printf("D! Query for %s returned %d objects", alias, len(objects))
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.MaxQuery)
	total := 0
	for _, object := range objects {
		pq := types.PerfQuerySpec{
			Entity:     object.ref,
			MaxSample:  nIntervals,
			MetricId:   nil,
			IntervalId: interval,
		}
		if interval > 20 {
			startTime := now.Add(-time.Duration(interval) * time.Second)
			pq.StartTime = &startTime
			pq.EndTime = &now
		}
		if e.Parent.MaxSamples > 1 && hasLatest {
			pq.StartTime = &latest
			pq.EndTime = &now
		}
		pqs = append(pqs, pq)
		total++

		// Filled up a chunk or at end of data? Run a query with the collected objects
		//
		if len(pqs) >= int(e.Parent.MaxQuery) || total == len(objects) {
			log.Printf("D! Querying %d objects of type %s for %s. Total processed: %d. Total objects %d\n", len(pqs), alias, e.Url.Host, total, len(objects))
			metrics, err := p.Query(ctx, pqs)
			if err != nil {
				log.Printf("E! Error processing resource type %s", alias)
				return err
			}

			ems, err := p.ToMetricSeries(ctx, metrics)
			if err != nil {
				return err
			}

			// Iterate through result and fields list
			//
			for _, em := range ems {
				moid := em.Entity.Reference().Value
				for _, v := range em.Value {
					name := v.Name
					for idx, value := range v.Value {
						f := map[string]interface{}{name: value}
						objectName := nameCache[moid]
						parent := ""
						parentRef := objects[moid].parentRef
						if parentRef != nil {
							parent = nameCache[parentRef.Value]
						}

						t := map[string]string{
							"vcenter":  e.Url.Host,
							"hostname": objectName,
							"moid":     moid,
							"parent":   parent}
						if v.Instance != "" {
							t["instance"] = v.Instance
						}
						acc.AddFields(fullAlias, f, t, em.SampleInfo[idx].Timestamp)
					}
				}
			}
			pqs = make([]types.PerfQuerySpec, 0, e.Parent.MaxQuery)
		}
	}

	log.Printf("D! Collection of %s took %v\n", alias, time.Now().Sub(start))
	return nil
}

func (e *Endpoint) collect(acc telegraf.Accumulator) error {
	err := e.discover()
	if err != nil {
		return err
	}

	conn, err := NewConnection(e.Url)
	if err != nil {
		return err
	}

	defer conn.Close()

	ctx := context.Background()
	err = e.collectResourceType(conn.Perf, ctx, "cluster", acc, e.clusterMap, e.nameCache, e.Parent.ClusterInterval)
	if err != nil {
		return err
	}

	err = e.collectResourceType(conn.Perf, ctx, "host", acc, e.hostMap, e.nameCache, e.Parent.HostInterval)
	if err != nil {
		return err
	}

	err = e.collectResourceType(conn.Perf, ctx, "vm", acc, e.vmMap, e.nameCache, e.Parent.VmInterval)
	if err != nil {
		return err
	}
	return nil
}

func (e *Endpoint) getVMs(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.VirtualMachine
	err := root.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "runtime.host"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Runtime.Host}
	}
	return m, nil
}

func (e *Endpoint) getHosts(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.HostSystem
	err := root.Retrieve(ctx, []string{"HostSystem"}, []string{"summary", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func (e *Endpoint) getClusters(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"summary", "name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func (e *Endpoint) getDatastores(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datastore
	err := root.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.Summary.Name] = objectRef{ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

var sampleConfig = `
  ## List of vCenter URLs, including credentials. Note the "@" characted must be escaped as %40
  # vcenters = [ "https://administrator%40vsphere.local:VMware1!@vcenter.local/sdk" ]
`

func (v *VSphere) SampleConfig() string {
	return sampleConfig
}

func (v *VSphere) Description() string {
	return "Read metrics from VMware vCenter"
}

func (v *VSphere) vSphereInit() {
	if v.endpoints != nil {
		return
	}

	v.endpoints = make([]Endpoint, len(v.Vcenters))
	for i, rawUrl := range v.Vcenters {
		u, err := soap.ParseURL(rawUrl)
		if err != nil {
			log.Printf("E! Can't parse URL %s\n", rawUrl)
		}
		v.endpoints[i] = NewEndpoint(v, u)
	}
}

func (v *VSphere) Gather(acc telegraf.Accumulator) error {

	v.vSphereInit()

	start := time.Now()

	results := make(chan error)
	defer close(results)
	for _, ep := range v.endpoints {
		go func(target Endpoint) {
			results <- target.collect(acc)
		}(ep)
	}

	var finalErr error = nil
	for range v.endpoints {
		err := <-results
		if err != nil {
			log.Println("E!", err)
			finalErr = err
		}
	}

	// Add another counter to show how long it took to gather all the metrics on this cycle (can be used to tune # of vCenters and collection intervals per telegraf agent)
	acc.AddCounter("vsphere", map[string]interface{}{"gather.duration": time.Now().Sub(start).Seconds()}, nil, time.Now())

	return finalErr
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters:                []string{},
			VmInterval:              internal.Duration{Duration: time.Second * 20},
			HostInterval:            internal.Duration{Duration: time.Second * 20},
			ClusterInterval:         internal.Duration{Duration: time.Second * 300},
			DatastoreInterval:       internal.Duration{Duration: time.Second * 300},
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			MaxSamples:              10,
			MaxQuery:                64,
		}
	})
}
