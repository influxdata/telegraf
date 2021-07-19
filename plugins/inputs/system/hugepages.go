package system

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	newlineByte     = []byte("\n")
	colonByte       = []byte(":")
	kbPrecisionByte = []byte("kB")
)

const (
	// the path where statistics are kept per NUMA nodes
	numaNodePath = "/sys/devices/system/node"
	// the path to the meminfo file which is produced by kernel
	meminfoPath = "/proc/meminfo"

	// hugepages stat field names on meminfo file
	hugePagesTotalFieldName = "HugePages_Total"
	hugePagesFreeFieldName  = "HugePages_Free"
)

var hugepagesSampleConfig = `
  ## Path to a NUMA nodes
  # numa_node_path = "/sys/devices/system/node"
  ## Path to a meminfo file
  # meminfo_path = "/proc/meminfo"
`

type Hugepages struct {
	NUMANodePath string `toml:"numa_node_path"`
	MeminfoPath  string `toml:"meminfo_path"`
}

func (mem *Hugepages) Description() string {
	return "Collects hugepages metrics from kernel and per NUMA node"
}

func (mem *Hugepages) SampleConfig() string {
	return hugepagesSampleConfig
}

func (mem *Hugepages) Gather(acc telegraf.Accumulator) error {
	err := mem.gatherStatsPerNode(acc)
	if err != nil {
		return err
	}

	return mem.gatherStatsFromMeminfo(acc)
}

// GatherHugepagesStatsPerNode collects hugepages stats per NUMA nodes
func (mem *Hugepages) gatherStatsPerNode(acc telegraf.Accumulator) error {
	numaNodeMetrics, err := statsPerNUMA(mem.NUMANodePath)
	if err != nil {
		return err
	}

	for k, v := range numaNodeMetrics {
		metrics := make(map[string]interface{})
		tags := map[string]string{
			"node": k,
		}
		metrics["free"] = v.Free
		metrics["nr"] = v.NR
		acc.AddFields("hugepages", metrics, tags)
	}
	return nil
}

// GatherHugepagesStatsFromMeminfo collects hugepages statistics from meminfo file
func (mem *Hugepages) gatherStatsFromMeminfo(acc telegraf.Accumulator) error {
	tags := map[string]string{
		"name": "meminfo",
	}
	metrics := make(map[string]interface{})
	meminfoMetrics, err := statsFromMeminfo(mem.MeminfoPath)
	if err != nil {
		return err
	}

	for k, v := range meminfoMetrics {
		metrics[k] = v
	}
	acc.AddFields("hugepages", metrics, tags)
	return nil
}

type hugepagesNUMAStats struct {
	Free int
	NR   int
}

// statsPerNUMA gathers hugepages statistics from each NUMA node
func statsPerNUMA(path string) (map[string]hugepagesNUMAStats, error) {
	var hugepagesStats = make(map[string]hugepagesNUMAStats)
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return hugepagesStats, err
	}

	for _, d := range dirs {
		if !d.IsDir() || !strings.HasPrefix(d.Name(), "node") {
			continue
		}

		hugepagesFree := filepath.Join(path, d.Name(), "hugepages", "hugepages-2048kB", "free_hugepages")
		hugepagesNR := filepath.Join(path, d.Name(), "hugepages", "hugepages-2048kB", "nr_hugepages")

		free, err := ioutil.ReadFile(hugepagesFree)
		if err != nil {
			return hugepagesStats, err
		}

		nr, err := ioutil.ReadFile(hugepagesNR)
		if err != nil {
			return hugepagesStats, err
		}

		f, err := strconv.Atoi(string(bytes.TrimSuffix(free, newlineByte)))
		if err != nil {
			return hugepagesStats, err
		}
		n, err := strconv.Atoi(string(bytes.TrimSuffix(nr, newlineByte)))
		if err != nil {
			return hugepagesStats, err
		}

		hugepagesStats[d.Name()] = hugepagesNUMAStats{Free: f, NR: n}
	}
	return hugepagesStats, nil
}

// statsFromMeminfo gathers hugepages statistics from kernel
func statsFromMeminfo(path string) (map[string]interface{}, error) {
	stats := map[string]interface{}{}
	meminfo, err := ioutil.ReadFile(path)
	if err != nil {
		return stats, err
	}
	lines := bytes.Split(meminfo, newlineByte)
	for _, l := range lines {
		if bytes.Contains(l, kbPrecisionByte) {
			continue
		}
		fields := bytes.Fields(l)
		if len(fields) < 2 {
			continue
		}
		fieldName := string(bytes.TrimSuffix(fields[0], colonByte))
		if fieldName == hugePagesTotalFieldName || fieldName == hugePagesFreeFieldName {
			val, _ := strconv.Atoi(string(fields[1]))
			stats[fieldName] = val
		}
	}
	return stats, nil
}

func (mem *Hugepages) Init() {
	if mem.NUMANodePath == "" {
		mem.NUMANodePath = numaNodePath
	}
	if mem.MeminfoPath == "" {
		mem.MeminfoPath = meminfoPath
	}
}

func init() {
	inputs.Add("hugepages", func() telegraf.Input {
		return &Hugepages{}
	})
}
