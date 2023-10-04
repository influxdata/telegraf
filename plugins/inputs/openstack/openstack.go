// Package openstack implements an OpenStack input plugin for Telegraf
//
// The OpenStack input plug is a simple two phase metric collector.  In the first
// pass a set of gatherers are run against the API to cache collections of resources.
// In the second phase the gathered resources are combined and emitted as metrics.
//
// No aggregation is performed by the input plugin, instead queries to InfluxDB should
// be used to gather global totals of things such as tag frequency.
//
//go:generate ../../../tools/readme_config_includer/generator
package openstack

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats"
	cinder_services "github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/services"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumetenants"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
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
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	typePort    = regexp.MustCompile(`_rx$|_rx_drop$|_rx_errors$|_rx_packets$|_tx$|_tx_drop$|_tx_errors$|_tx_packets$`)
	typeCPU     = regexp.MustCompile(`cpu[0-9]{1,2}_time$`)
	typeStorage = regexp.MustCompile(`_errors$|_read$|_read_req$|_write$|_write_req$`)
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
	OutputSecrets    bool            `toml:"output_secrets"`
	TagPrefix        string          `toml:"tag_prefix"`
	TagValue         string          `toml:"tag_value"`
	HumanReadableTS  bool            `toml:"human_readable_timestamps"`
	MeasureRequest   bool            `toml:"measure_openstack_requests"`
	Log              telegraf.Logger `toml:"-"`
	httpconfig.HTTPClientConfig

	// Locally cached clients
	identity *gophercloud.ServiceClient
	compute  *gophercloud.ServiceClient
	volume   *gophercloud.ServiceClient
	network  *gophercloud.ServiceClient
	stack    *gophercloud.ServiceClient

	// Locally cached resources
	openstackFlavors     map[string]flavors.Flavor
	openstackHypervisors []hypervisors.Hypervisor
	openstackProjects    map[string]projects.Project
	openstackServices    map[string]services.Service

	services map[string]bool
}

// convertTimeFormat, to convert time format based on HumanReadableTS
func (o *OpenStack) convertTimeFormat(t time.Time) interface{} {
	if o.HumanReadableTS {
		return t.Format("2006-01-02T15:04:05.999999999Z07:00")
	}
	return t.UnixNano()
}

func (*OpenStack) SampleConfig() string {
	return sampleConfig
}

// initialize performs any necessary initialization functions
func (o *OpenStack) Init() error {
	if len(o.EnabledServices) == 0 {
		o.EnabledServices = []string{"services", "projects", "hypervisors", "flavors", "networks", "volumes"}
	}
	sort.Strings(o.EnabledServices)
	if o.Username == "" || o.Password == "" {
		return fmt.Errorf("username or password can not be empty string")
	}
	if o.TagValue == "" {
		return fmt.Errorf("tag_value option can not be empty string")
	}

	// Check the enabled services
	o.services = make(map[string]bool, len(o.EnabledServices))
	for _, service := range o.EnabledServices {
		switch service {
		case "agents", "aggregates", "cinder_services", "flavors", "hypervisors",
			"networks", "nova_services", "ports", "projects", "servers", "services",
			"stacks", "storage_pools", "subnets", "volumes":
			o.services[service] = true
		default:
			return fmt.Errorf("invalid service %q", service)
		}
	}
	return nil
}

