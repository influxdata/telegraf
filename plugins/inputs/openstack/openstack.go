// Package openstack implements an OpenStack input plugin for Telegraf
//
// The OpenStack input plug is a simple two phase metric collector.  In the first
// pass a set of gatherers are run against the API to cache collections of resources.
// In the second phase the gathered resources are combined and emitted as metrics.
//
// No aggregation is performed by the input plugin, instead queries to InfluxDB should
// be used to gather global totals of things such as tag frequency.
package openstack

import (
	"crypto/tls"
	"net/http"
	"fmt"
	"reflect"
	"strconv"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/diagnostics"
	nova_services "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/services"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
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
	Orchestration serviceType = "orchestration"
)

// tagMap maps a tag name to value.
type tagMap map[string]string

// fieldMap maps a field to an arbitrary data type.
type fieldMap map[string]interface{}

// serverDiag is a map of server diagnostics fields
type serverDiag  map[string]interface{}

// serviceMap maps service id to a Service struct.
type serviceMap map[string]services.Service

// stackMap maps stack id to a Stack struct.
type stackMap map[string]stacks.ListedStack

// nova_serviceMap maps nova_service id to a nova Service struct.
type nova_serviceMap map[string]nova_services.Service

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
type hypervisorMap map[string]hypervisors.Hypervisor

// serverMap maps a server id to a Server struct.
type serverMap map[string]servers.Server

// flavorMap maps a flavor id to a Flavor struct.
type flavorMap map[string]flavors.Flavor

// subnetMap maps a subnet id to a Subnet struct.
type subnetMap map[string]subnets.Subnet

// portMap maps a port id to a Port struct.
type portMap map[string]ports.Port

// networkMap maps a network id to a Network struct.
type networkMap map[string]networks.Network

// aggregateMap maps a aggregate id to a Aggregate struct.
type aggregateMap map[int]aggregates.Aggregate


// agentMap maps a agent id to a agents struct.
type agentMap map[string]agents.Agent

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
	IdentityEndpoint    string
	Domain              string
	Project             string
	Username            string
	Password            string
	EnabledServices     []string
	ServerDiagnotics    bool
	InsecureSkipVerify  bool

	// Locally cached clients
	identity   *gophercloud.ServiceClient
	compute    *gophercloud.ServiceClient
	volume     *gophercloud.ServiceClient
	network    *gophercloud.ServiceClient
	stack      *gophercloud.ServiceClient

	// Locally cached resources
	services       serviceMap
	projects       projectMap
	hypervisors    hypervisorMap
	flavors        flavorMap
	servers        serverMap
	volumes        volumeMap
	storagePools   storagePoolMap
	subnets        subnetMap
	ports          portMap
	networks       networkMap
	aggregates     aggregateMap
	nova_services  nova_serviceMap
	agents         agentMap
	stacks         stackMap
	diag           serverDiag
}

// Description returns a description string of the input plugin and implements
// the Input interface.
func (o *OpenStack) Description() string {
	return "Collects performance metrics from OpenStack services"
}

// sampleConfig is a sample configuration file entry.
var sampleConfig = `
  ## This is the recommended interval to poll.
  interval = '30m'

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

  ## Services to be enabled
  #enabled_services = ["stacks","services", "projects", "hypervisors", "flavors", "servers", "volumes", "storage" , "subnets", "ports", "networks", "aggregates", "nova_services", "agents"]
  enabled_services = ["services", "projects", "hypervisors", "flavors", "networks", "volumes"]

  #Dependencies
  # | Service | Depends on |
  # | servers | projects, hypervisors, flavors |
  # | volumes | projects |

  ## Collect Server Diagnostics
  server_diagnotics = false

  InsecureSkipVerify = false
`

