package libvirt

import (
	"fmt"
	"github.com/antchfx/xquery/xml"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DATA_RETRIEVAL_WAIT = 0.3 //number of seconds between two data requests

//const LIBVIRT_URI = "qemu:///system"
//const LIBVIRT_URI = "test:///default"
//const LIBVIRT_URI = "test+tcp://127.0.0.1/root/test.xml"

type Libvirt struct {
	Libvirt_uri string
}

func (l *Libvirt) Description() string {
	return `VM metrics collector
# Metrics this collector generates:
# libvirt.vm.count - number of running VMs
# libvirt.vm.cpu.load - VM's current CPU load (can be higher than 100%)
# libvirt.vm.cpu.time - CPU time spent by VM
# libvirt.vm.disk.read.requests - number of total VM's read requests
# libvirt.vm.disk.read.bytes - number of total VM's read bytes
# libvirt.vm.disk.write.requests - number of total VM's write requests
# libvirt.vm.disk.write.bytes - number of total VM's write bytes
# libvirt.vm.disk.total.requests - number of total VM's read + write requests
# libvirt.vm.disk.total.bytes - number of total VM's read + write bytes
# libvirt.vm.disk.current.read.requests - number of VM's current read requests
# libvirt.vm.disk.current.read.bytes - number of VM's current read bytes
# libvirt.vm.disk.current.write.requests - number of VM's current write requests
# libvirt.vm.disk.current.write.bytes - number of VM's current write bytes
# libvirt.vm.disk.current.total.requests - number of VM's current read + write requests
# libvirt.vm.disk.current.total.bytes - number of VM's current read + write bytes
# libvirt.vm.memory - memory used by VM in kB
# libvirt.vm.max.memory - memory requested in VM's template in kB
# libvirt.vm.max.vcpus - number of CPU requested in VM's template
# libvirt.vm.network.rx - number of VM's received bytes via network
# libvirt.vm.network.tx - number of VM's transmitted bytes via network
# libvirt.vm.network.current.rx - VM's current network incoming bandwidth
# libvirt.vm.network.current.tx - VM's current network outcoming bandwidth
# libvirt.vm.cpustat.count - number of CPUs
# libvirt.vm.cpustat.cpu.mhz - CPU's MHz
# libvirt.vm.cpustat.cpu.cores - CPU's number of cores`
}

func (l *Libvirt) SampleConfig() string {
	return `## Metrics from:
  ## The libvirt Test driver is a per-process fake hypervisor driver,
  ## with a driver name of 'test'. The driver maintains all its state in memory.
  # libvirt_uri = "test:///default"

  ## Metrics from qemu
  # libvirt_uri = "qemu:///system"

  ## Metrics from test file in docker libvirt
  libvirt_uri = "test+tcp://127.0.0.1/root/test.xml"
`
}

func (l *Libvirt) Gather(acc telegraf.Accumulator) error {
	conn, err := libvirt.NewConnect(l.Libvirt_uri)
	if err != nil {
		acc.AddError(fmt.Errorf("Error while creating connection."))
		return err
	}

	doms := get_domains(conn, acc)

	get_VM_count(len(doms), acc)
	pids := get_pids(acc)

	for uuid, pid := range pids {
		load := get_cpu_load(pid, acc)

		dom, err := conn.LookupDomainByUUIDString(uuid)
		if err != nil {
			acc.AddError(fmt.Errorf("Error while looking up domain by UUID (string): %s\n", err))
			continue
		}

		name, err := dom.GetName()
		if err != nil {
			acc.AddError(fmt.Errorf("Error while getting name: %s\n", err))
			continue
		}

		tags := map[string]string{
			"deploy_id": name,
		}

		fields := map[string]interface{}{
			"load": load,
		}
		acc.AddFields("cpu", fields, tags)
	}

	for _, dom := range doms {
		get_cpu_info(dom, acc)
		get_disks_info(dom, acc)
		get_memory_info(dom, acc)
		get_network_info(dom, acc)

		dom.Free()
	}
	defer conn.Close()

	get_cpustat(acc)

	return nil
}

func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{}
	})
}

