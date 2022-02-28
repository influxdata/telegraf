package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// default path where global hugepages statistics are kept
	defaultGlobalHugepagePath = "/sys/kernel/mm/hugepages"
	// default path where per NUMA node statistics are kept
	defaultNumaNodePath = "/sys/devices/system/node"
	// default path to the meminfo file which is produced by kernel
	defaultMeminfoPath = "/proc/meminfo"

	globalHugepages  = "global"
	perNodeHugepages = "per_node"
	meminfoHugepages = "meminfo"
)

var (
	newlineByte     = []byte("\n")
	colonByte       = []byte(":")
	kbPrecisionByte = []byte("kB")

	hugepagesMetricsGlobal = []string{
		"free_hugepages",
		"nr_hugepages",
		"nr_hugepages_mempolicy",
		"nr_overcommit_hugepages",
		"resv_hugepages",
		"surplus_hugepages",
	}

	hugepagesMetricsPerNUMANode = []string{
		"free_hugepages",
		"nr_hugepages",
		"surplus_hugepages",
	}

	hugepagesMetricsFromMeminfo = []string{
		"HugePages_Total",
		"HugePages_Free",
		"HugePages_Rsvd",
		"HugePages_Surp",
		"Hugepagesize",
		"Hugetlb",
		"AnonHugePages",
		"ShmemHugePages",
		"FileHugePages",
	}
)

var hugepagesSampleConfig = `
  ## Path to global hugepages
  # global_hugepage_path = "/sys/kernel/mm/hugepages"
  ## Path to NUMA nodes
  # numa_node_path = "/sys/devices/system/node"
  ## Path to meminfo file
  # meminfo_path = "/proc/meminfo"
  ## Hugepages types to gather
  ## Supported options: "global", "per_node", "meminfo"
  # hugepages_types = ["global", "meminfo"]
`

type Hugepages struct {
	GlobalHugepagePath string   `toml:"global_hugepage_path"`
	NUMANodePath       string   `toml:"numa_node_path"`
	MeminfoPath        string   `toml:"meminfo_path"`
	HugepagesTypes     []string `toml:"hugepages_types"`

	gatherGlobal  bool
	gatherPerNode bool
	gatherMeminfo bool
}

func (h *Hugepages) Description() string {
	return "Collects hugepages metrics."
}

func (h *Hugepages) SampleConfig() string {
	return hugepagesSampleConfig
}

func (h *Hugepages) Init() error {
	err := h.parseHugepagesConfig()
	if err != nil {
		return err
	}

	if h.GlobalHugepagePath == "" {
		h.GlobalHugepagePath = defaultGlobalHugepagePath
	}
	if h.NUMANodePath == "" {
		h.NUMANodePath = defaultNumaNodePath
	}
	if h.MeminfoPath == "" {
		h.MeminfoPath = defaultMeminfoPath
	}

	return nil
}

