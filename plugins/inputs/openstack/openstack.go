// openstack implements an OpenStack input plugin for Telegraf
//
// The OpenStack input plug is a simple two phase metric collector.  In the first
// pass a set of gatherers are run against the API to cache collections of resources.
// In the second phase the gathered resources are combined and emitted as metrics.
//
// No aggregation is performed by the input plugin, instead queries to InfluxDB should
// be used to gather global totals of things such as tag frequency.
package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumetenants"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"strings"
)

const (
	// plugin is used to identify ourselves in log output
	plugin = "openstack"
)

// serviceType is an OpenStack service type
type serviceType string

const (
	volumeV2Service serviceType = "volumev2"
)

// tagMap maps a tag name to value.
type tagMap map[string]string

// fieldMap maps a field to an arbitrary data type.
type fieldMap map[string]interface{}

// serviceMap maps service id to a Service struct.
type serviceMap map[string]services.Service

// ContainsService indicates whether a particular service is enabled
func (s serviceMap) ContainsService(t serviceType) bool {
	for _, service := range s {
		if service.Type == string(t) {
			return true
		}
	}
	return false
}

// projectMap maps a project id to a Project struct.
type projectMap map[string]projects.Project

// hypervisorMap maps a hypervisor id a Hypervisor struct.
type hypervisorMap map[int]hypervisors.Hypervisor

// serverMap maps a server id to a Server struct.
type serverMap map[string]servers.Server

// flavorMap maps a flavor id to a Flavor struct.
type flavorMap map[string]flavors.Flavor

// volume is a structure used to unmarshal raw JSON from the API into.
type volume struct {
	volumes.Volume
	volumetenants.VolumeTenantExt
}

// volumeMap maps a volume id to a volume struct.
type volumeMap map[string]volume

// storagePoolMap maps a storage pool name to a StoragePool struct.
type storagePoolMap map[string]schedulerstats.StoragePool

// OpenStack is the main structure associated with a collection instance.
type OpenStack struct {
	// Configuration variables
	IdentityEndpoint string
	Domain           string
	Project          string
	Username         string
	Password         string

	// Locally cached clients
	identity *gophercloud.ServiceClient
	compute  *gophercloud.ServiceClient
	volume   *gophercloud.ServiceClient

	// Locally cached resources
	services     serviceMap
	projects     projectMap
	hypervisors  hypervisorMap
	flavors      flavorMap
	servers      serverMap
	volumes      volumeMap
	storagePools storagePoolMap
}

// Description returns a description string of the input plugin and implements
// the Input interface.
func (o *OpenStack) Description() string {
	return "Collects performance metrics from OpenStack services"
}

// sampleConfig is a sample configuration file entry.
var sampleConfig = `
  ## This is the recommended interval to poll.
  interval = '60m'

  ## The identity endpoint to authenticate against and get the
  ## service catalog from
  identity_endpoint = "https://my.openstack.cloud:5000"

  ## The domain to authenticate against when using a V3
  ## identity endpoint.  Defaults to 'default'
  domain = "default"

  ## The project to authenticate as
  project = "admin"

  ## The user to authenticate as, must have admin rights
  username = "admin"

  ## The user's password to authenticate with
  password = "Passw0rd"
`

// SampleConfig return a sample configuration file for auto-generation and
// implements the Input interface.
func (o *OpenStack) SampleConfig() string {
	return sampleConfig
}

// gather is a wrapper around library calls out to gophercloud that catches
// and recovers from panics.  Evidently if things like volumes don't exist
// then it will go down in flames.
func gather(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from crash: %v", r)
		}
	}()
	return f()
}