func get_domains(conn *libvirt.Connect, acc telegraf.Accumulator) []libvirt.Domain {
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	checkerr(acc, "Error while listing all domains", err)

	return doms
}

func get_VM_count(count int, acc telegraf.Accumulator) {
	tags := map[string]string{}

	fields := map[string]interface{}{
		"count": count,
	}

	acc.AddFields("", fields, tags)
}

func get_pids(acc telegraf.Accumulator) map[string]int {
	out, err := exec.Command("ps", "-ewwo", "pid,command").Output()
	if err != nil {
		acc.AddError(fmt.Errorf("Error while exec command:%s", err))
		return nil
	}

	regex, err := regexp.Compile("-uuid ([a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12})")
	if err != nil {
		acc.AddError(fmt.Errorf("Error while regexp compilation:%s", err))
		return nil
	}

	pids := map[string]int{}

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		match := regex.FindString(line)

		if match != "" {
			m := strings.TrimSpace(line)
			uuid := strings.Split(match, " ")[1]
			//fmt.Println("uuid ", uuid)
			pid := strings.Split(m, " ")[0]
			//fmt.Println("pid ", pid)

			pids[uuid], err = strconv.Atoi(pid)
			checkerr(acc, "Error while converting string pid to int", err)
		}
	}

	return pids
}

func get_cpu_load(pid int, acc telegraf.Accumulator) float64 {
	proc, err := process.NewProcess(int32(pid))
	checkerr(acc, "Error while getting process by pid", err)

	percent, err := proc.Percent((DATA_RETRIEVAL_WAIT * 1000) * time.Millisecond)
	checkerr(acc, "Error while getting cpu load", err)

	return percent
}

func get_cpu_info(dom libvirt.Domain, acc telegraf.Accumulator) {
	cpuinfo, err := dom.GetVcpus()
	if err != nil {
		acc.AddError(fmt.Errorf("Error while getting cpu info:%s", err))
		return
	}

	name, err := dom.GetName()
	checkerr(acc, "Error while getting name", err)

	for j := 0; j < len(cpuinfo); j++ {
		tags := map[string]string{
			"deploy_id": name,
		}

		cpu_time_fields := map[string]interface{}{
			"time": cpuinfo[j].CpuTime,
		}

		acc.AddFields("cpu", cpu_time_fields, tags, time.Now())
	}
}

func get_disks_info(dom libvirt.Domain, acc telegraf.Accumulator) {
	dxml, err := dom.GetXMLDesc(libvirt.DOMAIN_XML_SECURE)
	if err != nil {
		acc.AddError(fmt.Errorf("Error while getting xml to disk info:%s", err))
		return
	}

	name, err := dom.GetName()
	checkerr(acc, "Error while getting name", err)

	root, err := xmlquery.Parse(strings.NewReader(dxml))
	if err != nil {
		acc.AddError(fmt.Errorf("Error while parsing xml:%s", err))
		return
	}

	for _, n := range xmlquery.Find(root, "//domain/devices/disk/target") {
		path := n.SelectAttr("dev")

		dbs, err := dom.BlockStats(path)
		if err != nil {
			acc.AddError(fmt.Errorf("Error while getting disk info:%s", err))
			continue
		}

		time.Sleep((DATA_RETRIEVAL_WAIT * 1000) * time.Millisecond)
		dbs2, err := dom.BlockStats(path)
		if err != nil {
			acc.AddError(fmt.Errorf("Error while getting disk info:%s", err))
			continue
		}

		tags := map[string]string{
			"deploy_id": name,
			"device":    path,
		}

		disk_info_fields := map[string]interface{}{
			"read.request":   dbs.RdReq,
			"read.bytes":     dbs.RdBytes,
			"write.request":  dbs.WrReq,
			"write.bytes":    dbs.WrBytes,
			"total.requests": dbs.RdReq + dbs.WrReq,
			"total.bytes":    dbs.RdBytes + dbs.WrBytes,

			"current.read.request":  dataPerSecond(dbs.RdReq, dbs2.RdReq),
			"current.read.bytes":    dataPerSecond(dbs.RdBytes, dbs2.RdBytes),
			"current.write.request": dataPerSecond(dbs.WrReq, dbs2.WrReq),
			"current.write.bytes": dataPerSecond(dbs.RdReq, dbs2.RdReq) +
				dataPerSecond(dbs.WrReq, dbs2.WrReq),
			"current.total.requests": dataPerSecond(dbs.RdReq, dbs2.RdReq) +
				dataPerSecond(dbs.WrReq, dbs2.WrReq),
			"current.total.bytes": dataPerSecond(dbs.RdBytes, dbs2.RdBytes) +
				dataPerSecond(dbs.WrBytes, dbs2.WrBytes),
		}
		acc.AddFields("disk", disk_info_fields, tags, time.Now())
	}
}

