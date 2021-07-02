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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumetenants"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/diagnostics"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	nova_services "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/services"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// serviceType is an OpenStack service type
type serviceType string

const (
	// plugin is used to identify ourselves in log output
	plugin                      = "openstack"
	volumeV2Service serviceType = "volumev2"
	Orchestration   serviceType = "orchestration"
)

// volume is a structure used to unmarshal raw JSON from the API into.
type volume struct {
	volumes.Volume
	volumetenants.VolumeTenantExt
}

// OpenStack is the main structure associated with a collection instance.
type OpenStack struct {
	// Configuration variables
	IdentityEndpoint string          `toml:"authentication_endpoint"`
	Domain           string          `toml:"domain"`
	Project          string          `toml:"project"`
	Username         string          `toml:"username"`
	Password         string          `toml:"password"`
	EnabledServices  []string        `toml:"enabled_services"`
	ServerDiagnotics bool            `toml:"server_diagnotics"`
	Timeout          config.Duration `toml:"timeout"`

	// Locally cached clients
	identity *gophercloud.ServiceClient
	compute  *gophercloud.ServiceClient
	volume   *gophercloud.ServiceClient
	network  *gophercloud.ServiceClient
	stack    *gophercloud.ServiceClient

	// Locally cached resources
	agents         map[string]agents.Agent
	aggregates     map[int]aggregates.Aggregate
	diag           map[string]interface{}
	flavors        map[string]flavors.Flavor
	hypervisors    map[string]hypervisors.Hypervisor
	networks       map[string]networks.Network
	nova_services  map[string]nova_services.Service
	ports          map[string]ports.Port
	projects       map[string]projects.Project
	servers        map[string]servers.Server
	services       map[string]services.Service
	stacks         map[string]stacks.ListedStack
	storagePools   map[string]schedulerstats.StoragePool
	subnets        map[string]subnets.Subnet
	volumes        map[string]volume
	gather_servers bool

	Log telegraf.Logger `toml:"-"`

	tls.ClientConfig
}

// ContainsService indicates whether a particular service is enabled
func (o *OpenStack) ContainsService(t serviceType) bool {
	for _, service := range o.services {
		if service.Type == string(t) {
			return true
		}
	}
	return false
}

// Description returns a description string of the input plugin and implements
// the Input interface.
func (o *OpenStack) Description() string {
	return "Collects performance metrics from OpenStack services"
}