func (o *OpenStack) Start(_ telegraf.Accumulator) error {
	o.openstackFlavors = map[string]flavors.Flavor{}
	o.openstackHypervisors = []hypervisors.Hypervisor{}
	o.openstackProjects = map[string]projects.Project{}
	o.openstackServices = map[string]services.Service{}

	// Authenticate against Keystone and get a token provider
	provider, err := openstack.NewClient(o.IdentityEndpoint)
	if err != nil {
		return fmt.Errorf("unable to create client for OpenStack endpoint: %w", err)
	}

	ctx := context.Background()
	client, err := o.HTTPClientConfig.CreateClient(ctx, o.Log)
	if err != nil {
		return err
	}

	provider.HTTPClient = *client

	// Authenticate to the endpoint
	authOption := gophercloud.AuthOptions{
		IdentityEndpoint: o.IdentityEndpoint,
		DomainName:       o.Domain,
		TenantName:       o.Project,
		Username:         o.Username,
		Password:         o.Password,
		AllowReauth:      true,
	}
	if err := openstack.Authenticate(provider, authOption); err != nil {
		return fmt.Errorf("unable to authenticate OpenStack user: %w", err)
	}

	// Create required clients and attach to the OpenStack struct
	o.identity, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("unable to create V3 identity client: %w", err)
	}
	o.compute, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("unable to create V2 compute client: %w", err)
	}
	o.network, err = openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("unable to create V2 network client: %w", err)
	}

	// Determine the services available at the endpoint
	if err := o.availableServices(); err != nil {
		return fmt.Errorf("failed to get resource openstack services: %w", err)
	}

	// Setup the optional services
	var hasOrchestration bool
	var hasBlockStorage bool
	for _, available := range o.openstackServices {
		switch available.Type {
		case "orchestration":
			o.stack, err = openstack.NewOrchestrationV1(provider, gophercloud.EndpointOpts{})
			if err != nil {
				return fmt.Errorf("unable to create V1 stack client: %w", err)
			}
			hasOrchestration = true
		case "volumev3":
			o.volume, err = openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{})
			if err != nil {
				return fmt.Errorf("unable to create V3 volume client: %w", err)
			}
		}
	}

	// Check if we need to disable services that are enabled by the user
	if !hasOrchestration {
		if o.services["stacks"] {
			o.Log.Warn("Disabling \"stacks\" service because orchestration is not available at the endpoint!")
			delete(o.services, "stacks")
		}
	}
	if !hasBlockStorage {
		for _, s := range []string{"cinder_services", "storage_pools", "volumes"} {
			if o.services[s] {
				o.Log.Warnf("Disabling %q service because block-storage is not available at the endpoint!", s)
				delete(o.services, s)
			}
		}
	}

	return nil
}

func (o *OpenStack) Stop() {}

// Gather gathers resources from the OpenStack API and accumulates metrics.  This
// implements the Input interface.
func (o *OpenStack) Gather(acc telegraf.Accumulator) error {
	callDuration := make(map[string]interface{}, len(o.services))

	// Prepare the shared resources
	if o.services["hypervisors"] || o.services["servers"] || o.ServerDiagnotics {
		start := time.Now()
		if err := o.gatherHypervisors(); err != nil {
			acc.AddError(fmt.Errorf("failed to get resource \"hypervisors\": %w", err))
		}
		if o.services["hypervisors"] {
			callDuration["hypervisors"] = time.Since(start).Nanoseconds()
		}
	}
	// Servers were already queried, so use this information
	if o.services["servers"] || o.ServerDiagnotics {
		start := time.Now()
		if err := o.gatherServers(acc); err != nil {
			return fmt.Errorf("failed to get resource \"servers\": %w", err)
		}
		callDuration["servers"] = time.Since(start).Nanoseconds()
	}

	for service := range o.services {
		var err error

		start := time.Now()
		switch service {
		case "services":
			// As Services are already gathered in Init(), using this to accumulate them.
			o.accumulateServices(acc)
			continue
		case "projects":
			err = o.gatherProjects(acc)
		case "hypervisors":
			// Gathered as part of the shared resource
			o.accumulateHypervisor(acc)
			continue
		case "flavors":
			err = o.gatherFlavors(acc)
		case "servers":
			// Gathered as part of the shared resource
		case "volumes":
			err = o.gatherVolumes(acc)
		case "storage_pools":
			err = o.gatherStoragePools(acc)
		case "subnets":
			err = o.gatherSubnets(acc)
		case "ports":
			err = o.gatherPorts(acc)
		case "networks":
			err = o.gatherNetworks(acc)
		case "aggregates":
			err = o.gatherAggregates(acc)
		case "nova_services":
			err = o.gatherNovaServices(acc)
		case "cinder_services":
			err = o.gatherCinderServices(acc)
		case "agents":
			err = o.gatherAgents(acc)
		case "stacks":
			err = o.gatherStacks(acc)
		default:
			return fmt.Errorf("invalid service %q", service)
		}
		if err != nil {
			acc.AddError(fmt.Errorf("failed to get resource %q: %w", service, err))
		}
		callDuration[service] = time.Since(start).Nanoseconds()
	}

	if o.MeasureRequest {
		for service, duration := range callDuration {
			acc.AddFields("openstack_request_duration", map[string]interface{}{service: duration}, map[string]string{})
		}
	}

	return nil
}