// Gather gathers resources from the OpenStack API and accumulates metrics.  This
// implements the Input interface.
func (o *OpenStack) Gather(acc telegraf.Accumulator) error {
	// Perform any required set up
	if err := o.initialize(); err != nil {
		return err
	}

	// Gather resources.  Note service harvesting must come first as the other
	// gatherers are dependant on this information.
	gatherers := map[string]func() error{
		"services":      o.gatherServices,
		"projects":      o.gatherProjects,
		"hypervisors":   o.gatherHypervisors,
		"flavors":       o.gatherFlavors,
		"servers":       o.gatherServers,
		"volumes":       o.gatherVolumes,
		"storage pools": o.gatherStoragePools,
	}
	for resources, gatherer := range gatherers {
		if err := gather(gatherer); err != nil {
			log.Println("W!", plugin, "failed to get", resources, ":", err)
		}
	}

	// Accumulate statistics
	accumulators := []func(telegraf.Accumulator){
		o.accumulateIdentity,
		o.accumulateHypervisors,
		o.accumulateServers,
		o.accumulateVolumes,
		o.accumulateStoragePools,
	}
	for _, accumulator := range accumulators {
		accumulator(acc)
	}

	return nil
}

// initialize performs any necessary initialization functions
func (o *OpenStack) initialize() error {
	// Authenticate against Keystone and get a token provider
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: o.IdentityEndpoint,
		DomainName:       o.Domain,
		TenantName:       o.Project,
		Username:         o.Username,
		Password:         o.Password,
	})
	if err != nil {
		return fmt.Errorf("Unable to authenticate OpenStack user: %v", err)
	}

	// Create required clients and attach to the OpenStack struct
	if o.identity, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V3 identity client: %v", err)
	}
	if o.compute, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 compute client: %v", err)
	}
	if o.volume, err = openstack.NewBlockStorageV2(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 block storage client: %v", err)
	}

	// Initialize resource maps and slices
	o.services = serviceMap{}
	o.projects = projectMap{}
	o.hypervisors = hypervisorMap{}
	o.flavors = flavorMap{}
	o.servers = serverMap{}
	o.volumes = volumeMap{}
	o.storagePools = storagePoolMap{}

	return nil
}

// gatherServices collects services from the OpenStack API.
func (o *OpenStack) gatherServices() error {
	page, err := services.List(o.identity, &services.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list services: %v", err)
	}
	services, err := services.ExtractServices(page)
	if err != nil {
		return fmt.Errorf("unable to extract services")
	}
	for _, service := range services {
		o.services[service.ID] = service
	}
	return nil
}

// gatherProjects collects projects from the OpenStack API.
func (o *OpenStack) gatherProjects() error {
	page, err := projects.List(o.identity, &projects.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list projects: %v", err)
	}
	projects, err := projects.ExtractProjects(page)
	if err != nil {
		return fmt.Errorf("unable to extract projects: %v", err)
	}
	for _, project := range projects {
		o.projects[project.ID] = project
	}
	return nil
}

// gatherHypervisors collects hypervisors from the OpenStack API.
func (o *OpenStack) gatherHypervisors() error {
	page, err := hypervisors.List(o.compute).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list hypervisors: %v", err)
	}
	hypervisors, err := hypervisors.ExtractHypervisors(page)
	if err != nil {
		return fmt.Errorf("unable to extract hypervisors: %v", err)
	}
	for _, hypervisor := range hypervisors {
		o.hypervisors[hypervisor.ID] = hypervisor
	}
	return nil
}

// gatherFlavors collects flavors from the OpenStack API.
func (o *OpenStack) gatherFlavors() error {
	page, err := flavors.ListDetail(o.compute, &flavors.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list flavors: %v", err)
	}
	flavors, err := flavors.ExtractFlavors(page)
	if err != nil {
		return fmt.Errorf("unable to extract flavors: %v", err)
	}
	for _, flavor := range flavors {
		o.flavors[flavor.ID] = flavor
	}
	return nil
}

// gatherServers collects servers from the OpenStack API.
func (o *OpenStack) gatherServers() error {
	page, err := servers.List(o.compute, &servers.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list servers: %v", err)
	}
	servers, err := servers.ExtractServers(page)
	if err != nil {
		return fmt.Errorf("unable to extract servers: %v", err)
	}
	for _, server := range servers {
		o.servers[server.ID] = server
	}
	return nil
}

// gatherVolumes collects volumes from the OpenStack API.
func (o *OpenStack) gatherVolumes() error {
	if !o.services.ContainsService(volumeV2Service) {
		return nil
	}
	page, err := volumes.List(o.volume, &volumes.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list volumes: %v", err)
	}
	v := []volume{}
	volumes.ExtractVolumesInto(page, &v)
	for _, volume := range v {
		o.volumes[volume.ID] = volume
	}
	return nil
}