// sampleConfig is a sample configuration file entry.
var sampleConfig = `
  ## The recommended interval to poll is '30m'

  ## The identity endpoint to authenticate against and get the service catalog from.
  authentication_endpoint = "https://my.openstack.cloud:5000"

  ## The domain to authenticate against when using a V3 identity endpoint. Defaults to 'default'.
  # domain = "default"

  ## The project to authenticate as. Defaults to 'admin'.
  # project = "admin"

  ## User authentication credentials. Must have admin rights. username defaults to 'admin'.
  # username = "admin"
  password = "password"

  ## Available services are: 
  ## "agents", "aggregates", "flavors", "hypervisors", "networks", "nova_services",
  ## "ports", "projects", "servers", "services", "stacks", "storage_pools", "subnets", "volumes"
  # enabled_services = ["services", "projects", "hypervisors", "flavors", "networks", "volumes"]

  ## Collect Server Diagnostics
  # server_diagnotics = false

  ## Amount of time allowed to complete the HTTP(s) request. Defaults to '15s'.
  # timeout = "15s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

// SampleConfig return a sample configuration file for auto-generation and
// implements the Input interface.
func (o *OpenStack) SampleConfig() string {
	return sampleConfig
}

// initialize performs any necessary initialization functions
func (o *OpenStack) Init() error {

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
		return fmt.Errorf("unable to create Newclient for OpenStack endpoint %v", err)
	}

	tlsCfg, err := o.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	provider.HTTPClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(o.Timeout),
	}

	if err := openstack.Authenticate(provider, authOption); err != nil {
		return fmt.Errorf("unable to authenticate OpenStack user %v", err)
	}

	// Create required clients and attach to the OpenStack struct
	if o.identity, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V3 identity client %v", err)
	}

	if err := o.gatherServices(); err != nil {
		o.Log.Warnf("failed to get resource openstack services %v", err)
	}

	if o.compute, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 compute client %v", err)
	}

	// Create required clients and attach to the OpenStack struct
	if o.network, err = openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{}); err != nil {
		return fmt.Errorf("unable to create V2 network client %v", err)
	}

	// The Orchestration service is optional
	if o.ContainsService(Orchestration) {
		if o.stack, err = openstack.NewOrchestrationV1(provider, gophercloud.EndpointOpts{}); err != nil {
			return fmt.Errorf("unable to create V1 stack client %v", err)
		}
	}

	// The Cinder volume storage service is optional
	if o.ContainsService(volumeV2Service) {
		if o.volume, err = openstack.NewBlockStorageV2(provider, gophercloud.EndpointOpts{}); err != nil {
			return fmt.Errorf("unable to create V2 volume client %v", err)
		}
	}

	o.gather_servers = false

	return nil
}

// Gather gathers resources from the OpenStack API and accumulates metrics.  This
// implements the Input interface.
func (o *OpenStack) Gather(acc telegraf.Accumulator) error {

	// Gather resources.  Note service harvesting must come first as the other
	// gatherers are dependant on this information.
	gatherers := map[string]func() error{
		"projects":      o.gatherProjects,
		"hypervisors":   o.gatherHypervisors,
		"flavors":       o.gatherFlavors,
		"servers":       o.gatherServers,
		"volumes":       o.gatherVolumes,
		"storage_pools": o.gatherStoragePools,
		"subnets":       o.gatherSubnets,
		"ports":         o.gatherPorts,
		"networks":      o.gatherNetworks,
		"aggregates":    o.gatherAggregates,
		"nova_services": o.gatherNovaServices,
		"agents":        o.gatherAgents,
		"stacks":        o.gatherStacks,
	}

	for _, service := range o.EnabledServices {
		if service != "services" {
			gatherer := gatherers[service]
			if err := gatherer(); err != nil {
				o.Log.Warnf("failed to get resource %q %v", service, err)
			}
		}
	}

	// Accumulate statistics
	accumulators := map[string]func(telegraf.Accumulator){
		"services":      o.accumulateServices,
		"projects":      o.accumulateProjects,
		"hypervisors":   o.accumulateHypervisors,
		"flavors":       o.accumulateFlavors,
		"servers":       o.accumulateServers,
		"volumes":       o.accumulateVolumes,
		"storage_pools": o.accumulateStoragePools,
		"subnets":       o.accumulateSubnets,
		"ports":         o.accumulatePorts,
		"networks":      o.accumulateNetworks,
		"aggregates":    o.accumulateAggregates,
		"nova_services": o.accumulateNovaServices,
		"agents":        o.accumulateAgents,
		"stacks":        o.accumulateStacks,
	}

	for _, service := range o.EnabledServices {
		accumulator := accumulators[service]
		accumulator(acc)
	}

	if o.ServerDiagnotics && !o.gather_servers {
		if err := o.gatherServers(); err != nil {
			o.Log.Warnf("failed to get resource servers %v", err)
		} else {
			o.accumulateServerDiagnostics(acc)
		}
	} else if o.ServerDiagnotics {
		o.accumulateServerDiagnostics(acc)
	}

	return nil
}

// gatherStacks collects stacks from the OpenStack API.
func (o *OpenStack) gatherStacks() error {
	page, err := stacks.List(o.stack, &stacks.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list stacks %v", err)
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
		return fmt.Errorf("unable to list services %v", err)
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
		return fmt.Errorf("unable to list nova_services %v", err)
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
		return fmt.Errorf("unable to list subnets %v", err)
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
		return fmt.Errorf("unable to list ports %v", err)
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
		return fmt.Errorf("unable to list networks %v", err)
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
		return fmt.Errorf("unable to list newtron agents %v", err)
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
		return fmt.Errorf("unable to list aggregates %v", err)
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
		return fmt.Errorf("unable to list projects %v", err)
	}
	projects, err := projects.ExtractProjects(page)
	if err != nil {
		return fmt.Errorf("unable to extract projects %v", err)
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
		return fmt.Errorf("unable to list hypervisors %v", err)
	}
	hypervisors, err := hypervisors.ExtractHypervisors(page)
	if err != nil {
		return fmt.Errorf("unable to extract hypervisors %v", err)
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
		return fmt.Errorf("unable to list flavors %v", err)
	}
	flavors, err := flavors.ExtractFlavors(page)
	if err != nil {
		return fmt.Errorf("unable to extract flavors %v", err)
	}
	for _, flavor := range flavors {
		o.flavors[flavor.ID] = flavor
	}
	return nil
}

// gatherServers collects servers from the OpenStack API.
func (o *OpenStack) gatherServers() error {
	o.gather_servers = true
	page, err := servers.List(o.compute, &servers.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list servers %v", err)
	}
	servers, err := servers.ExtractServers(page)
	if err != nil {
		return fmt.Errorf("unable to extract servers %v", err)
	}
	for _, server := range servers {
		o.servers[server.ID] = server
		if o.ServerDiagnotics && server.Status == "ACTIVE" {
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
	page, err := volumes.List(o.volume, &volumes.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list volumes %v", err)
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
	results, err := schedulerstats.List(o.volume, &schedulerstats.ListOpts{Detail: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list storage pools %v", err)
	}
	storagePools, err := schedulerstats.ExtractStoragePools(results)
	if err != nil {
		return fmt.Errorf("unable to extract storage pools %v", err)
	}
	for _, storagePool := range storagePools {
		o.storagePools[storagePool.Name] = storagePool
	}
	return nil
}

// accumulateServerDiagnostics accumulates statistics from the compute(nova) service.
// currently only supports 'libvirt' driver.
func (o *OpenStack) accumulateServerDiagnostics(acc telegraf.Accumulator) {

	var type_port = regexp.MustCompile(`_rx$|_rx_drop$|_rx_errors$|_rx_packets$|_tx$|_tx_drop$|_tx_errors$|_tx_packets$`)
	var type_cpu = regexp.MustCompile(`cpu[0-9]{1,2}_time$`)

	for server_id, diagnostic := range o.diag {

		tags := map[string]string{
			"server_id": server_id,
		}
		fields := map[string]interface{}{}

		port_name := make(map[string]bool)

		// for metrics other than port/interface or cpu
		other_metrics := make(map[string]interface{})
		s, ok := diagnostic.(map[string]interface{})
		if !ok {
			o.Log.Warnf("unknown type for diagnostics %T", diagnostic)
			continue
		}

		for k, v := range s {
			if type_port.MatchString(k) {
				port_name[strings.Split(k, "_")[0]] = true
			} else if type_cpu.MatchString(k) {
				fields[k] = v
			} else {
				other_metrics[k] = v
			}
		}

		fields["hdd_errors"] = other_metrics["hdd_errors"]
		fields["hdd_read"] = other_metrics["hdd_read"]
		fields["hdd_read_req"] = other_metrics["hdd_read_req"]
		fields["hdd_write"] = other_metrics["hdd_write"]
		fields["hdd_write_req"] = other_metrics["hdd_write_req"]
		fields["memory"] = other_metrics["memory"]
		fields["memory-actual"] = other_metrics["memory-actual"]
		fields["memory-rss"] = other_metrics["memory-rss"]
		fields["memory-swap_in"] = other_metrics["memory-swap_in"]
		fields["vda_errors"] = other_metrics["vda_errors"]
		fields["vda_read"] = other_metrics["vda_read"]
		fields["vda_read_req"] = other_metrics["vda_read_req"]
		fields["vda_write"] = other_metrics["vda_write"]
		fields["vda_write_req"] = other_metrics["vda_write_req"]
		tags["no_of_ports"] = strconv.Itoa(len(port_name))

		if len(port_name) > 0 {
			for key := range port_name {
				fields["port_rx"] = s[key+"_rx"]
				fields["port_rx_drop"] = s[key+"_rx_drop"]
				fields["port_rx_errors"] = s[key+"_rx_errors"]
				fields["port_rx_packets"] = s[key+"_rx_packets"]
				fields["port_tx"] = s[key+"_tx"]
				fields["port_tx_drop"] = s[key+"_tx_drop"]
				fields["port_tx_errors"] = s[key+"_tx_errors"]
				fields["port_tx_packets"] = s[key+"_tx_packets"]
				tags["port_name"] = key
				acc.AddFields("openstack_server_diagnostics", fields, tags)
			}
		} else {
			acc.AddFields("openstack_server_diagnostics", fields, tags)
		}
	}
}

// accumulateStacks accumulates statistics from the stack service.
func (o *OpenStack) accumulateStacks(acc telegraf.Accumulator) {
	for _, stack := range o.stacks {
		tags := map[string]string{
			"creation_time": stack.CreationTime.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"description":   stack.Description,
			"id":            stack.ID,
			"name":          stack.Name,
			"stack_tags":    strings.Join(stack.Tags[:], ","),
			"updated_time":  stack.UpdatedTime.Format("2006-01-02T15:04:05.999999999Z07:00"),
		}
		fields := map[string]interface{}{
			"status":        stack.Status,
			"status_reason": stack.StatusReason,
		}
		acc.AddFields("openstack_stack", fields, tags)
	}
}

// accumulateFlavors accumulates statistics from the flavor service.
func (o *OpenStack) accumulateFlavors(acc telegraf.Accumulator) {
	for _, flavor := range o.flavors {
		tags := map[string]string{
			"id":        flavor.ID,
			"name":      flavor.Name,
			"is_public": strconv.FormatBool(flavor.IsPublic),
		}
		fields := map[string]interface{}{
			"disk":        flavor.Disk,
			"ram":         flavor.RAM,
			"rxtx_factor": flavor.RxTxFactor,
			"swap":        flavor.Swap,
			"vcpus":       flavor.VCPUs,
			"ephemeral":   flavor.Ephemeral,
		}
		acc.AddFields("openstack_flavor", fields, tags)
	}
}

// accumulateProjects accumulates statistics from the identity service.
func (o *OpenStack) accumulateProjects(acc telegraf.Accumulator) {
	for _, project := range o.projects {
		tags := map[string]string{
			"description":  project.Description,
			"domain_id":    project.DomainID,
			"id":           project.ID,
			"name":         project.Name,
			"parent_id":    project.ParentID,
			"project_tags": strings.Join(project.Tags[:], ","),
		}
		fields := map[string]interface{}{
			"is_domain": project.IsDomain,
			"enabled":   project.Enabled,
			"projects":  len(o.projects),
		}
		acc.AddFields("openstack_identity", fields, tags)
	}
}

// accumulateHypervisors accumulates statistics from hypervisors.
func (o *OpenStack) accumulateHypervisors(acc telegraf.Accumulator) {
	for _, hypervisor := range o.hypervisors {
		tags := map[string]string{
			"cpu_vendor":              hypervisor.CPUInfo.Vendor,
			"cpu_arch":                hypervisor.CPUInfo.Arch,
			"cpu_model":               hypervisor.CPUInfo.Model,
			"cpu_features":            strings.Join(hypervisor.CPUInfo.Features[:], ","),
			"status":                  hypervisor.Status,
			"state":                   hypervisor.State,
			"host_ip":                 hypervisor.HostIP,
			"hypervisor_hostname":     hypervisor.HypervisorHostname,
			"hypervisor_type":         hypervisor.HypervisorType,
			"hypervisor_version":      strconv.Itoa(hypervisor.HypervisorVersion),
			"id":                      hypervisor.ID,
			"service_host":            hypervisor.Service.Host,
			"service_id":              hypervisor.Service.ID,
			"service_disabled_reason": hypervisor.Service.DisabledReason,
		}
		fields := map[string]interface{}{

			"cpu_topology_sockets": hypervisor.CPUInfo.Topology.Sockets,
			"cpu_topology_cores":   hypervisor.CPUInfo.Topology.Cores,
			"cpu_topology_threads": hypervisor.CPUInfo.Topology.Threads,
			"current_workload":     hypervisor.CurrentWorkload,
			"disk_available_least": hypervisor.DiskAvailableLeast,
			"free_disk_gb":         hypervisor.FreeDiskGB,
			"free_ram_mb":          hypervisor.FreeRamMB,
			"local_gb":             hypervisor.LocalGB,
			"local_gb_used":        hypervisor.LocalGBUsed,
			"memory_mb":            hypervisor.MemoryMB,
			"memory_mb_used":       hypervisor.MemoryMBUsed,
			"running_vms":          hypervisor.RunningVMs,
			"vcpus":                hypervisor.VCPUs,
			"vcpus_used":           hypervisor.VCPUsUsed,
		}
		acc.AddFields("openstack_hypervisor", fields, tags)
	}
}

// accumulateServers accumulates statistics about servers.
func (o *OpenStack) accumulateServers(acc telegraf.Accumulator) {
	var compute_hosts = map[string]map[string]string{}
	// establishing relation between server host_id and hypervisor host_name.
	for _, project := range o.projects {
		compute_hosts[project.ID] = map[string]string{}
		for _, hypervisor := range o.hypervisors {
			h := sha256.New224()
			h.Write([]byte(string(project.ID) + string(hypervisor.HypervisorHostname)))
			compute_hosts[project.ID][hex.EncodeToString(h.Sum(nil))] = hypervisor.HypervisorHostname
		}
	}
	for _, server := range o.servers {

		tags := map[string]string{}

		// Extract the flavor details to avoid joins (ignore errors and leave as zero values)
		var vcpus, ram, disk int
		if flavorIDInterface, ok := server.Flavor["id"]; ok {
			if flavorID, ok := flavorIDInterface.(string); ok {
				tags["flavor"] = flavorID
				if flavor, ok := o.flavors[flavorID]; ok {
					vcpus = flavor.VCPUs
					ram = flavor.RAM
					disk = flavor.Disk
				}
			}
		}

		if imageIDInterface, ok := server.Image["id"]; ok {
			if imageID, ok := imageIDInterface.(string); ok {
				tags["image"] = imageID
			}
		}

		// Try derive the associated project
		project := "unknown"
		if p, ok := o.projects[server.TenantID]; ok {
			project = p.Name
		}

		host_name := "unknown"
		for k, v := range compute_hosts[server.TenantID] {
			if k == server.HostID {
				host_name = v
			}
		}

		tags["id"] = server.ID
		tags["tenant_id"] = server.TenantID
		tags["user_id"] = server.UserID
		tags["name"] = server.Name
		tags["updated"] = server.Updated.Format("2006-01-02T15:04:05.999999999Z07:00")
		tags["created"] = server.Created.Format("2006-01-02T15:04:05.999999999Z07:00")
		tags["host_id"] = server.HostID
		tags["status"] = strings.ToLower(server.Status)
		tags["key_name"] = server.KeyName
		tags["fault_created"] = server.Fault.Created.Format("2006-01-02T15:04:05.999999999Z07:00")
		tags["fault_message"] = server.Fault.Message
		tags["host_name"] = host_name
		tags["project"] = project

		fields := map[string]interface{}{
			"progress":         server.Progress,
			"accessIPv4":       server.AccessIPv4,
			"accessIPv6":       server.AccessIPv6,
			"addresses":        len(server.Addresses),
			"adminPass":        server.AdminPass,
			"security_groups":  len(server.SecurityGroups),
			"volumes_attached": len(server.AttachedVolumes),
			"fault_code":       server.Fault.Code,
			"fault_details":    server.Fault.Details,
			"vcpus":            vcpus,
			"ram_mb":           ram,
			"disk_gb":          disk,
		}
		if len(server.AttachedVolumes) > 0 {
			for _, AttachedVolume := range server.AttachedVolumes {
				fields["volume_id"] = AttachedVolume.ID
				acc.AddFields("openstack_server", fields, tags)
			}
		} else {
			acc.AddFields("openstack_server", fields, tags)
		}
	}
}

// accumulateVolumes accumulates statistics about volumes.
func (o *OpenStack) accumulateVolumes(acc telegraf.Accumulator) {
	for _, volume := range o.volumes {

		tags := map[string]string{
			"id":                   volume.ID,
			"status":               volume.Status,
			"availability_zone":    volume.AvailabilityZone,
			"created_at":           volume.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"updated_at":           volume.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"name":                 volume.Name,
			"description":          volume.Description,
			"volume_type":          volume.VolumeType,
			"snapshot_id":          volume.SnapshotID,
			"source_volid":         volume.SourceVolID,
			"user_id":              volume.UserID,
			"bootable":             volume.Bootable,
			"replication_status":   volume.ReplicationStatus,
			"consistency_group_id": volume.ConsistencyGroupID,
		}

		fields := map[string]interface{}{
			"size":              volume.Size,
			"total_attachments": len(volume.Attachments),
			"encrypted":         volume.Encrypted,
			"multiattach":       volume.Multiattach,
		}

		if len(volume.Attachments) > 0 {
			for _, attachment := range volume.Attachments {

				tags["attachment_attached_at"] = attachment.AttachedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
				tags["attachment_attachment_id"] = attachment.AttachmentID
				tags["attachment_device"] = attachment.Device
				tags["attachment_host_name"] = attachment.HostName

				fields["attachment_server_id"] = attachment.ServerID

				acc.AddFields("openstack_volume", fields, tags)
			}
		} else {
			acc.AddFields("openstack_volume", fields, tags)
		}
	}
}

// accumulateStoragePools accumulates statistics about storage pools.
func (o *OpenStack) accumulateStoragePools(acc telegraf.Accumulator) {
	for _, storagePool := range o.storagePools {
		tags := map[string]string{
			"name":                storagePool.Capabilities.VolumeBackendName,
			"driver_version":      storagePool.Capabilities.DriverVersion,
			"storage_protocol":    storagePool.Capabilities.StorageProtocol,
			"vendor_name":         storagePool.Capabilities.VendorName,
			"volume_backend_name": storagePool.Capabilities.VolumeBackendName,
		}
		fields := map[string]interface{}{
			"total_capacity_gb": storagePool.Capabilities.TotalCapacityGB,
			"free_capacity_gb":  storagePool.Capabilities.FreeCapacityGB,
		}
		acc.AddFields("openstack_storage_pool", fields, tags)
	}

}

// accumulateServices accumulates statistics from services.
func (o *OpenStack) accumulateServices(acc telegraf.Accumulator) {
	for _, service := range o.services {
		tags := map[string]string{
			"service_id": service.ID,
			"name":       service.Type,
		}
		fields := map[string]interface{}{
			"service_enabled": service.Enabled,
		}
		acc.AddFields("openstack_service", fields, tags)
	}
}

// accumulateSubnets accumulates statistics from subnets.
func (o *OpenStack) accumulateSubnets(acc telegraf.Accumulator) {
	for _, subnet := range o.subnets {

		var allocation_pools []string
		for _, pool := range subnet.AllocationPools {
			allocation_pools = append(allocation_pools, pool.Start+"-"+pool.End)
		}
		tags := map[string]string{
			"id":                subnet.ID,
			"network_id":        subnet.NetworkID,
			"name":              subnet.Name,
			"description":       subnet.Description,
			"ip_version":        strconv.Itoa(subnet.IPVersion),
			"cidr":              subnet.CIDR,
			"gateway_ip":        subnet.GatewayIP,
			"dns_nameservers":   strings.Join(subnet.DNSNameservers[:], ","),
			"allocation_pools":  strings.Join(allocation_pools[:], ","),
			"tenant_id":         subnet.TenantID,
			"project_id":        subnet.ProjectID,
			"ipv6_address_mode": subnet.IPv6AddressMode,
			"ipv6_ra_mode":      subnet.IPv6RAMode,
			"subnet_pool_id":    subnet.SubnetPoolID,
			"subnet_tags":       strings.Join(subnet.Tags[:], ","),
		}
		fields := map[string]interface{}{
			"dhcp_enabled": subnet.EnableDHCP,
		}
		acc.AddFields("openstack_subnet", fields, tags)
	}
}

// accumulateNetworks accumulates statistics from networks.
func (o *OpenStack) accumulateNetworks(acc telegraf.Accumulator) {
	for _, network := range o.networks {
		tags := map[string]string{
			"id":                      network.ID,
			"name":                    network.Name,
			"description":             network.Description,
			"status":                  strings.ToLower(network.Status),
			"tenant_id":               network.TenantID,
			"updated_at":              network.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"created_at":              network.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"project_id":              network.ProjectID,
			"availability_zone_hints": strings.Join(network.AvailabilityZoneHints[:], ","),
			"network_tags":            strings.Join(network.Tags[:], ","),
		}
		fields := map[string]interface{}{
			"admin_state_up": network.AdminStateUp,
			"subnets":        len(network.Subnets),
			"shared":         network.Shared,
		}

		if len(network.Subnets) > 0 {
			for _, subnet := range network.Subnets {
				fields["subnet_id"] = subnet
				acc.AddFields("openstack_network", fields, tags)
			}
		} else {
			acc.AddFields("openstack_network", fields, tags)
		}
	}
}

// accumulatePorts accumulates statistics from ports.
func (o *OpenStack) accumulatePorts(acc telegraf.Accumulator) {
	for _, port := range o.ports {
		tags := map[string]string{
			"id":              port.ID,
			"network_id":      port.NetworkID,
			"name":            port.Name,
			"description":     port.Description,
			"status":          port.Status,
			"mac_address":     port.MACAddress,
			"tenant_id":       port.TenantID,
			"project_id":      port.ProjectID,
			"device_owner":    port.DeviceOwner,
			"security_groups": strings.Join(port.SecurityGroups[:], ","),
			"device_id":       port.DeviceID,
			"port_tags":       strings.Join(port.Tags[:], ","),
		}
		fields := map[string]interface{}{
			"admin_state_up":        port.AdminStateUp,
			"fixed_ips":             len(port.FixedIPs),
			"allowed_address_pairs": len(port.AllowedAddressPairs),
		}
		if len(port.FixedIPs) > 0 {
			for _, ip := range port.FixedIPs {
				fields["subnet_id"] = ip.SubnetID
				fields["ip_address"] = ip.IPAddress
				acc.AddFields("openstack_port", fields, tags)
			}
		} else {
			acc.AddFields("openstack_port", fields, tags)
		}

	}
}

// accumulateAggregates accumulates statistics from aggregates.
func (o *OpenStack) accumulateAggregates(acc telegraf.Accumulator) {
	for _, aggregate := range o.aggregates {
		tags := map[string]string{

			"availability_zone": aggregate.AvailabilityZone,
			"id":                strconv.Itoa(aggregate.ID),
			"name":              aggregate.Name,
			"created_at":        aggregate.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"updated_at":        aggregate.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"deleted_at":        aggregate.DeletedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		}
		fields := map[string]interface{}{
			"aggregate_hosts": len(aggregate.Hosts),
			"deleted":         aggregate.Deleted,
		}
		if len(aggregate.Hosts) > 0 {
			for _, host := range aggregate.Hosts {
				fields["aggregate_host"] = host
				acc.AddFields("openstack_aggregate", fields, tags)
			}
		} else {
			acc.AddFields("openstack_aggregate", fields, tags)
		}
	}
}

// accumulateNovaServices accumulates statistics from nova_services.
func (o *OpenStack) accumulateNovaServices(acc telegraf.Accumulator) {
	for _, nova_service := range o.nova_services {
		tags := map[string]string{
			"name":            nova_service.Binary,
			"disabled_reason": nova_service.DisabledReason,
			"host_machine":    nova_service.Host,
			"id":              nova_service.ID,
			"state":           nova_service.State,
			"status":          nova_service.Status,
			"updated_at":      nova_service.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"zone":            nova_service.Zone,
		}
		fields := map[string]interface{}{
			"forced_down": nova_service.ForcedDown,
		}
		acc.AddFields("openstack_nova_service", fields, tags)
	}
}

// accumulateAgents accumulates statistics from agents.
func (o *OpenStack) accumulateAgents(acc telegraf.Accumulator) {
	for _, agent := range o.agents {
		tags := map[string]string{
			"id":                  agent.ID,
			"agent_type":          agent.AgentType,
			"availability_zone":   agent.AvailabilityZone,
			"binary":              agent.Binary,
			"created_at":          agent.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"started_at":          agent.StartedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"heartbeat_timestamp": agent.HeartbeatTimestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
			"description":         agent.Description,
			"agent_host":          agent.Host,
			"topic":               agent.Topic,
		}
		fields := map[string]interface{}{
			"admin_state_up":   agent.AdminStateUp,
			"alive":            agent.Alive,
			"resources_synced": agent.ResourcesSynced,
		}
		acc.AddFields("openstack_newtron_agent", fields, tags)
	}
}

// init registers a callback which creates a new OpenStack input instance.
func init() {
	inputs.Add("openstack", func() telegraf.Input {
		return &OpenStack{
			Domain:           "default",
			Project:          "admin",
			Username:         "admin",
			EnabledServices:  []string{"services", "projects", "hypervisors", "flavors", "networks", "volumes"},
			ServerDiagnotics: false,
			Timeout:          config.Duration(time.Second * 15),
			agents:           map[string]agents.Agent{},
			aggregates:       map[int]aggregates.Aggregate{},
			diag:             map[string]interface{}{},
			flavors:          map[string]flavors.Flavor{},
			hypervisors:      map[string]hypervisors.Hypervisor{},
			networks:         map[string]networks.Network{},
			nova_services:    map[string]nova_services.Service{},
			ports:            map[string]ports.Port{},
			projects:         map[string]projects.Project{},
			servers:          map[string]servers.Server{},
			services:         map[string]services.Service{},
			stacks:           map[string]stacks.ListedStack{},
			storagePools:     map[string]schedulerstats.StoragePool{},
			subnets:          map[string]subnets.Subnet{},
			volumes:          map[string]volume{},
		}
	})
}