// SampleConfig return a sample configuration file for auto-generation and
// implements the Input interface.
func (o *OpenStack) SampleConfig() string {
	return sampleConfig
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
		"projects":        o.gatherProjects,
		"hypervisors":     o.gatherHypervisors,
		"flavors":         o.gatherFlavors,
		"servers":         o.gatherServers,
		"volumes":         o.gatherVolumes,
		"storage pools":   o.gatherStoragePools,
		"subnets":         o.gatherSubnets,
		"ports":           o.gatherPorts,
		"networks":        o.gatherNetworks,
		"aggregates":      o.gatherAggregates,
		"nova_services":   o.gatherNovaServices,
		"agents":          o.gatherAgents,
		"stacks":          o.gatherStacks,
	}
	for resources, gatherer := range gatherers {
		for _, i := range o.EnabledServices {
		    if resources == i {
		        if err := gatherer(); err != nil {
			        log.Println("W!", plugin, "failed to get", resources, err)
				}
			}
	    }
	}

	// Accumulate statistics
	accumulators := map[string]func(telegraf.Accumulator){
	    "services":       o.accumulateServices,
	    "projects":       o.accumulateIdentity,
	    "hypervisors":    o.accumulateHypervisors,
	    "flavors":        o.accumulateFlavors,
	    "servers":        o.accumulateServers,
	    "volumes":        o.accumulateVolumes,
	    "storage pools":  o.accumulateStoragePools,
	    "subnets":        o.accumulateSubnets,
	    "ports":          o.accumulatePorts,
	    "networks":       o.accumulateNetworks,
	    "aggregates":     o.accumulateAggregates,
	    "nova_services":  o.accumulateNovaServices,
	    "agents":         o.accumulateAgents,
	    "stacks":         o.accumulateStacks,
	}
	for resources, accumulator := range accumulators {
		for _, i := range o.EnabledServices {
		    if resources == i {
				accumulator(acc)
			}
		}
	}
	return nil
}

// initialize performs any necessary initialization functions
func (o *OpenStack) initialize() error {
	// Authenticate against Keystone and get a token provider
	authOption := gophercloud.AuthOptions{
		IdentityEndpoint: o.IdentityEndpoint,
		DomainName:       o.Domain,
		TenantName:       o.Project,
		Username:         o.Username,
		Password:         o.Password,
	}
	provider, err := openstack.NewClient(authOption.IdentityEndpoint)
	if err != nil {
		return fmt.Errorf("Unable to create Newclient for OpenStack endpoint: %v", err)
	}

	tlsconfig := &tls.Config{}
	tlsconfig.InsecureSkipVerify = o.InsecureSkipVerify
	transport := &http.Transport{TLSClientConfig: tlsconfig}
	provider.HTTPClient = http.Client{
		Transport: transport,
	}

	err = openstack.Authenticate(provider, authOption)

	if err != nil {
		return fmt.Errorf("Unable to authenticate OpenStack user: %v", err)
	}

	// Create required clients and attach to the OpenStack struct
	if o.identity, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V3 identity client: %v", err)
	}

	o.services = serviceMap{}
	o.gatherServices()

	if o.compute, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 compute client: %v", err)
	}

	// Create required clients and attach to the OpenStack struct
	if o.network, err = openstack.NewNetworkV2(provider,gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 network client: %v", err)
	}

	// The Orchestration service is optional
	if o.services.ContainsService(Orchestration) {
		if o.stack, err = openstack.NewOrchestrationV1(provider,gophercloud.EndpointOpts{}); err != nil {
			return fmt.Errorf("unable to create V1 stack client: %v", err)
		}
	}

	// The Cinder volume storage service is optional
	if o.services.ContainsService(volumeV2Service) {
		if o.volume, err = openstack.NewBlockStorageV2(provider, gophercloud.EndpointOpts{}); err != nil {
			return fmt.Errorf("unable to create V2 volume client: %v", err)
		}
	}

	// Initialize resource maps and slices
	o.projects = projectMap{}
	o.hypervisors = hypervisorMap{}
	o.flavors = flavorMap{}
	o.servers = serverMap{}
	o.volumes = volumeMap{}
	o.storagePools = storagePoolMap{}
	o.subnets = subnetMap{}
	o.networks = networkMap{}
	o.ports = portMap{}
	o.aggregates = aggregateMap{}
	o.nova_services = nova_serviceMap{}
	o.agents = agentMap{}
        o.diag = serverDiag{}
	return nil
}

