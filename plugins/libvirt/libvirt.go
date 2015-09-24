package libvirt

import (
	lv "github.com/alexzorin/libvirt-go"
	"github.com/influxdb/telegraf/plugins"
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

func (l *Libvirt) Gather(acc plugins.Accumulator) error {
	connection, err := lv.NewVirConnectionReadOnly(l.Uri)
	if err != nil {
		return err
	}
	defer connection.CloseConnection()

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

func (m *Libvirt) gatherDomain(acc plugins.Accumulator, domain lv.VirDomain, tags map[string]string) error {
	domainInfo, err := domain.GetInfo()
	if err != nil {
		return err
	}

	acc.Add("cpu_time", domainInfo.GetCpuTime(), tags)
	acc.Add("max_mem", domainInfo.GetMaxMem(), tags)
	acc.Add("memory", domainInfo.GetMemory(), tags)
	acc.Add("nr_virt_cpu", domainInfo.GetNrVirtCpu(), tags)

	return nil
}

func init() {
	plugins.Add("libvirt", func() plugins.Plugin {
		return &Libvirt{}
	})
}
