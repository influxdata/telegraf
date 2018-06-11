package libvirt

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/digitalocean/go-libvirt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	gouuid "github.com/satori/go.uuid"
	"net"
	"time"
)

const (
	sampleConfig = `
  ## Libvirt plugin polls domain metric data
  libvirt_sock = "/var/run/libvirt/libvirt-sock"
`
)

// MockPlugin struct should be named the same as the Plugin
type LibVirt struct {
	client      *libvirt.Libvirt
	LibVirtSock string `toml:"libvirt_sock"`
}

// Description will appear directly above the plugin definition in the config file
func (l *LibVirt) Description() string {
	return `This is plugin for libvirt`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (l *LibVirt) SampleConfig() string {
	return sampleConfig
}

func init() {
	virt := &LibVirt{}
	inputs.Add("libvirt", func() telegraf.Input { return virt })
}

// getDomainINfo gets the domain dumpxml data. We do this because there is not a better way to get domain specific disks
// and interfaces for disk and network metrics
func (l *LibVirt) getDomainInfo(dom libvirt.Domain, acc telegraf.Accumulator) *etree.Element {
	rXML, err := l.client.DomainGetXMLDesc(dom, 0)

	if err != nil {
		acc.AddError(fmt.Errorf("error getting domain xml dump: %+v", err))
		return nil
	}

	doc := etree.NewDocument()
	err = doc.ReadFromString(rXML)

	if err != nil {
		acc.AddError(fmt.Errorf("error reading dumpxml data: %+v", err))
		return nil
	}

	return doc.SelectElement("domain")
}

// Find all the disks in the XML data this is useful when using ceph for example
func (l *LibVirt) getDomainDiskDevices(domainInfo *etree.Element, acc telegraf.Accumulator) []string {

	var diskDevices []string

	deviceElements := domainInfo.SelectElements("devices")

	for _, device := range deviceElements {
		if device.Tag == "devices" {
			for _, disk := range device.SelectElements("disk") {
				if disk.SelectElement("source") != nil {
					diskDevices = append(diskDevices, disk.SelectElement("target").SelectAttr("dev").Value)
				}
			}
		}
	}

	return diskDevices
}

// getDomainInterfaceDevices parses the domain network interfaces to gather metric data on
func (l *LibVirt) getDomainInterfaces(domainInfo *etree.Element, acc telegraf.Accumulator) []string {
	var ifaces []string

	ifaceElements := domainInfo.SelectElement("devices").SelectElements("interface")

	for _, iface := range ifaceElements {
		ifaces = append(ifaces, iface.SelectElement("target").SelectAttr("dev").Value)
	}

	return ifaces
}

// gatherDomainDiskData specific disk usage metrics
func (l *LibVirt) gatherDomainDiskData(dom libvirt.Domain, disks []string, acc telegraf.Accumulator) {

	for _, d := range disks {
		tags := make(map[string]string)
		fields := make(map[string]interface{})

		if blkParams, _, err := l.client.DomainBlockStatsFlags(dom, d, 8, 0); err == nil {
			for _, blk := range blkParams {
				fields[blk.Field] = blk.Value.Get()
			}
		} else {
			acc.AddError(fmt.Errorf("unable to get block device data for: %v: %v", d, err))
		}

		rAlloc, rCapacity, rPhysical, err := l.client.DomainGetBlockInfo(dom, d, 0)
		if err != nil {
			acc.AddError(fmt.Errorf("failed to get block info for %s on domain %s it may not be available. Error: %v", d, dom.Name, err))
			return
		}

		fields["allocation"] = int64(rAlloc)
		fields["capacity"] = int64(rCapacity)
		fields["physical"] = int64(rPhysical)
		tags["name"] = dom.Name
		tags["device"] = d

		id, err := gouuid.FromBytes(dom.UUID[:])

		if err != nil {
			acc.AddError(fmt.Errorf("unable to get propper uuid: %s", err))
		} else {
			tags["uuid"] = id.String()
		}

		acc.AddFields("libirt_dom_volumes", fields, tags)

	}
}

// gatherDomainInterfaceData gathers network interface specific data for the given domain
func (l *LibVirt) gatherDomainInterfaceData(dom libvirt.Domain, ifaces []string, acc telegraf.Accumulator) {
	for _, i := range ifaces {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		rxBytes, rxPackets, rxErrors, rxDrops, txBytes, txPackets, txErrors, txDrops, err := l.client.DomainInterfaceStats(dom, i)

		if err != nil {
			acc.AddError(fmt.Errorf("error processing interface stats for %+v: %+v", i, err))
			continue
		}

		fields["rx_bytes"] = int64(rxBytes)
		fields["rx_packets"] = int64(rxPackets)
		fields["rx_errors"] = int64(rxErrors)
		fields["rx_drops"] = int64(rxDrops)
		fields["tx_bytes"] = int64(txBytes)
		fields["tx_packets"] = int64(txPackets)
		fields["tx_errors"] = int64(txErrors)
		fields["tx_drops"] = int64(txDrops)
		tags["iface"] = i
		tags["domain"] = dom.Name

		id, err := gouuid.FromBytes(dom.UUID[:])

		if err != nil {
			acc.AddError(fmt.Errorf("unable to get propper uuid: %s", err))
		} else {
			tags["uuid"] = id.String()
		}

		acc.AddFields("libvirt_dom_interface", fields, tags)
	}
}

// gatherDomainData grabs perfEvents, memory, and CPU info for the domain
func (l *LibVirt) gatherDomainData(dom libvirt.Domain, acc telegraf.Accumulator) {

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	perfEvents, err := l.client.DomainGetPerfEvents(dom, 0)

	if err != nil {
		acc.AddError(err)
		return
	}

	for _, p := range perfEvents {
		fields[p.Field] = int64(p.Value.Get().(int32))
	}

	state, maxMem, currentMemory, cpuTotal, cpuTime, err := l.client.DomainGetInfo(dom)

	if err != nil {
		acc.AddError(err)
		return
	}

	cpuParams, _, err := l.client.DomainGetCPUStats(dom, 2, 0, uint32(cpuTotal), 0)

	if err != nil {
		acc.AddError(err)
		return
	}

	for _, c := range cpuParams {
		fields[c.Field] = int64(c.Value.Get().(uint64))
	}

	if err != nil {
		acc.AddError(err)
	}

	usedMem := maxMem / currentMemory

	fields["mem_used_percent"] = int64(usedMem)
	fields["state"] = int64(state)
	fields["max_memory"] = int64(maxMem)
	fields["total_memory"] = int64(maxMem)
	fields["cpu_count"] = int64(cpuTotal)
	fields["cpu_time"] = int64(cpuTime)
	tags["name"] = dom.Name

	id, err := gouuid.FromBytes(dom.UUID[:])

	if err != nil {
		acc.AddError(fmt.Errorf("unable to get propper uuid: %s", err))
	} else {
		tags["uuid"] = id.String()
	}

	acc.AddFields("libvirt_domain", fields, tags)
}

// Gather defines what data the plugin will gather.
func (l *LibVirt) Gather(acc telegraf.Accumulator) error {

	c, err := net.DialTimeout("unix", l.LibVirtSock, 2*time.Second)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to dial libvirt: %v", err))
	}

	defer c.Close()

	l.client = libvirt.New(c)
	if err := l.client.Connect(); err != nil {
		acc.AddError(fmt.Errorf("failed to connect: %v", err))
		return err
	}

	doms, err := l.client.Domains()

	if err != nil {
		acc.AddError(fmt.Errorf("failed to get domains: %v", err))
		return err
	}

	for _, d := range doms {
		// Get the domain dumpxml
		// It contains the data for the Domain specific disks and interfaces which are not fetchable with any GetDomain calls
		// in the DO libvirt-go lib.
		domainInfo := l.getDomainInfo(d, acc)
		disks := l.getDomainDiskDevices(domainInfo, acc)
		ifaces := l.getDomainInterfaces(domainInfo, acc)

		l.gatherDomainData(d, acc)
		l.gatherDomainDiskData(d, disks, acc)
		l.gatherDomainInterfaceData(d, ifaces, acc)
	}

	return nil
}
