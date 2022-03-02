//go:build linux
// +build linux

package hugepages

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
	// path to root huge page control directory
	rootHugepagePath = "/sys/kernel/mm/hugepages"
	// path where per NUMA node statistics are kept
	numaNodePath = "/sys/devices/system/node"
	// path to the meminfo file
	meminfoPath = "/proc/meminfo"

	rootHugepages    = "root"
	perNodeHugepages = "per_node"
	meminfoHugepages = "meminfo"
)

var (
	newlineByte     = []byte("\n")
	colonByte       = []byte(":")
	kbPrecisionByte = []byte("kB")

	hugepagesMetricsRoot = []string{
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
  ## Supported huge page types:
  ##   - "root" - based on root huge page control directory: /sys/kernel/mm/hugepages
  ##   - "per_node" - based on per NUMA node directories: /sys/devices/system/node/node[0-9]*/hugepages
  ##   - "meminfo" - based on /proc/meminfo file
  # hugepages_types = ["root", "per_node"]
`

type Hugepages struct {
	HugepagesTypes []string `toml:"hugepages_types"`

	gatherRoot    bool
	gatherPerNode bool
	gatherMeminfo bool

	rootHugepagePath string
	numaNodePath     string
	meminfoPath      string
}

func (h *Hugepages) Description() string {
	return "Gathers huge pages measurements."
}

func (h *Hugepages) SampleConfig() string {
	return hugepagesSampleConfig
}

func (h *Hugepages) Init() error {
	err := h.parseHugepagesConfig()
	if err != nil {
		return err
	}

	h.rootHugepagePath = rootHugepagePath
	h.numaNodePath = numaNodePath
	h.meminfoPath = meminfoPath

	return nil
}

func (h *Hugepages) Gather(acc telegraf.Accumulator) error {
	var err error

	if h.gatherRoot {
		err = h.gatherRootStats(acc)
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

// gatherStatsPerNode collects root hugepages statistics
func (h *Hugepages) gatherRootStats(acc telegraf.Accumulator) error {
	return h.gatherFromHugepagePath(h.rootHugepagePath, hugepagesMetricsRoot, "hugepages_"+rootHugepages, nil, acc)
}

// gatherStatsPerNode collects hugepages statistics per NUMA node
func (h *Hugepages) gatherStatsPerNode(acc telegraf.Accumulator) error {
	nodeDirs, err := ioutil.ReadDir(h.numaNodePath)
	if err != nil {
		return err
	}

	// read metrics from: node*/hugepages/hugepages-*/*
	for _, nodeDir := range nodeDirs {
		if !nodeDir.IsDir() || !strings.HasPrefix(nodeDir.Name(), "node") {
			continue
		}

		nodeNumber := strings.TrimPrefix(nodeDir.Name(), "node")
		_, err := strconv.Atoi(nodeNumber)
		if err != nil {
			continue
		}

		perNodeTags := map[string]string{
			"node": nodeNumber,
		}
		hugepagesPath := filepath.Join(h.numaNodePath, nodeDir.Name(), "hugepages")
		err = h.gatherFromHugepagePath(hugepagesPath, hugepagesMetricsPerNUMANode, "hugepages_"+perNodeHugepages, perNodeTags, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Hugepages) gatherFromHugepagePath(hugepagesPath string, possibleMetrics []string, measurementName string,
	tagsToUse map[string]string, acc telegraf.Accumulator) error {
	// read metrics from: hugepages/hugepages-*/*
	hugepagesDirs, err := ioutil.ReadDir(hugepagesPath)
	if err != nil {
		return err
	}

	for _, hugepagesDir := range hugepagesDirs {
		if !hugepagesDir.IsDir() || !strings.HasPrefix(hugepagesDir.Name(), "hugepages-") {
			continue
		}

		hugepagesSize := strings.TrimPrefix(strings.TrimSuffix(hugepagesDir.Name(), "kB"), "hugepages-")
		_, err := strconv.Atoi(hugepagesSize)
		if err != nil {
			continue
		}

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
			tags["hugepages_size_kb"] = hugepagesSize

			acc.AddFields(measurementName, metrics, tags)
		}
	}
	return nil
}

// gatherStatsFromMeminfo collects hugepages statistics from meminfo file
func (h *Hugepages) gatherStatsFromMeminfo(acc telegraf.Accumulator) error {
	meminfo, err := ioutil.ReadFile(h.meminfoPath)
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
				fieldName = fieldName + "_kb"
			}
			metrics[fieldName] = fieldValue
		}
	}

	acc.AddFields("hugepages_"+meminfoHugepages, metrics, map[string]string{})
	return nil
}

func (h *Hugepages) parseHugepagesConfig() error {
	// default
	if h.HugepagesTypes == nil {
		h.gatherRoot = true
		h.gatherMeminfo = true
		return nil
	}

	// empty array
	if len(h.HugepagesTypes) == 0 {
		return fmt.Errorf("plugin was configured with nothing to read")
	}

	for _, hugepagesType := range h.HugepagesTypes {
		switch hugepagesType {
		case rootHugepages:
			h.gatherRoot = true
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
