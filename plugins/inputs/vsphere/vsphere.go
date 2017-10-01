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
	"sync"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
	"fmt"
)

type Endpoint struct {
	Parent *VSphere
	Url string
	intervals []int32
	mux sync.Mutex
}

type VSphere struct {
	Vcenters []string
	VmInterval int32
	HostInterval int32
	ClusterInterval int32
	DatastoreInterval int32
	endpoints []Endpoint
	mux sync.Mutex
}

type ResourceGetter func (context.Context, *view.ContainerView) (map[string]types.ManagedObjectReference, error)

type InstanceMetrics map[string]map[string]interface{}

func (e *Endpoint) Init(p *performance.Manager) error {
	e.mux.Lock()
	defer e.mux.Unlock()
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

func (e *Endpoint) CollectResourceType(p *performance.Manager, ctx context.Context, alias string, acc telegraf.Accumulator,
	getter ResourceGetter, root *view.ContainerView, interval int32) error {

	start := time.Now()
	objects, err := getter(ctx, root)
	if err != nil {
		return err
	}
	pqs := make([]types.PerfQuerySpec, len(objects))
	nameLookup := make(map[string]string)
	idx := 0
	for name, mor := range objects {
		nameLookup[mor.Reference().Value] = name;
		// Collect metrics
		//
		ams, err := p.AvailableMetric(ctx, mor, interval)
		if err != nil {
			return err
		}
		pqs[idx] = types.PerfQuerySpec{
			Entity: mor,
			MaxSample: 1,
			MetricId: ams,
			IntervalId: interval,
		}
		idx++
	}

	metrics, err := p.Query(ctx, pqs )
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
		im := make(InstanceMetrics)
		for _, v := range em.Value {
			name := v.Name
			m, found := im[v.Instance]
			if !found {
				m = make(map[string]interface{})
				im[v.Instance] = m
			}
			m[name] = v.Value[0]
		}
		for k, m := range im {
			moid := em.Entity.Reference().Value
			tags := map[string]string{
				"source": nameLookup[moid],
				"moid": moid}
			if k != "" {
				tags["instance"] = k
			}
			acc.AddFields("vsphere." + alias, m, tags)
		}
	}
	fmt.Println(time.Now().Sub(start))
	return nil
}

func (e *Endpoint) Collect(acc telegraf.Accumulator) error {
	ctx := context.Background()
	u, err := soap.ParseURL(e.Url)
	if(err != nil) {
		return err
	}
	c, err := govmomi.NewClient(ctx, u, true)
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
	p.Destroy(ctx)

	// Load cache if needed
	e.Init(p)

	err = e.CollectResourceType(p, ctx, "vm", acc, e.getVMs, v, e.Parent.VmInterval)
	if err != nil {
		return err
	}

	err = e.CollectResourceType(p, ctx, "host", acc, e.getHosts, v, e.Parent.HostInterval)
	if err != nil {
		return err
	}

	err = e.CollectResourceType(p, ctx, "cluster", acc, e.getClusters, v, e.Parent.ClusterInterval)
	if err != nil {
		return err
	}

	err = e.CollectResourceType(p, ctx, "datastore", acc, e.getDatastores, v, e.Parent.DatastoreInterval)
	if err != nil {
		return err
	}
	return nil
}

func (e *Endpoint) getVMs(ctx context.Context, root *view.ContainerView) (map[string]types.ManagedObjectReference, error) {
	var resources []mo.VirtualMachine
	err := root.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(map[string]types.ManagedObjectReference)
	for _, r := range resources {
		m[r.Summary.Config.Name] = r.ExtensibleManagedObject.Reference()
	}
	return m, nil
}

func (e *Endpoint) getHosts(ctx context.Context, root *view.ContainerView) (map[string]types.ManagedObjectReference, error) {
	var resources []mo.HostSystem
	err := root.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(map[string]types.ManagedObjectReference)
	for _, r := range resources {
		m[r.Summary.Config.Name] = r.ExtensibleManagedObject.Reference()
	}
	return m, nil
}

func (e *Endpoint) getClusters(ctx context.Context, root *view.ContainerView) (map[string]types.ManagedObjectReference, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"summary"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(map[string]types.ManagedObjectReference)
	for _, r := range resources {
		m[r.Name] = r.ExtensibleManagedObject.Reference()
	}
	return m, nil
}

func (e *Endpoint) getDatastores(ctx context.Context, root *view.ContainerView) (map[string]types.ManagedObjectReference, error) {
	var resources []mo.Datastore
	err := root.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(map[string]types.ManagedObjectReference)
	for _, r := range resources {
		m[r.Summary.Name] = r.ExtensibleManagedObject.Reference()
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
	v.mux.Lock()
	defer v.mux.Unlock()


	if v.endpoints != nil {
		return
	}
	v.endpoints = make([]Endpoint, len(v.Vcenters))
	for i, u := range v.Vcenters {
		v.endpoints[i] = Endpoint{ Url: u, Parent: v }
	}
}

func (v *VSphere) Gather(acc telegraf.Accumulator) error {
	v.Init()
	for _, ep := range v.endpoints {
		err := ep.Collect(acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters: []string {},
			VmInterval: 20,
			HostInterval: 20,
			ClusterInterval: 300,
			DatastoreInterval: 300,
		}
	})
}