// availableServices collects the available endpoint services via API
func (o *OpenStack) availableServices() error {
	page, err := services.List(o.identity, nil).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list services: %w", err)
	}
	extractedServices, err := services.ExtractServices(page)
	if err != nil {
		return fmt.Errorf("unable to extract services: %w", err)
	}
	for _, service := range extractedServices {
		o.openstackServices[service.ID] = service
	}

	return nil
}

// gatherStacks collects and accumulates stacks data from the OpenStack API.
func (o *OpenStack) gatherStacks(acc telegraf.Accumulator) error {
	page, err := stacks.List(o.stack, &stacks.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list stacks: %w", err)
	}
	extractedStacks, err := stacks.ExtractStacks(page)
	if err != nil {
		return fmt.Errorf("unable to extract stacks: %w", err)
	}
	for _, stack := range extractedStacks {
		tags := map[string]string{
			"description": stack.Description,
			"name":        stack.Name,
		}
		for _, stackTag := range stack.Tags {
			tags[o.TagPrefix+stackTag] = o.TagValue
		}
		fields := map[string]interface{}{
			"status":        strings.ToLower(stack.Status),
			"id":            stack.ID,
			"status_reason": stack.StatusReason,
			"creation_time": o.convertTimeFormat(stack.CreationTime),
			"updated_time":  o.convertTimeFormat(stack.UpdatedTime),
		}
		acc.AddFields("openstack_stack", fields, tags)
	}

	return nil
}

// gatherNovaServices collects and accumulates nova_services data from the OpenStack API.
func (o *OpenStack) gatherNovaServices(acc telegraf.Accumulator) error {
	page, err := nova_services.List(o.compute, &nova_services.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list nova_services: %w", err)
	}
	novaServices, err := nova_services.ExtractServices(page)
	if err != nil {
		return fmt.Errorf("unable to extract nova_services: %w", err)
	}
	for _, novaService := range novaServices {
		tags := map[string]string{
			"name":         novaService.Binary,
			"host_machine": novaService.Host,
			"state":        novaService.State,
			"status":       strings.ToLower(novaService.Status),
			"zone":         novaService.Zone,
		}
		fields := map[string]interface{}{
			"id":              novaService.ID,
			"disabled_reason": novaService.DisabledReason,
			"forced_down":     novaService.ForcedDown,
			"updated_at":      o.convertTimeFormat(novaService.UpdatedAt),
		}
		acc.AddFields("openstack_nova_service", fields, tags)
	}

	return nil
}

// gatherCinderServices collects and accumulates cinder_services data from the OpenStack API.
func (o *OpenStack) gatherCinderServices(acc telegraf.Accumulator) error {
	page, err := cinder_services.List(o.volume, &cinder_services.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list cinder_services: %w", err)
	}
	cinderServices, err := cinder_services.ExtractServices(page)
	if err != nil {
		return fmt.Errorf("unable to extract cinder_services: %w", err)
	}
	for _, cinderService := range cinderServices {
		tags := map[string]string{
			"name":         cinderService.Binary,
			"cluster":      cinderService.Cluster,
			"host_machine": cinderService.Host,
			"state":        cinderService.State,
			"status":       strings.ToLower(cinderService.Status),
			"zone":         cinderService.Zone,
		}
		fields := map[string]interface{}{
			"id":                 cinderService.ActiveBackendID,
			"disabled_reason":    cinderService.DisabledReason,
			"frozen":             cinderService.Frozen,
			"replication_status": cinderService.ReplicationStatus,
			"updated_at":         o.convertTimeFormat(cinderService.UpdatedAt),
		}
		acc.AddFields("openstack_cinder_service", fields, tags)
	}

	return nil
}