func dataPerSecond(d1, d2 int64) float64 {
	return float64(d2-d1) / DATA_RETRIEVAL_WAIT
}

func get_memory_info(dom libvirt.Domain, acc telegraf.Accumulator) {
	domainInfo, err := dom.GetInfo()
	if err != nil {
		acc.AddError(fmt.Errorf("Error while getting memory info:%s", err))
		return
	}

	name, err := dom.GetName()
	checkerr(acc, "Error while getting name", err)

	tags := map[string]string{
		"deploy_id": name,
	}

	fields := map[string]interface{}{
		"memory": domainInfo.Memory,
	}
	acc.AddFields("", fields, tags, time.Now())

	fields_max := map[string]interface{}{
		"memory": domainInfo.MaxMem,
		"vcpus":  domainInfo.NrVirtCpu,
	}
	acc.AddFields("max", fields_max, tags, time.Now())
}

func get_network_info(dom libvirt.Domain, acc telegraf.Accumulator) {
	dxml, err := dom.GetXMLDesc(libvirt.DOMAIN_XML_SECURE)
	checkerr(acc, "Error while getting xml to network info", err)

	name, err := dom.GetName()
	checkerr(acc, "Error while getting name", err)

	root, err := xmlquery.Parse(strings.NewReader(dxml))
	checkerr(acc, "Error while parsing xml", err)

	for _, n := range xmlquery.Find(root, "//domain/devices/interface/target") {
		path := n.SelectAttr("dev")

		dis, err := dom.InterfaceStats(path)
		checkerr(acc, "Error while getting network info", err)
		time.Sleep((DATA_RETRIEVAL_WAIT * 1000) * time.Millisecond)
		dis2, err := dom.InterfaceStats(path)
		checkerr(acc, "Error while getting network info", err)

		tags := map[string]string{
			"deploy_id": name,
		}

		fields := map[string]interface{}{
			"rx":         dis.RxBytes,
			"tx":         dis.TxBytes,
			"current.rx": dataPerSecond(dis.RxBytes, dis2.RxBytes),
			"current.tx": dataPerSecond(dis.TxBytes, dis2.TxBytes),
		}
		acc.AddFields("network", fields, tags, time.Now())
	}
}

func get_cpustat(acc telegraf.Accumulator) {
	cpus, err := cpu.Info()
	checkerr(acc, "Error while getting cpu info", err)

	tags := map[string]string{}

	fields := map[string]interface{}{
		"count": len(cpus),
	}

	acc.AddFields("cpustat", fields, tags, time.Now())

	for _, c := range cpus {
		tags := map[string]string{
			"model":     c.ModelName,
			"processor": fmt.Sprint(c.CPU),
		}

		fields := map[string]interface{}{
			"cpu.cores": c.Cores,
			"cpu.mhz":   c.Mhz,
		}

		acc.AddFields("cpustat", fields, tags, time.Now())
	}
}

func checkerr(acc telegraf.Accumulator, text string, err error) {
	if err != nil {
		acc.AddError(fmt.Errorf("%s: %s\n", text, err))
	}
}
