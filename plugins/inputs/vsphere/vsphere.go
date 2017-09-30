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
)

const Interval = 20

type Endpoint struct {
	Url string
	ciCache map[int32]types.PerfCounterInfo
	mux sync.Mutex
}

type VSphere struct {
	Vcenters []string
	endpoints []Endpoint
}

func (e *Endpoint) Init(p *performance.Manager) error {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.ciCache != nil {
		return nil
	}
	ctx := context.Background()
	defer p.Destroy(ctx)
	e.ciCache = make(map[int32]types.PerfCounterInfo)
	cis, err := p.CounterInfo(ctx)
	if err != nil {
		return err
	}
	for _, ci := range cis {
		e.ciCache[ci.Key] = ci
	}
	return nil
}

func (e *Endpoint) CollectResourceType(p *performance.Manager, ctx context.Context, alias string, acc telegraf.Accumulator,
	objects map[string]types.ManagedObjectReference) error {

	for name, mor := range objects {
		// Collect metrics
		//
		ams, err := p.AvailableMetric(ctx, mor, Interval)
		if err != nil {
			return err
		}
		pqs := types.PerfQuerySpec{
			Entity: mor,
			MaxSample: 1,
			MetricId: ams,
			IntervalId: 20,
		}
		metrics, err := p.Query(ctx, []types.PerfQuerySpec{ pqs })
		if err != nil {
			return err
		}
		fields := make(map[string]interface{})
		ems, err := p.ToMetricSeries(ctx, metrics)
		if err != nil {
			return err
		}

		// Iterate through result and fields list
		//
		for _, em := range ems {
			for _, v := range em.Value {
				name := v.Name
				if v.Instance != "" {
					name += "." + v.Instance
				}
				fields[name] = v.Value[0]
			}
		}
		tags := map[string]string {
			"entityName": name,
			"entityId": mor.Value}
		acc.AddFields("vsphere." + alias, fields, tags)
	}
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
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return err
	}

	defer v.Destroy(ctx)

	p := performance.NewManager(c.Client)
	p.Destroy(ctx)

	// Load cache if needed
	e.Init(p)

	vms, err := e.getVMs(ctx, v)
	if err != nil {
		return err
	}
	err = e.CollectResourceType(p, ctx, "vm", acc, vms)
	if err != nil {
		return err
	}
	hosts, err := e.getHosts(ctx, v)
	if err != nil {
		return err
	}
	err = e.CollectResourceType(p, ctx, "host", acc, hosts)
	if err != nil {
		return err
	}
	clusters, err := e.getClusters(ctx, v)
	if err != nil {
		return err
	}
	err = e.CollectResourceType(p, ctx, "cluster", acc, clusters)
	if err != nil {
		return err
	}
	datastores, err := e.getDatastores(ctx, v)
	if err != nil {
		return err
	}
	err = e.CollectResourceType(p, ctx, "datastore", acc, datastores)
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

func (v *VSphere) Gather(acc telegraf.Accumulator) error {
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
		return &VSphere{ Vcenters: []string {} }
	})
}



