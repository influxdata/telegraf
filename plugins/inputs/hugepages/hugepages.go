package system

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// default path where statistics are kept per NUMA nodes
	defaultNumaNodePath = "/sys/devices/system/node"
	// default path to the meminfo file which is produced by kernel
	defaultMeminfoPath = "/proc/meminfo"
)

var (
	newlineByte     = []byte("\n")
	colonByte       = []byte(":")
	kbPrecisionByte = []byte("kB")

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
  ## Path to a NUMA nodes
  # numa_node_path = "/sys/devices/system/node"
  ## Path to a meminfo file
  # meminfo_path = "/proc/meminfo"
`

type Hugepages struct {
	NUMANodePath string `toml:"numa_node_path"`
	MeminfoPath  string `toml:"meminfo_path"`
}

func (h *Hugepages) Description() string {
	return "Collects hugepages metrics from kernel and per NUMA node"
}

func (h *Hugepages) SampleConfig() string {
	return hugepagesSampleConfig
}

func (h *Hugepages) Init() error {
	if h.NUMANodePath == "" {
		h.NUMANodePath = defaultNumaNodePath
	}
	if h.MeminfoPath == "" {
		h.MeminfoPath = defaultMeminfoPath
	}

	return nil
}

func (h *Hugepages) Gather(acc telegraf.Accumulator) error {
	err := h.gatherStatsPerNode(acc)
	if err != nil {
		return err
	}

	return h.gatherStatsFromMeminfo(acc)
}

// gatherStatsPerNode collects hugepages stats per NUMA nodes
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

		hugepagesPath := filepath.Join(h.NUMANodePath, nodeDir.Name(), "hugepages")
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
				if mode := metricFile.Mode(); !mode.IsRegular() || !choice.Contains(metricFile.Name(), hugepagesMetricsPerNUMANode) {
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
				tags := map[string]string{
					"node":           nodeDir.Name(),
					"hugepages_size": hugepagesSize,
				}
				acc.AddFields("hugepages", metrics, tags)
			}
		}
	}
	return nil
}

// gatherStatsFromMeminfo collects hugepages statistics from meminfo file
func (h *Hugepages) gatherStatsFromMeminfo(acc telegraf.Accumulator) error {
	metrics := make(map[string]interface{})
	tags := map[string]string{
		"name": "meminfo",
	}

	meminfo, err := ioutil.ReadFile(h.MeminfoPath)
	if err != nil {
		return err
	}

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

	acc.AddFields("hugepages", metrics, tags)
	return nil
}

func init() {
	inputs.Add("hugepages", func() telegraf.Input {
		return &Hugepages{}
	})
}