// gatherStoragePools collects storage pools from the OpenStack API.
func (o *OpenStack) gatherStoragePools() error {
	if !o.services.ContainsService(volumeV2Service) {
		return nil
	}
	results, err := schedulerstats.List(o.volume, &schedulerstats.ListOpts{Detail: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list storage pools: %v", err)
	}
	storagePools, err := schedulerstats.ExtractStoragePools(results)
	if err != nil {
		return fmt.Errorf("unable to extract storage pools: %v", err)
	}
	for _, storagePool := range storagePools {
		o.storagePools[storagePool.Name] = storagePool
	}
	return nil
}

// accumulateIdentity accumulates statistics from the identity service.
func (o *OpenStack) accumulateIdentity(acc telegraf.Accumulator) {
	fields := fieldMap{
		"projects": len(o.projects),
	}
	acc.AddFields("openstack_identity", fields, tagMap{})
}

// accumulateHypervisors accumulates statistics from hypervisors.
func (o *OpenStack) accumulateHypervisors(acc telegraf.Accumulator) {
	for _, hypervisor := range o.hypervisors {
		tags := tagMap{
			"name": hypervisor.HypervisorHostname,
		}
		fields := fieldMap{
			"memory_mb":      hypervisor.MemoryMB,
			"memory_mb_used": hypervisor.MemoryMBUsed,
			"running_vms":    hypervisor.RunningVMs,
			"vcpus":          hypervisor.VCPUs,
			"vcpus_used":     hypervisor.VCPUsUsed,
		}
		acc.AddFields("openstack_hypervisor", fields, tags)
	}
}

// accumulateServers accumulates statistics about servers.
func (o *OpenStack) accumulateServers(acc telegraf.Accumulator) {
	for _, server := range o.servers {
		// Extract the flavor details to avoid joins (ignore errors and leave as zero values)
		var vcpus, ram, disk int
		if flavorIDInterface, ok := server.Flavor["id"]; ok {
			if flavorID, ok := flavorIDInterface.(string); ok {
				if flavor, ok := o.flavors[flavorID]; ok {
					vcpus = flavor.VCPUs
					ram = flavor.RAM
					disk = flavor.Disk
				}
			}
		}

		// Try derive the associated project
		project := "unknown"
		if p, ok := o.projects[server.TenantID]; ok {
			project = p.Name
		}

		tags := tagMap{
			"name":    server.Name,
			"project": project,
		}
		fields := fieldMap{
			"status":  strings.ToLower(server.Status),
			"vcpus":   vcpus,
			"ram_mb":  ram,
			"disk_gb": disk,
		}
		acc.AddFields("openstack_server", fields, tags)
	}
}

// accumulateVolumes accumulates statistics about volumes.
func (o *OpenStack) accumulateVolumes(acc telegraf.Accumulator) {
	for _, volume := range o.volumes {
		// Give empty types some form of field key
		volumeType := "unknown"
		if len(volume.VolumeType) != 0 {
			volumeType = volume.VolumeType
		}

		// Try derive the associated project
		project := "unknown"
		if p, ok := o.projects[volume.TenantID]; ok {
			project = p.Name
		}

		tags := tagMap{
			"name":    volume.Name,
			"project": project,
			"type":    volumeType,
		}
		fields := fieldMap{
			"size_gb": volume.Size,
		}
		acc.AddFields("openstack_volume", fields, tags)
	}
}

// accumulateStoragePools accumulates statistics about storage pools.
func (o *OpenStack) accumulateStoragePools(acc telegraf.Accumulator) {
	for _, storagePool := range o.storagePools {
		tags := tagMap{
			"name": storagePool.Capabilities.VolumeBackendName,
		}
		fields := fieldMap{
			"total_capacity_gb": storagePool.Capabilities.TotalCapacityGB,
			"free_capacity_gb":  storagePool.Capabilities.FreeCapacityGB,
		}
		acc.AddFields("openstack_storage_pool", fields, tags)
	}

}

// init registers a callback which creates a new OpenStack input instance.
func init() {
	inputs.Add("openstack", func() telegraf.Input {
		return &OpenStack{
			Domain: "Default",
		}
	})
}