// gatherStacks collects stacks from the OpenStack API.
func (o *OpenStack) gatherStacks() error {
	page, err := stacks.List(o.stack, &stacks.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list stacks: %v", err)
	}
	stacks, err := stacks.ExtractStacks(page)
	if err != nil {
		return fmt.Errorf("unable to extract stacks")
	}
	for _, stack := range stacks {
		o.stacks[stack.ID] = stack
	}
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

// gathernova_services collects nova_services from the OpenStack API.
func (o *OpenStack) gatherNovaServices() error {
	page, err := nova_services.List(o.compute, &nova_services.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list nova_services: %v", err)
	}
	nova_services, err := nova_services.ExtractServices(page)
	if err != nil {
		return fmt.Errorf("unable to extract nova_services")
	}
	for _, nova_service := range nova_services {
		o.nova_services[nova_service.ID] = nova_service
	}
	return nil
}

// gatherSubnets collects subnets from the OpenStack API.
func (o *OpenStack) gatherSubnets() error {
	page, err := subnets.List(o.network, &subnets.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list subnets: %v", err)
	}
	subnets, err := subnets.ExtractSubnets(page)
	if err != nil {
		return fmt.Errorf("unable to extract subnets")
	}
	for _, subnet := range subnets {
		o.subnets[subnet.ID] = subnet
	}
	return nil
}

// gatherPorts collects ports from the OpenStack API.
func (o *OpenStack) gatherPorts() error {
	page, err := ports.List(o.network, &ports.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list ports: %v", err)
	}
	ports, err := ports.ExtractPorts(page)
	if err != nil {
		return fmt.Errorf("unable to extract ports")
	}
	for _, port := range ports {
		o.ports[port.ID] = port
	}
	return nil
}

// gatherNetworks collects networks from the OpenStack API.
func (o *OpenStack) gatherNetworks() error {
	page, err := networks.List(o.network, &networks.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list networks: %v", err)
	}
	networks, err := networks.ExtractNetworks(page)
	if err != nil {
		return fmt.Errorf("unable to extract networks")
	}
	for _, network := range networks {
		o.networks[network.ID] = network
	}
	return nil
}

// gatherAgents collects agents from the OpenStack API.
func (o *OpenStack) gatherAgents() error {
	page, err := agents.List(o.network, &agents.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list newtron agents: %v", err)
	}
	agents, err := agents.ExtractAgents(page)
	if err != nil {
		return fmt.Errorf("unable to extract newtron agents")
	}
	for _, agent := range agents {
		o.agents[agent.ID] = agent
	}
	return nil
}


// gatherAggregates collects aggregates from the OpenStack API.
func (o *OpenStack) gatherAggregates() error {
	page, err := aggregates.List(o.compute).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list aggregates: %v", err)
	}
	aggregates, err := aggregates.ExtractAggregates(page)
	if err != nil {
		return fmt.Errorf("unable to extract aggregates")
	}
	for _, aggregate := range aggregates {
		o.aggregates[aggregate.ID] = aggregate
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
		if ( o.ServerDiagnotics && server.Status == "ACTIVE" ) {
			diagnostic, err := diagnostics.Get(o.compute, server.ID).Extract()
			o.diag[server.ID] = diagnostic
            if err != nil {
	            return fmt.Errorf("unable to get diagnostics for server(%v) %v", server.ID, err)
			}
		}
	}
	return nil
}

