package libvirt

import (
	lv "github.com/libvirt/libvirt-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  # specify a libvirt connection uri
  uri = "qemu:///system"
`

type Libvirt struct {
	Uri string
}

func (l *Libvirt) SampleConfig() string {
	return sampleConfig
}

func (l *Libvirt) Description() string {
	return "Read domain infos from a libvirt deamon"
}

func (l *Libvirt) Gather(acc telegraf.Accumulator) error {
	connection, err := lv.NewConnectReadOnly(l.Uri)
	if err != nil {
		return err
	}
	defer connection.Close()

	domains, err := connection.ListDomains()
	if err != nil {
		return err
	}

	for _, domainId := range domains {
		domain, err := connection.LookupDomainById(domainId)
		if err != nil {
			return err
		}

		domainName, _ := domain.GetName()
		tags := map[string]string{"domain": domainName}
		l.gatherDomain(acc, domain, tags)
	}

	return nil
}

func (m *Libvirt) gatherDomain(acc telegraf.Accumulator, domain *lv.Domain, tags map[string]string) error {
	domainInfo, err := domain.GetInfo()
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"cpu_time":    domainInfo.CpuTime,
		"max_mem":     domainInfo.MaxMem,
		"memory":      domainInfo.Memory,
		"nr_virt_cpu": uint64(domainInfo.NrVirtCpu),
	}

	acc.AddFields("libvirt", fields, tags)

	return nil
}

func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{}
	})
}
