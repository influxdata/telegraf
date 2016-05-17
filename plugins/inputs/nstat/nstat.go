package nstat

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	zeroByte    = []byte("0")
	newLineByte = []byte("\n")
	colonByte   = []byte(":")
)

// default file paths
const (
	NET_NETSTAT = "/net/netstat"
	NET_SNMP    = "/net/snmp"
	NET_SNMP6   = "/net/snmp6"
	NET_PROC    = "/proc"
)

// env variable names
const (
	ENV_NETSTAT = "PROC_NET_NETSTAT"
	ENV_SNMP    = "PROC_NET_SNMP"
	ENV_SNMP6   = "PROC_NET_SNMP6"
	ENV_ROOT    = "PROC_ROOT"
)

type Nstat struct {
	ProcNetNetstat string `toml:"proc_net_netstat"`
	ProcNetSNMP    string `toml:"proc_net_snmp"`
	ProcNetSNMP6   string `toml:"proc_net_snmp6"`
	DumpZeros      bool   `toml:"dump_zeros"`
}

var sampleConfig = `
  ## file paths for proc files. If empty default paths will be used:
  ##    /proc/net/netstat, /proc/net/snmp, /proc/net/snmp6
  ## These can also be overridden with env variables, see README.
  proc_net_netstat = ""
  proc_net_snmp    = ""
  proc_net_snmp6   = ""
  ## dump metrics with 0 values too
  dump_zeros       = true
`

func (ns *Nstat) Description() string {
	return "Collect kernel snmp counters and network interface statistics"
}

func (ns *Nstat) SampleConfig() string {
	return sampleConfig
}

func (ns *Nstat) Gather(acc telegraf.Accumulator) error {
	// load paths, get from env if config values are empty
	ns.loadPaths()

	netstat, err := ioutil.ReadFile(ns.ProcNetNetstat)
	if err != nil {
		return err
	}

	// collect netstat data
	err = ns.gatherNetstat(netstat, acc)
	if err != nil {
		return err
	}

	// collect SNMP data
	snmp, err := ioutil.ReadFile(ns.ProcNetSNMP)
	if err != nil {
		return err
	}
	err = ns.gatherSNMP(snmp, acc)
	if err != nil {
		return err
	}

	// collect SNMP6 data
	snmp6, err := ioutil.ReadFile(ns.ProcNetSNMP6)
	if err != nil {
		return err
	}
	err = ns.gatherSNMP6(snmp6, acc)
	if err != nil {
		return err
	}
	return nil
}

func (ns *Nstat) gatherNetstat(data []byte, acc telegraf.Accumulator) error {
	metrics, err := loadUglyTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"name": "netstat",
	}
	acc.AddFields("nstat", metrics, tags)
	return nil
}

func (ns *Nstat) gatherSNMP(data []byte, acc telegraf.Accumulator) error {
	metrics, err := loadUglyTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"name": "snmp",
	}
	acc.AddFields("nstat", metrics, tags)
	return nil
}

func (ns *Nstat) gatherSNMP6(data []byte, acc telegraf.Accumulator) error {
	metrics, err := loadGoodTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"name": "snmp6",
	}
	acc.AddFields("nstat", metrics, tags)
	return nil
}

// loadPaths can be used to read paths firstly from config
// if it is empty then try read from env variables
func (ns *Nstat) loadPaths() {
	if ns.ProcNetNetstat == "" {
		ns.ProcNetNetstat = proc(ENV_NETSTAT, NET_NETSTAT)
	}
	if ns.ProcNetSNMP == "" {
		ns.ProcNetSNMP = proc(ENV_SNMP, NET_SNMP)
	}
	if ns.ProcNetSNMP6 == "" {
		ns.ProcNetSNMP = proc(ENV_SNMP6, NET_SNMP6)
	}
}

// loadGoodTable can be used to parse string heap that
// headers and values are arranged in right order
func loadGoodTable(table []byte, dumpZeros bool) (map[string]interface{}, error) {
	entries := map[string]interface{}{}
	fields := bytes.Fields(table)
	var value int64
	var err error
	// iterate over two values each time
	// first value is header, second is value
	for i := 0; i < len(fields); i = i + 2 {
		// counter is zero
		if bytes.Equal(fields[i+1], zeroByte) {
			if !dumpZeros {
				continue
			} else {
				entries[string(fields[i])] = int64(0)
				continue
			}
		}
		// the counter is not zero, so parse it.
		value, err = strconv.ParseInt(string(fields[i+1]), 10, 64)
		if err == nil {
			entries[string(fields[i])] = value
		}
	}
	return entries, nil
}

// loadUglyTable can be used to parse string heap that
// the headers and values are splitted with a newline
func loadUglyTable(table []byte, dumpZeros bool) (map[string]interface{}, error) {
	entries := map[string]interface{}{}
	// split the lines by newline
	lines := bytes.Split(table, newLineByte)
	var value int64
	var err error
	// iterate over lines, take 2 lines each time
	// first line contains header names
	// second line contains values
	for i := 0; i < len(lines); i = i + 2 {
		if len(lines[i]) == 0 {
			continue
		}
		headers := bytes.Fields(lines[i])
		prefix := bytes.TrimSuffix(headers[0], colonByte)
		metrics := bytes.Fields(lines[i+1])

		for j := 1; j < len(headers); j++ {
			// counter is zero
			if bytes.Equal(metrics[j], zeroByte) {
				if !dumpZeros {
					continue
				} else {
					entries[string(append(prefix, headers[j]...))] = int64(0)
					continue
				}
			}
			// the counter is not zero, so parse it.
			value, err = strconv.ParseInt(string(metrics[j]), 10, 64)
			if err == nil {
				entries[string(append(prefix, headers[j]...))] = value
			}
		}
	}
	return entries, nil
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// try to read root path, or use default root path
	root := os.Getenv(ENV_ROOT)
	if root == "" {
		root = NET_PROC
	}
	return root + path
}

func init() {
	inputs.Add("nstat", func() telegraf.Input {
		return &Nstat{}
	})
}