// gatherVolumes collects volumes from the OpenStack API.
func (o *OpenStack) gatherVolumes() error {
	//if !o.services.ContainsService(volumeV2Service) {
	//	return nil
	//}
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
	//if !o.services.ContainsService(volumeV2Service) {
	//	return nil
	//}
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

// accumulateServerDiagnostics accumulates statistics from the compute(nova) service.
func (o *OpenStack) accumulateServerDiagnostics(acc telegraf.Accumulator) {
	for server_id, diagnostic := range o.diag {
	    tags := tagMap{
			"server_id":  server_id,
		}
		fields := fieldMap{ 
			"server_id":  server_id,
		}
		switch reflect.TypeOf(diagnostic).Kind() {
		case reflect.Map:
			s := reflect.ValueOf(diagnostic).MapRange()
			port_map := make(map[string]string)
			port_count := 0
			for s.Next() {
				if strings.Contains(s.Key().Interface().(string), "tap") {
					tap_list := strings.Split(s.Key().Interface().(string), "_")
					if val, ok := port_map[tap_list[0]]; ok {
						fields["port_"+val+"_"+strings.Join(tap_list[1:], "_")]  = s.Value().Interface().(float64)
						tags["port_"+val] = tap_list[0]						
					} else {
						port_count += 1
						port_map[tap_list[0]] = strconv.Itoa(port_count)
						fields["port_"+port_map[tap_list[0]]+"_"+strings.Join(tap_list[1:], "_")]  = s.Value().Interface().(float64)
						tags["port_"+port_map[tap_list[0]]] = tap_list[0]
					}
				} else {
					fields[s.Key().Interface().(string)] = s.Value().Interface().(float64)
				}
			}
			fields["no_of_ports"] = port_count
		}	
		acc.AddFields("openstack_server_diagnostics", fields, tags)
	}
}

// accumulateStacks accumulates statistics from the stack service.
func (o *OpenStack) accumulateStacks(acc telegraf.Accumulator) {
	for _, stack := range o.stacks {
		tags := tagMap{
			"id":              stack.ID,
		}
	    fields := fieldMap{
			"creation_time":       stack.CreationTime.Format("2006-01-02T15:04:05.999999999Z07:00"),                            
			"description":         stack.Description,                  
			"id":                  stack.ID,         
			"stack_name":          stack.Name,                 
			"stack_status":        stack.Status,                   
			"stack_status_reason": stack.StatusReason,                          
			"updated_time":        stack.UpdatedTime.Format("2006-01-02T15:04:05.999999999Z07:00"),                                                      
	    }
	    acc.AddFields("openstack_stack", fields, tags)
    }
}

// accumulateFlavors accumulates statistics from the flavor service.
func (o *OpenStack) accumulateFlavors(acc telegraf.Accumulator) {
	for _, flavor := range o.flavors {
		tags := tagMap{
			"id":              flavor.ID,
		}
	    fields := fieldMap{  
			"id":              flavor.ID,        
			"disk":            flavor.Disk,          
			"ram":             flavor.RAM,         
			"name":            flavor.Name,          
			"rxtx_factor":     flavor.RxTxFactor,                 
			"swap_mb":         flavor.Swap,       
			"vcpus":           flavor.VCPUs,           
			"is_public":       flavor.IsPublic,               
			"ephemeral":       flavor.Ephemeral,                             
	    }
	    acc.AddFields("openstack_flavor", fields, tags)
    }
}


// accumulateIdentity accumulates statistics from the identity service.
func (o *OpenStack) accumulateIdentity(acc telegraf.Accumulator) {
	for _, project := range o.projects {
		tags := tagMap{
			"id":              project.ID,
		}
	    fields := fieldMap{
		    "projects":        len(o.projects),
		    "is_domain":       project.IsDomain,             
		    "description":     project.Description,               
		    "domain_id":       project.DomainID,             
		    "enabled":         project.Enabled,           
		    "id":              project.ID,      
		    "name":            project.Name,        
		    "parent_id":       project.ParentID,                  
	    }
	    acc.AddFields("openstack_identity", fields, tags)
    }
}

// accumulateHypervisors accumulates statistics from hypervisors.
func (o *OpenStack) accumulateHypervisors(acc telegraf.Accumulator) {
	for _, hypervisor := range o.hypervisors {
		tags := tagMap{
			"id":                      hypervisor.ID,
		}
		fields := fieldMap{
			"cpu_vendor":              hypervisor.CPUInfo.Vendor,
			"cpu_arch":                hypervisor.CPUInfo.Arch,
			"cpu_model":               hypervisor.CPUInfo.Model,
			"cpu_features":            hypervisor.CPUInfo.Features,
			"cpu_topology_sockets":    hypervisor.CPUInfo.Topology.Sockets,    
            "cpu_topology_cores":      hypervisor.CPUInfo.Topology.Cores,
            "cpu_topology_threads":    hypervisor.CPUInfo.Topology.Threads,
			"current_workload":        hypervisor.CurrentWorkload,               
			"status":                  hypervisor.Status,     
			"state":                   hypervisor.State,    
			"disk_available_least":    hypervisor.DiskAvailableLeast,                   
			"host_ip":                 hypervisor.HostIP,      
			"free_disk_gb":            hypervisor.FreeDiskGB,
			"free_ram_mb":             hypervisor.FreeRamMB,          
			"hypervisor_hostname":     hypervisor.HypervisorHostname,                  
			"hypervisor_type":         hypervisor.HypervisorType,              
			"version":                 hypervisor.HypervisorVersion,
			"id":                      hypervisor.ID,
			"local_gb":                hypervisor.LocalGB,
			"local_gb_used":           hypervisor.LocalGBUsed,            
			"memory_mb":               hypervisor.MemoryMB,        
			"memory_mb_used":          hypervisor.MemoryMBUsed,             
			"running_vms":             hypervisor.RunningVMs,                
			"service_host":            hypervisor.Service.Host,
			"service_id":              hypervisor.Service.ID,
			"service_disabled_reason": hypervisor.Service.DisabledReason,  
			"vcpus":                   hypervisor.VCPUs,    
			"vcpus_used":              hypervisor.VCPUsUsed,         
		}
		acc.AddFields("openstack_hypervisor", fields, tags)
	}
}

// accumulateServers accumulates statistics about servers.
func (o *OpenStack) accumulateServers(acc telegraf.Accumulator) {
	var compute_hosts = map[string]map[string]string{}
	for _, project := range o.projects {
		compute_hosts[project.ID] = map[string]string{}
		for _, hypervisor := range o.hypervisors {
			h := sha256.New224()
			h.Write([]byte(string(project.ID) + string(hypervisor.HypervisorHostname)))
			compute_hosts[project.ID][hex.EncodeToString(h.Sum(nil))] = hypervisor.HypervisorHostname
		}
	}
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
			"id":               server.ID,  
			"name":             server.Name,
			"project":          project,
			"tenant_id":        server.TenantID,
			"host_id":          server.HostID,
		}

		host_name := "na"
		for k,v := range compute_hosts[server.TenantID] {
			if k == server.HostID {
				host_name = v
			}
		}
		fields := fieldMap{
			"status":           strings.ToLower(server.Status),
			"vcpus":            vcpus,
			"ram_mb":           ram,
			"disk_gb":          disk,
			"id":               server.ID,     
            "tenant_id":        server.TenantID,            
            "user_id":          server.UserID,          
            "name":             server.Name,       
            "updated":          server.Updated.Format("2006-01-02T15:04:05.999999999Z07:00"),          
			"created":          server.Created.Format("2006-01-02T15:04:05.999999999Z07:00"),          
			"host_name":        host_name,
            "host_id":          server.HostID,        
            "progress":         server.Progress,           
            "accessIPv4":       server.AccessIPv4,             
            "accessIPv6":       server.AccessIPv6,             
            "image":            server.Image["id"],    
            "flavor":           server.Flavor["id"],         
            "addresses":        len(server.Addresses),            
            "key_name":         server.KeyName,           
            "adminPass":        server.AdminPass,            
            "security_groups":  len(server.SecurityGroups),                  
            "volumes_attached": len(server.AttachedVolumes),                   
			"fault_code":       server.Fault.Code,
			"fault_created":    server.Fault.Created,
			"fault_details":    server.Fault.Details,
			"fault_message":    server.Fault.Message, 
		}
		acc.AddFields("openstack_server", fields, tags)
	}
	if  (o.ServerDiagnotics) {
		o.accumulateServerDiagnostics(acc)
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

		attachment_attached_at   := "na"           
		attachment_attachment_id := "na"           
		attachment_device        := "na"    
		attachment_host_name     := "na"       
		attachment_id            := "na"
		attachment_server_id     := "na"       
		attachment_volume_id     := "na"       

		// only getting first attachment in Attachments 
		for _, attachment := range volume.Attachments {
			            
				attachment_attached_at     =  attachment.AttachedAt.Format("2006-01-02T15:04:05.999999999Z07:00") 
				attachment_attachment_id   =  attachment.AttachmentID                    
		        attachment_device          =  attachment.Device     
		        attachment_host_name       =  attachment.HostName        
		        attachment_id              =  attachment.ID 
		        attachment_server_id       =  attachment.ServerID        
				attachment_volume_id       =  attachment.VolumeID   
			//	break
			//}
		
			tags := tagMap{
				"name":                             volume.Name,
				"project":                          project,
				"type":                             volumeType,
		        "id":                               volume.ID, 
		        "attachment_server_id":             attachment_server_id, 
			} 
			
			fields := fieldMap{ 
				"size_gb":                          volume.Size,
		        "id":                               volume.ID,             
		        "status":                           volume.Status,                 
		        "size":                             volume.Size,               
		        "availability_zone":                volume.AvailabilityZone,                            
		        "created_at":                       volume.CreatedAt,            
				"updated_at":                       volume.UpdatedAt,
				"total_attachments":                len(volume.Attachments),            
				"attachment_attached_at":           attachment_attached_at,  
				"attachment_attachment_id":         attachment_attachment_id,
		        "attachment_device":                attachment_device,       
		        "attachment_host_name":             attachment_host_name,    
		        "attachment_id":                    attachment_id,           
		        "attachment_server_id":             attachment_server_id,    
		        "attachment_volume_id":             attachment_volume_id,    
		        "name":                             volume.Name,               
		        "description":                      volume.Description,                      
		        "volume_type":                      volume.VolumeType,                      
		        "snapshot_id":                      volume.SnapshotID,                      
		        "source_volid":                     volume.SourceVolID,                       
		        "user_id":                          volume.UserID,                  
		        "bootable":                         volume.Bootable,                   
		        "encrypted":                        volume.Encrypted,                    
		        "replication_status":               volume.ReplicationStatus,                             
		        "consistency_group_id":             volume.ConsistencyGroupID,                              
		        "multiattach":                      volume.Multiattach,                      
			}
			acc.AddFields("openstack_volume", fields, tags)
		}
	}
}