func (h *Hugepages) Gather(acc telegraf.Accumulator) error {
	var err error

	if h.gatherGlobal {
		err = h.gatherGlobalStats(acc)
		if err != nil {
			return err
		}
	}

	if h.gatherPerNode {
		err = h.gatherStatsPerNode(acc)
		if err != nil {
			return err
		}
	}

	if h.gatherMeminfo {
		err = h.gatherStatsFromMeminfo(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

// gatherStatsPerNode collects global hugepages statistics
func (h *Hugepages) gatherGlobalStats(acc telegraf.Accumulator) error {
	globalTags := map[string]string{
		"name": globalHugepages,
	}
	return h.gatherFromHugepagePath(h.GlobalHugepagePath, hugepagesMetricsGlobal, globalTags, acc)
}

// gatherStatsPerNode collects hugepages statistics per NUMA node
func (h *Hugepages) gatherStatsPerNode(acc telegraf.Accumulator) error {
	nodeDirs, err := ioutil.ReadDir(h.NUMANodePath)
	if err != nil {
		return err
	}

	// read metrics from: node*/hugepages/hugepages-*/*
	for _, nodeDir := range nodeDirs {
		if !nodeDir.IsDir() || !strings.HasPrefix(nodeDir.Name(), "node") {
			continue
		}

		perNodeTags := map[string]string{
			"name": perNodeHugepages,
			"node": nodeDir.Name(),
		}
		hugepagesPath := filepath.Join(h.NUMANodePath, nodeDir.Name(), "hugepages")
		err = h.gatherFromHugepagePath(hugepagesPath, hugepagesMetricsPerNUMANode, perNodeTags, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Hugepages) gatherFromHugepagePath(hugepagesPath string, possibleMetrics []string, tagsToUse map[string]string, acc telegraf.Accumulator) error {
	// read metrics from: hugepages/hugepages-*/*
	hugepagesDirs, err := ioutil.ReadDir(hugepagesPath)
	if err != nil {
		return err
	}

	for _, hugepagesDir := range hugepagesDirs {
		if !hugepagesDir.IsDir() || !strings.HasPrefix(hugepagesDir.Name(), "hugepages-") {
			continue
		}
		hugepagesSize := strings.TrimPrefix(hugepagesDir.Name(), "hugepages-")

		metricsPath := filepath.Join(hugepagesPath, hugepagesDir.Name())
		metricFiles, err := ioutil.ReadDir(metricsPath)
		if err != nil {
			return err
		}

		metrics := make(map[string]interface{})
		for _, metricFile := range metricFiles {
			if mode := metricFile.Mode(); !mode.IsRegular() || !choice.Contains(metricFile.Name(), possibleMetrics) {
				continue
			}

			metricBytes, err := ioutil.ReadFile(filepath.Join(metricsPath, metricFile.Name()))
			if err != nil {
				return err
			}

			metricValue, err := strconv.Atoi(string(bytes.TrimSuffix(metricBytes, newlineByte)))
			if err != nil {
				return err
			}

			metrics[metricFile.Name()] = metricValue
		}

		if len(metrics) > 0 {
			tags := make(map[string]string)
			for key, value := range tagsToUse {
				tags[key] = value
			}
			tags["hugepages_size"] = hugepagesSize

			acc.AddFields("hugepages", metrics, tags)
		}
	}
	return nil
}

// gatherStatsFromMeminfo collects hugepages statistics from meminfo file
func (h *Hugepages) gatherStatsFromMeminfo(acc telegraf.Accumulator) error {
	meminfo, err := ioutil.ReadFile(h.MeminfoPath)
	if err != nil {
		return err
	}

	metrics := make(map[string]interface{})
	lines := bytes.Split(meminfo, newlineByte)
	for _, line := range lines {
		fields := bytes.Fields(line)
		if len(fields) < 2 {
			continue
		}
		fieldName := string(bytes.TrimSuffix(fields[0], colonByte))
		if choice.Contains(fieldName, hugepagesMetricsFromMeminfo) {
			fieldValue, err := strconv.Atoi(string(fields[1]))
			if err != nil {
				return err
			}

			if bytes.Contains(line, kbPrecisionByte) {
				fieldName = fieldName + "_kB"
			}
			metrics[fieldName] = fieldValue
		}
	}

	tags := map[string]string{
		"name": meminfoHugepages,
	}
	acc.AddFields("hugepages", metrics, tags)
	return nil
}

func (h *Hugepages) parseHugepagesConfig() error {
	// default
	if h.HugepagesTypes == nil {
		h.gatherGlobal = true
		h.gatherMeminfo = true
		return nil
	}

	// empty array
	if len(h.HugepagesTypes) == 0 {
		return fmt.Errorf("plugin was configured with nothing to read")
	}

	for _, hugepagesType := range h.HugepagesTypes {
		switch hugepagesType {
		case globalHugepages:
			h.gatherGlobal = true
		case perNodeHugepages:
			h.gatherPerNode = true
		case meminfoHugepages:
			h.gatherMeminfo = true
		default:
			return fmt.Errorf("provided hugepages type `%s` is not valid", hugepagesType)
		}
	}

	return nil
}

func init() {
	inputs.Add("hugepages", func() telegraf.Input {
		return &Hugepages{}
	})
}