// gatherSubnets collects and accumulates subnets data from the OpenStack API.
func (o *OpenStack) gatherSubnets(acc telegraf.Accumulator) error {
	page, err := subnets.List(o.network, &subnets.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list subnets: %w", err)
	}
	extractedSubnets, err := subnets.ExtractSubnets(page)
	if err != nil {
		return fmt.Errorf("unable to extract subnets: %w", err)
	}
	for _, subnet := range extractedSubnets {
		var allocationPools []string
		for _, pool := range subnet.AllocationPools {
			allocationPools = append(allocationPools, pool.Start+"-"+pool.End)
		}
		tags := map[string]string{
			"network_id":        subnet.NetworkID,
			"name":              subnet.Name,
			"description":       subnet.Description,
			"ip_version":        strconv.Itoa(subnet.IPVersion),
			"cidr":              subnet.CIDR,
			"gateway_ip":        subnet.GatewayIP,
			"tenant_id":         subnet.TenantID,
			"project_id":        subnet.ProjectID,
			"ipv6_address_mode": subnet.IPv6AddressMode,
			"ipv6_ra_mode":      subnet.IPv6RAMode,
			"subnet_pool_id":    subnet.SubnetPoolID,
		}
		for _, subnetTag := range subnet.Tags {
			tags[o.TagPrefix+subnetTag] = o.TagValue
		}
		fields := map[string]interface{}{
			"id":               subnet.ID,
			"dhcp_enabled":     subnet.EnableDHCP,
			"dns_nameservers":  strings.Join(subnet.DNSNameservers[:], ","),
			"allocation_pools": strings.Join(allocationPools[:], ","),
		}
		acc.AddFields("openstack_subnet", fields, tags)
	}
	return nil
}