// accumulateStoragePools accumulates statistics about storage pools.
func (o *OpenStack) accumulateStoragePools(acc telegraf.Accumulator) {
	for _, storagePool := range o.storagePools {
		tags := tagMap{
			"name": storagePool.Capabilities.VolumeBackendName,
		}
		fields := fieldMap{
			"total_capacity_gb":      storagePool.Capabilities.TotalCapacityGB,
			"free_capacity_gb":       storagePool.Capabilities.FreeCapacityGB,
			"driver_version":         storagePool.Capabilities.DriverVersion,   
			"storage_protocol":       storagePool.Capabilities.StorageProtocol,     
			"vendor_name":            storagePool.Capabilities.VendorName,
			"volume_backend_name":    storagePool.Capabilities.VolumeBackendName,        
		}
		acc.AddFields("openstack_storage_pool", fields, tags)
	}

}

// accumulateServices accumulates statistics from services.
func (o *OpenStack) accumulateServices(acc telegraf.Accumulator) {
	for _, service := range o.services {
		tags := tagMap{
			"service_id":      service.ID,
		}
		fields := fieldMap{
			"name":            service.Type,
			"service_id":      service.ID,
			"service_enabled": service.Enabled,
		}
		acc.AddFields("openstack_service", fields, tags)
	}
}

