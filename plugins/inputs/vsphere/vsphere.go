package vsphere

import (
	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"context"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
	"log"
	"net/url"
)

type Endpoint struct {
	Parent *VSphere
	Url *url.URL
	intervals []int32
	lastColl map[string]time.Time
}

type VSphere struct {
	Vcenters []string
	VmInterval int32
	HostInterval int32
	ClusterInterval int32
	DatastoreInterval int32
	MaxSamples int32
	MaxQuery int32
	endpoints []Endpoint
}

type objectRef struct {
	name string
	ref types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
}

type objectMap map[string]objectRef

type InstanceMetrics map[string]map[string]interface{}

func (e *Endpoint) init(p *performance.Manager) error {
	if e.intervals == nil {
		// Load interval table
		//
		ctx := context.Background()
		list, err := p.HistoricalInterval(ctx)
		if err != nil {
			return err
		}
		e.intervals = make([]int32, len(list))
		for k, i := range list {
			e.intervals[k] = i.SamplingPeriod
		}
	}
	return nil
}

func (e *Endpoint) collectResourceType(p *performance.Manager, ctx context.Context, alias string, acc telegraf.Accumulator,
	objects objectMap, nameCache map[string]string, interval int32) error {

	// Interval = -1 means collection for this metric was diabled, so don't even bother.
	//
	if interval == -1 {
		return nil
	}

	// Do we have new data yet?
	//
	now := time.Now()
	nIntervals := int32(1)
	latest, hasLatest := e.lastColl[alias]
	if (hasLatest) {
		elapsed := time.Now().Sub(latest).Seconds()
		if elapsed < float64(interval) {
			// No new data would be available. We're outta here!
			//
			return nil;
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
	total := 0;
	for _, object := range objects {
		pq := types.PerfQuerySpec{
			Entity: object.ref,
			MaxSample: nIntervals,
			MetricId: nil,
			IntervalId: interval,
		}
		if(interval > 20) {
			startTime := now.Add(-time.Duration(interval) * time.Second)
			pq.StartTime = &startTime
			pq.EndTime = &now
		}
		if(e.Parent.MaxSamples > 1 && hasLatest) {
			pq.StartTime = &latest
			pq.EndTime = &now
		}
		pqs = append(pqs, pq)
		total++

		// Filled up a chunk or at end of data? Run a query with the collected objects
		//
		if len(pqs) >= int(e.Parent.MaxQuery) || total == len(objects)  {
			log.Printf("D! Querying %d objects of type %s for %s. Total processed: %d. Total objects %d\n", len(pqs), alias, e.Url.Host, total, len(objects))
			metrics, err := p.Query(ctx, pqs)
			if err != nil {
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
						f := map[string]interface{} { name: value }
						objectName := nameCache[moid]
						parent := ""
						parentRef := objects[moid].parentRef
						//log.Printf("Parentref=%s", parentRef)
						if parentRef != nil {
							parent = nameCache[parentRef.Value]
						}

						t := map[string]string{
							"vcenter": e.Url.Host,
							"source": objectName,
							"moid": moid,
							"parent": parent}
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
	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, e.Url, true)
	if(err != nil) {
		return err
	}

	defer c.Logout(ctx)

	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{ }, true)
	if err != nil {
		return err
	}

	defer v.Destroy(ctx)

	p := performance.NewManager(c.Client)

	// This causes strange error messages in the vCenter console. Possibly due to a bug in
	// govmomi. We're commenting it out for now. Should be benign since the logout should
	// destroy all resources anyway.
	//
	//defer p.Destroy(ctx)

	// Load cache if needed
	e.init(p)

	nameCache := make(map[string]string)

	// Collect cluster metrics
	//
	clusterMap, err := e.getClusters(ctx, v);
	if err != nil {
		return err
	}
	for _, cluster := range clusterMap {
		nameCache[cluster.ref.Reference().Value] = cluster.name
	}
	err = e.collectResourceType(p, ctx, "cluster", acc, clusterMap, nameCache, e.Parent.ClusterInterval)
	if err != nil {
		return err
	}

	// Collect host metrics
	//
	hostMap, err := e.getHosts(ctx, v)
	if err != nil {
		return err
	}
	for _, host := range hostMap {
		nameCache[host.ref.Reference().Value] = host.name
	}
	err = e.collectResourceType(p, ctx, "host", acc, hostMap, nameCache, e.Parent.HostInterval)
	if err != nil {
		return err
	}

	// Collect vm metrics
	//
	vmMap, err := e.getVMs(ctx, v)
	for _, vm := range vmMap {
		nameCache[vm.ref.Reference().Value] = vm.name
	}
	err = e.collectResourceType(p, ctx, "vm", acc, vmMap, nameCache, e.Parent.VmInterval)
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
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Runtime.Host }
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
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent }
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
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent }
	}
	return m, nil
}

func (e *Endpoint) getDatastores(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datastore
	err := root.Retrieve(ctx, []string{"Datastore"}, []string{"summary" }, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.Summary.Name] = objectRef{ ref:r.ExtensibleManagedObject.Reference(), parentRef: r.Parent }
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

func (v *VSphere) Init()  {
	if v.endpoints != nil {
		return
	}
	v.endpoints = make([]Endpoint, len(v.Vcenters))
	for i, rawUrl := range v.Vcenters {
		u, err := soap.ParseURL(rawUrl);
		if(err != nil) {
			log.Printf("E! Can't parse URL %s\n", rawUrl)
		}
		v.endpoints[i] = Endpoint{
			Url: u,
			Parent: v,
			lastColl: make(map[string]time.Time)}
	}
}

func (v *VSphere) Gather(acc telegraf.Accumulator) error {
	v.Init()
	results := make(chan error)
	defer close(results)
	for _, ep := range v.endpoints {
		go func(target Endpoint) {
			results <- target.collect(acc)
		}(ep)
	}
	var finalErr error = nil
	for range v.endpoints {
		err := <- results
		if err != nil {
			log.Println("E!", err)
			finalErr = err
		}
	}
	return finalErr
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters: []string {},
			VmInterval: 20,
			HostInterval: 20,
			ClusterInterval: 300,
			DatastoreInterval: 300,
			MaxSamples: 10,
			MaxQuery: 64,
		}
	})
}