// gatherPorts collects and accumulates ports data from the OpenStack API.
func (o *OpenStack) gatherPorts(acc telegraf.Accumulator) error {
	page, err := ports.List(o.network, &ports.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list ports: %w", err)
	}
	extractedPorts, err := ports.ExtractPorts(page)
	if err != nil {
		return fmt.Errorf("unable to extract ports: %w", err)
	}
	for _, port := range extractedPorts {
		tags := map[string]string{
			"network_id":   port.NetworkID,
			"name":         port.Name,
			"description":  port.Description,
			"status":       strings.ToLower(port.Status),
			"tenant_id":    port.TenantID,
			"project_id":   port.ProjectID,
			"device_owner": port.DeviceOwner,
			"device_id":    port.DeviceID,
		}
		for _, portTag := range port.Tags {
			tags[o.TagPrefix+portTag] = o.TagValue
		}
		fields := map[string]interface{}{
			"id":                    port.ID,
			"mac_address":           port.MACAddress,
			"admin_state_up":        port.AdminStateUp,
			"fixed_ips":             len(port.FixedIPs),
			"allowed_address_pairs": len(port.AllowedAddressPairs),
			"security_groups":       strings.Join(port.SecurityGroups[:], ","),
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
	return nil
}

// gatherNetworks collects and accumulates networks data from the OpenStack API.
func (o *OpenStack) gatherNetworks(acc telegraf.Accumulator) error {
	page, err := networks.List(o.network, &networks.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list networks: %w", err)
	}
	extractedNetworks, err := networks.ExtractNetworks(page)
	if err != nil {
		return fmt.Errorf("unable to extract networks: %w", err)
	}
	for _, network := range extractedNetworks {
		tags := map[string]string{
			"name":        network.Name,
			"description": network.Description,
			"status":      strings.ToLower(network.Status),
			"tenant_id":   network.TenantID,
			"project_id":  network.ProjectID,
		}
		for _, networkTag := range network.Tags {
			tags[o.TagPrefix+networkTag] = o.TagValue
		}
		fields := map[string]interface{}{
			"id":                      network.ID,
			"admin_state_up":          network.AdminStateUp,
			"subnets":                 len(network.Subnets),
			"shared":                  network.Shared,
			"availability_zone_hints": strings.Join(network.AvailabilityZoneHints[:], ","),
			"updated_at":              o.convertTimeFormat(network.UpdatedAt),
			"created_at":              o.convertTimeFormat(network.CreatedAt),
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
	return nil
}

// gatherAgents collects and accumulates agents data from the OpenStack API.
func (o *OpenStack) gatherAgents(acc telegraf.Accumulator) error {
	page, err := agents.List(o.network, &agents.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list neutron agents: %w", err)
	}
	extractedAgents, err := agents.ExtractAgents(page)
	if err != nil {
		return fmt.Errorf("unable to extract neutron agents: %w", err)
	}
	for _, agent := range extractedAgents {
		tags := map[string]string{
			"agent_type":        agent.AgentType,
			"availability_zone": agent.AvailabilityZone,
			"binary":            agent.Binary,
			"description":       agent.Description,
			"agent_host":        agent.Host,
			"topic":             agent.Topic,
		}
		fields := map[string]interface{}{
			"id":                  agent.ID,
			"admin_state_up":      agent.AdminStateUp,
			"alive":               agent.Alive,
			"resources_synced":    agent.ResourcesSynced,
			"created_at":          o.convertTimeFormat(agent.CreatedAt),
			"started_at":          o.convertTimeFormat(agent.StartedAt),
			"heartbeat_timestamp": o.convertTimeFormat(agent.HeartbeatTimestamp),
		}
		acc.AddFields("openstack_neutron_agent", fields, tags)
	}
	return nil
}

// gatherAggregates collects and accumulates aggregates data from the OpenStack API.
func (o *OpenStack) gatherAggregates(acc telegraf.Accumulator) error {
	page, err := aggregates.List(o.compute).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list aggregates: %w", err)
	}
	extractedAggregates, err := aggregates.ExtractAggregates(page)
	if err != nil {
		return fmt.Errorf("unable to extract aggregates: %w", err)
	}
	for _, aggregate := range extractedAggregates {
		tags := map[string]string{
			"availability_zone": aggregate.AvailabilityZone,
			"name":              aggregate.Name,
		}
		fields := map[string]interface{}{
			"id":              aggregate.ID,
			"aggregate_hosts": len(aggregate.Hosts),
			"deleted":         aggregate.Deleted,
			"created_at":      o.convertTimeFormat(aggregate.CreatedAt),
			"updated_at":      o.convertTimeFormat(aggregate.UpdatedAt),
			"deleted_at":      o.convertTimeFormat(aggregate.DeletedAt),
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
	return nil
}

// gatherProjects collects and accumulates projects data from the OpenStack API.
func (o *OpenStack) gatherProjects(acc telegraf.Accumulator) error {
	page, err := projects.List(o.identity, &projects.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list projects: %w", err)
	}
	extractedProjects, err := projects.ExtractProjects(page)
	if err != nil {
		return fmt.Errorf("unable to extract projects: %w", err)
	}
	for _, project := range extractedProjects {
		o.openstackProjects[project.ID] = project
		tags := map[string]string{
			"description": project.Description,
			"domain_id":   project.DomainID,
			"name":        project.Name,
			"parent_id":   project.ParentID,
		}
		for _, projectTag := range project.Tags {
			tags[o.TagPrefix+projectTag] = o.TagValue
		}
		fields := map[string]interface{}{
			"id":        project.ID,
			"is_domain": project.IsDomain,
			"enabled":   project.Enabled,
			"projects":  len(extractedProjects),
		}
		acc.AddFields("openstack_identity", fields, tags)
	}
	return nil
}

// gatherHypervisors collects and accumulates hypervisors data from the OpenStack API.
func (o *OpenStack) gatherHypervisors() error {
	page, err := hypervisors.List(o.compute, hypervisors.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list hypervisors: %w", err)
	}
	extractedHypervisors, err := hypervisors.ExtractHypervisors(page)
	if err != nil {
		return fmt.Errorf("unable to extract hypervisors: %w", err)
	}
	o.openstackHypervisors = extractedHypervisors

	return nil
}

// gatherFlavors collects and accumulates flavors data from the OpenStack API.
func (o *OpenStack) gatherFlavors(acc telegraf.Accumulator) error {
	page, err := flavors.ListDetail(o.compute, &flavors.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list flavors: %w", err)
	}
	extractedflavors, err := flavors.ExtractFlavors(page)
	if err != nil {
		return fmt.Errorf("unable to extract flavors: %w", err)
	}
	for _, flavor := range extractedflavors {
		o.openstackFlavors[flavor.ID] = flavor
		tags := map[string]string{
			"name":      flavor.Name,
			"is_public": strconv.FormatBool(flavor.IsPublic),
		}
		fields := map[string]interface{}{
			"id":          flavor.ID,
			"disk":        flavor.Disk,
			"ram":         flavor.RAM,
			"rxtx_factor": flavor.RxTxFactor,
			"swap":        flavor.Swap,
			"vcpus":       flavor.VCPUs,
			"ephemeral":   flavor.Ephemeral,
		}
		acc.AddFields("openstack_flavor", fields, tags)
	}
	return nil
}

// gatherVolumes collects and accumulates volumes data from the OpenStack API.
func (o *OpenStack) gatherVolumes(acc telegraf.Accumulator) error {
	page, err := volumes.List(o.volume, &volumes.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list volumes: %w", err)
	}
	v := []volume{}
	if err := volumes.ExtractVolumesInto(page, &v); err != nil {
		return fmt.Errorf("unable to extract volumes: %w", err)
	}
	for _, volume := range v {
		tags := map[string]string{
			"status":               strings.ToLower(volume.Status),
			"availability_zone":    volume.AvailabilityZone,
			"name":                 volume.Name,
			"description":          volume.Description,
			"volume_type":          volume.VolumeType,
			"snapshot_id":          volume.SnapshotID,
			"source_volid":         volume.SourceVolID,
			"bootable":             volume.Bootable,
			"replication_status":   strings.ToLower(volume.ReplicationStatus),
			"consistency_group_id": volume.ConsistencyGroupID,
		}
		fields := map[string]interface{}{
			"id":                volume.ID,
			"size":              volume.Size,
			"total_attachments": len(volume.Attachments),
			"encrypted":         volume.Encrypted,
			"multiattach":       volume.Multiattach,
			"created_at":        o.convertTimeFormat(volume.CreatedAt),
			"updated_at":        o.convertTimeFormat(volume.UpdatedAt),
		}
		if o.OutputSecrets {
			tags["user_id"] = volume.UserID
		}
		if len(volume.Attachments) > 0 {
			for _, attachment := range volume.Attachments {
				if !o.HumanReadableTS {
					fields["attachment_attached_at"] = attachment.AttachedAt.UnixNano()
				} else {
					fields["attachment_attached_at"] = attachment.AttachedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
				}
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
	return nil
}

// gatherStoragePools collects and accumulates storage pools data from the OpenStack API.
func (o *OpenStack) gatherStoragePools(acc telegraf.Accumulator) error {
	results, err := schedulerstats.List(o.volume, &schedulerstats.ListOpts{Detail: true}).AllPages()
	if err != nil {
		return fmt.Errorf("unable to list storage pools: %w", err)
	}
	storagePools, err := schedulerstats.ExtractStoragePools(results)
	if err != nil {
		return fmt.Errorf("unable to extract storage pools: %w", err)
	}
	for _, storagePool := range storagePools {
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
	return nil
}

// gatherServers collects servers from the OpenStack API.
func (o *OpenStack) gatherServers(acc telegraf.Accumulator) error {
	for _, hypervisor := range o.openstackHypervisors {
		page, err := servers.List(o.compute, &servers.ListOpts{AllTenants: true, Host: hypervisor.HypervisorHostname}).AllPages()
		if err != nil {
			return fmt.Errorf("unable to list servers: %w", err)
		}
		extractedServers, err := servers.ExtractServers(page)
		if err != nil {
			return fmt.Errorf("unable to extract servers: %w", err)
		}
		for _, server := range extractedServers {
			if o.services["server"] {
				o.accumulateServer(acc, server, hypervisor.HypervisorHostname)
			}
			if o.ServerDiagnotics && server.Status == "ACTIVE" {
				diagnostic, err := diagnostics.Get(o.compute, server.ID).Extract()
				if err != nil {
					acc.AddError(fmt.Errorf("unable to get diagnostics for server %q: %w", server.ID, err))
					continue
				}
				o.accumulateServerDiagnostics(acc, hypervisor.HypervisorHostname, server.ID, diagnostic)
			}
		}
	}
	return nil
}

func (o *OpenStack) accumulateHypervisor(acc telegraf.Accumulator) {
	for _, hypervisor := range o.openstackHypervisors {
		tags := map[string]string{
			"cpu_vendor":              hypervisor.CPUInfo.Vendor,
			"cpu_arch":                hypervisor.CPUInfo.Arch,
			"cpu_model":               hypervisor.CPUInfo.Model,
			"status":                  strings.ToLower(hypervisor.Status),
			"state":                   hypervisor.State,
			"hypervisor_hostname":     hypervisor.HypervisorHostname,
			"hypervisor_type":         hypervisor.HypervisorType,
			"hypervisor_version":      strconv.Itoa(hypervisor.HypervisorVersion),
			"service_host":            hypervisor.Service.Host,
			"service_id":              hypervisor.Service.ID,
			"service_disabled_reason": hypervisor.Service.DisabledReason,
		}
		for _, cpuFeature := range hypervisor.CPUInfo.Features {
			tags["cpu_feature_"+cpuFeature] = "true"
		}
		fields := map[string]interface{}{
			"id":                   hypervisor.ID,
			"host_ip":              hypervisor.HostIP,
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

// accumulateServices accumulates statistics of services.
func (o *OpenStack) accumulateServices(acc telegraf.Accumulator) {
	for _, service := range o.openstackServices {
		tags := map[string]string{
			"name": service.Type,
		}
		fields := map[string]interface{}{
			"service_id":      service.ID,
			"service_enabled": service.Enabled,
		}
		acc.AddFields("openstack_service", fields, tags)
	}
}

// accumulateServer accumulates statistics of a server.
func (o *OpenStack) accumulateServer(acc telegraf.Accumulator, server servers.Server, hostName string) {
	tags := map[string]string{}
	// Extract the flavor details to avoid joins (ignore errors and leave as zero values)
	var vcpus, ram, disk int
	if flavorIDInterface, ok := server.Flavor["id"]; ok {
		if flavorID, ok := flavorIDInterface.(string); ok {
			tags["flavor"] = flavorID
			if flavor, ok := o.openstackFlavors[flavorID]; ok {
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
	if p, ok := o.openstackProjects[server.TenantID]; ok {
		project = p.Name
	}
	tags["tenant_id"] = server.TenantID
	tags["name"] = server.Name
	tags["host_id"] = server.HostID
	tags["status"] = strings.ToLower(server.Status)
	tags["key_name"] = server.KeyName
	tags["host_name"] = hostName
	tags["project"] = project
	fields := map[string]interface{}{
		"id":               server.ID,
		"progress":         server.Progress,
		"accessIPv4":       server.AccessIPv4,
		"accessIPv6":       server.AccessIPv6,
		"addresses":        len(server.Addresses),
		"security_groups":  len(server.SecurityGroups),
		"volumes_attached": len(server.AttachedVolumes),
		"fault_code":       server.Fault.Code,
		"fault_details":    server.Fault.Details,
		"fault_message":    server.Fault.Message,
		"vcpus":            vcpus,
		"ram_mb":           ram,
		"disk_gb":          disk,
		"fault_created":    o.convertTimeFormat(server.Fault.Created),
		"updated":          o.convertTimeFormat(server.Updated),
		"created":          o.convertTimeFormat(server.Created),
	}
	if o.OutputSecrets {
		tags["user_id"] = server.UserID
		fields["adminPass"] = server.AdminPass
	}
	if len(server.AttachedVolumes) == 0 {
		acc.AddFields("openstack_server", fields, tags)
	} else {
		for _, AttachedVolume := range server.AttachedVolumes {
			fields["volume_id"] = AttachedVolume.ID
			acc.AddFields("openstack_server", fields, tags)
		}
	}
}

// accumulateServerDiagnostics accumulates statistics from the compute(nova) service.
// currently only supports 'libvirt' driver.
func (o *OpenStack) accumulateServerDiagnostics(acc telegraf.Accumulator, _, serverID string, diagnostic map[string]interface{}) {
	tags := map[string]string{
		"server_id": serverID,
	}
	fields := map[string]interface{}{}
	portName := make(map[string]bool)
	storageName := make(map[string]bool)
	memoryStats := make(map[string]interface{})
	for k, v := range diagnostic {
		if typePort.MatchString(k) {
			portName[strings.Split(k, "_")[0]] = true
		} else if typeCPU.MatchString(k) {
			fields[k] = v
		} else if typeStorage.MatchString(k) {
			storageName[strings.Split(k, "_")[0]] = true
		} else {
			memoryStats[k] = v
		}
	}
	fields["memory"] = memoryStats["memory"]
	fields["memory-actual"] = memoryStats["memory-actual"]
	fields["memory-rss"] = memoryStats["memory-rss"]
	fields["memory-swap_in"] = memoryStats["memory-swap_in"]
	tags["no_of_ports"] = strconv.Itoa(len(portName))
	tags["no_of_disks"] = strconv.Itoa(len(storageName))
	for key := range storageName {
		fields["disk_errors"] = diagnostic[key+"_errors"]
		fields["disk_read"] = diagnostic[key+"_read"]
		fields["disk_read_req"] = diagnostic[key+"_read_req"]
		fields["disk_write"] = diagnostic[key+"_write"]
		fields["disk_write_req"] = diagnostic[key+"_write_req"]
		tags["disk_name"] = key
		acc.AddFields("openstack_server_diagnostics", fields, tags)
	}
	for key := range portName {
		fields["port_rx"] = diagnostic[key+"_rx"]
		fields["port_rx_drop"] = diagnostic[key+"_rx_drop"]
		fields["port_rx_errors"] = diagnostic[key+"_rx_errors"]
		fields["port_rx_packets"] = diagnostic[key+"_rx_packets"]
		fields["port_tx"] = diagnostic[key+"_tx"]
		fields["port_tx_drop"] = diagnostic[key+"_tx_drop"]
		fields["port_tx_errors"] = diagnostic[key+"_tx_errors"]
		fields["port_tx_packets"] = diagnostic[key+"_tx_packets"]
		tags["port_name"] = key
		acc.AddFields("openstack_server_diagnostics", fields, tags)
	}
}

// init registers a callback which creates a new OpenStack input instance.
func init() {
	inputs.Add("openstack", func() telegraf.Input {
		return &OpenStack{
			Domain:    "default",
			Project:   "admin",
			TagPrefix: "openstack_tag_",
			TagValue:  "true",
		}
	})
}