// accumulateSubnets accumulates statistics from subnets.
func (o *OpenStack) accumulateSubnets(acc telegraf.Accumulator) {
	for _, subnet := range o.subnets {
		tags := tagMap{
			"subnet_id":      subnet.ID,
		}
		fields := fieldMap{
			"subnet_id":          subnet.ID,
			"name":               subnet.Name,
			"network_id":         subnet.NetworkID,
			"ip_version":         strconv.Itoa(subnet.IPVersion),
			"tenant_id":          subnet.TenantID,
			"project_id":         subnet.ProjectID,
			"dhcp_enabled":       subnet.EnableDHCP,
			"cidr":               subnet.CIDR,
			"gateway_ip":         subnet.GatewayIP,
			"dns_nameservers":    len(subnet.DNSNameservers),
			"ipv6_address_mode":  subnet.IPv6AddressMode,
			"ipv6_ra_mode":       subnet.IPv6RAMode,
			"subnet_pool_id":     subnet.SubnetPoolID,
		}
		acc.AddFields("openstack_subnet", fields, tags)
	}
}

// accumulateNetworks accumulates statistics from networks.
func (o *OpenStack) accumulateNetworks(acc telegraf.Accumulator) {
	for _, network := range o.networks {
		tags := tagMap{
			"id":                        network.ID,   
            "tenant_id":                 network.TenantID,
		}
		fields := fieldMap{
			"id":                        network.ID,
            "description":               network.Description,       
            "admin_state_up":            network.AdminStateUp,          
            "status":                    network.Status,
			"name":                      network.Name,
            "subnets":                   len(network.Subnets),   
            "tenant_id":                 network.TenantID,     
            "updated_at":                network.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
            "created_at":                network.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),      
            "project_id":                network.ProjectID,      
            "shared":                    network.Shared,  
            "availability_zone_hints":   len(network.AvailabilityZoneHints),
		}
		acc.AddFields("openstack_network", fields, tags)
	}
}

