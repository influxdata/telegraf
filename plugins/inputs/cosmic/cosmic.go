package cosmic

import (
	"github.com/MissionCriticalCloud/go-cosmic/cosmic"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Cosmic struct {
	Url       string
	Apikey    string
	Secretkey string
	Timeout   int64

	tls.ClientConfig

	Domainid string

	client *cosmic.CosmicClient
}

var sampleConfig = `
  ##
  ## Connection parameters
  ##

  ## Cosmic API endpoint
  url = "https://localhost/client/api"
  ## The API key to use for metrics collection
  apikey = "xxx"
  ## The corresponding secret key
  secretkey = "xxx"
  ## Timeout in seconds per http request to the Cosmic API endpoint
  timeout = 60

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ##
  ## Metric collection parameters
  ##

  ## The domain for which to collect metrics
  # domainid = "00000000-0000-0000-0000-000000000000"
`

func (c *Cosmic) Description() string {
	return "Gather metrics from resources available in Cosmic"
}

func (c *Cosmic) SampleConfig() string {
	return sampleConfig
}

func (c *Cosmic) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		// make new tls config
		tlsCfg, err := c.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		c.client = cosmic.NewClient(c.Url, c.Apikey, c.Secretkey, tlsCfg, c.Timeout)
	}

	c.GatherVirtualMachineMetrics(acc)
	c.GatherVolumeMetrics(acc)
	c.GatherPublicIPMetrics(acc)

	return nil
}

func (c *Cosmic) GatherVirtualMachineMetrics(acc telegraf.Accumulator) error {
	var p = cosmic.ListVirtualMachinesParams{}

	if c.Domainid != "" {
		p.SetDomainid(c.Domainid)
	}

	p.SetListall(true)

	var listVirtualMachinesResponse, error = c.client.VirtualMachine.ListVirtualMachines(&p)

	if error != nil {
		return error
	}

	c.ProcessVirtualMachineMetrics(acc, listVirtualMachinesResponse.VirtualMachines)

	return nil
}

func (c *Cosmic) ProcessVirtualMachineMetrics(acc telegraf.Accumulator, virtualmachines []*cosmic.VirtualMachine) {
	for _, virtualmachine := range virtualmachines {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		fields["cpunumber"] = virtualmachine.Cpunumber
		fields["memory"] = virtualmachine.Memory
		fields["state"] = virtualmachine.State
		fields["hostid"] = virtualmachine.Hostid
		fields["hostname"] = virtualmachine.Hostname
		fields["serviceofferingid"] = virtualmachine.Serviceofferingid
		fields["serviceofferingname"] = virtualmachine.Serviceofferingname

		tags["id"] = virtualmachine.Id
		tags["name"] = virtualmachine.Name
		tags["account"] = virtualmachine.Account
		tags["created"] = virtualmachine.Created
		tags["displayname"] = virtualmachine.Displayname
		tags["domain"] = virtualmachine.Domain
		tags["domainid"] = virtualmachine.Domainid
		tags["hypervisor"] = virtualmachine.Hypervisor
		tags["instancename"] = virtualmachine.Instancename
		tags["templateid"] = virtualmachine.Templateid
		tags["templatename"] = virtualmachine.Templatename
		tags["templatedisplaytext"] = virtualmachine.Templatedisplaytext
		tags["userid"] = virtualmachine.Userid
		tags["username"] = virtualmachine.Username
		tags["zoneid"] = virtualmachine.Zoneid
		tags["zonename"] = virtualmachine.Zonename

		acc.AddFields("cosmic_virtualmachine_metrics", fields, tags)
	}
}

func (c *Cosmic) GatherVolumeMetrics(acc telegraf.Accumulator) error {
	var p = cosmic.ListVolumesParams{}

	if c.Domainid != "" {
		p.SetDomainid(c.Domainid)
	}

	p.SetListall(true)

	var listVolumeResponse, error = c.client.Volume.ListVolumes(&p)

	if error != nil {
		return error
	}

	c.ProcessVolumeMetrics(acc, listVolumeResponse.Volumes)

	return nil
}

func (c *Cosmic) ProcessVolumeMetrics(acc telegraf.Accumulator, volumes []*cosmic.Volume) {
	for _, volume := range volumes {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		fields["size"] = volume.Size
		fields["state"] = volume.State
		fields["attached"] = volume.Attached
		fields["destroyed"] = volume.Destroyed
		fields["deviceid"] = volume.Deviceid
		fields["diskofferingdisplaytext"] = volume.Diskofferingdisplaytext
		fields["diskofferingid"] = volume.Diskofferingid
		fields["diskofferingname"] = volume.Diskofferingname
		fields["path"] = volume.Path
		fields["storage"] = volume.Storage
		fields["storageid"] = volume.Storageid
		fields["virtualmachineid"] = volume.Virtualmachineid
		fields["vmdisplayname"] = volume.Vmdisplayname
		fields["vmname"] = volume.Vmname
		fields["vmstate"] = volume.Vmstate

		tags["id"] = volume.Id
		tags["name"] = volume.Name
		tags["account"] = volume.Account
		tags["created"] = volume.Created
		tags["domain"] = volume.Domain
		tags["domainid"] = volume.Domainid
		tags["hypervisor"] = volume.Hypervisor
		tags["zoneid"] = volume.Zoneid
		tags["zonename"] = volume.Zonename

		acc.AddFields("cosmic_volume_metrics", fields, tags)
	}
}

func (c *Cosmic) GatherPublicIPMetrics(acc telegraf.Accumulator) error {
	var p = cosmic.ListPublicIpAddressesParams{}

	if c.Domainid != "" {
		p.SetDomainid(c.Domainid)
	}

	p.SetListall(true)

	var listPublicIpAddressesresponse, error = c.client.PublicIPAddress.ListPublicIpAddresses(&p)

	if error != nil {
		return error
	}

	c.ProcessPublicIPMetrics(acc, listPublicIpAddressesresponse.PublicIpAddresses)

	return nil
}

func (c *Cosmic) ProcessPublicIPMetrics(acc telegraf.Accumulator, publicIpAddresses []*cosmic.PublicIpAddress) {
	for _, publicIpAddress := range publicIpAddresses {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		fields["aclid"] = publicIpAddress.Aclid
		fields["state"] = publicIpAddress.State
		fields["vpcid"] = publicIpAddress.Vpcid

		tags["id"] = publicIpAddress.Id
		tags["account"] = publicIpAddress.Account
		tags["domain"] = publicIpAddress.Domain
		tags["domainid"] = publicIpAddress.Domainid
		tags["ipaddress"] = publicIpAddress.Ipaddress
		tags["zoneid"] = publicIpAddress.Zoneid
		tags["zonename"] = publicIpAddress.Zonename

		acc.AddFields("cosmic_publicipaddress_metrics", fields, tags)
	}
}

func init() {
	inputs.Add("cosmic", func() telegraf.Input {
		return &Cosmic{}
	})
}