// accumulatePorts accumulates statistics from ports.
func (o *OpenStack) accumulatePorts(acc telegraf.Accumulator) {
	for _, port := range o.ports {
		tags := tagMap{
			"id":                    port.ID,          
            "status":                port.Status,
            "network_id":            port.NetworkID,
		}
		fields := fieldMap{
			"id":                    port.ID,
            "network_id":            port.NetworkID,        
            "description":           port.Description,          
            "admin_state_up":        port.AdminStateUp,
			"name":                  port.Name,          
            "status":                port.Status,    
            "mac_address":           port.MACAddress,         
            "fixed_ips":             len(port.FixedIPs),       
            "tenant_id":             port.TenantID,       
            "project_id":            port.ProjectID,        
            "device_owner":          port.DeviceOwner,          
            "security_groups":       len(port.SecurityGroups),             
            "device_id":             port.DeviceID,       
            "allowed_address_pairs": len(port.AllowedAddressPairs),
		}
		acc.AddFields("openstack_port", fields, tags)
	}
}

// accumulateAggregates accumulates statistics from aggregates.
func (o *OpenStack) accumulateAggregates(acc telegraf.Accumulator) {
	for _, aggregate := range o.aggregates {
		tags := tagMap{
			"id":    strconv.Itoa(aggregate.ID),
			"name":  aggregate.Name,
		}
		fields := fieldMap{    
			"id":                        strconv.Itoa(aggregate.ID),
			"availability_zone":         aggregate.AvailabilityZone,   
			"name":                      aggregate.Name,                
            "hosts":                     len(aggregate.Hosts),       
			"updated_at":                aggregate.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
            "created_at":                aggregate.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),  
            "deleted_at":                aggregate.DeletedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),   
            "deleted":                   aggregate.Deleted,         
		}
		acc.AddFields("openstack_aggregate", fields, tags)
	}
}

// accumulateNovaServices accumulates statistics from nova_services.
func (o *OpenStack) accumulateNovaServices(acc telegraf.Accumulator) {
	for _, nova_service := range o.nova_services {
		tags := tagMap{
			"id":              nova_service.ID,
			"binary":          nova_service.Binary,
		}
		fields := fieldMap{       
			"id":                        nova_service.ID,  
			"updated_at":                nova_service.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"binary":                    nova_service.Binary,       
            "disabled_reason":           nova_service.DisabledReason,                
			"host_machine":              nova_service.Host,      
            "state":                     nova_service.State,      
            "status":                    nova_service.Status,       
            "zone":                      nova_service.Zone,     
         
		}
		acc.AddFields("openstack_nova_service", fields, tags)
	}
}

// accumulateAgents accumulates statistics from agents.
func (o *OpenStack) accumulateAgents(acc telegraf.Accumulator) {
	for _, agent := range o.agents {
		tags := tagMap{
			"id":                        agent.ID,
			"binary":                    agent.Binary,
		}
		fields := fieldMap{
			"admin_state_up":            agent.AdminStateUp,             
			"agent_type":                agent.AgentType,         
			"alive":                     agent.Alive,
			"binary":                    agent.Binary,
			"availability_zone":         agent.AvailabilityZone,             
			"created_at":                agent.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),  
			"started_at":                agent.StartedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"heartbeat_timestamp":       agent.HeartbeatTimestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"description":               agent.Description,          
			"host_name":                 agent.Host,   
			"topic":                     agent.Topic,             
		}
		acc.AddFields("openstack_newtron_agent", fields, tags)
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

