//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package kernel_ksm

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type KernelKsm struct {
	sysPath string
}

func (*KernelKsm) SampleConfig() string {
	return sampleConfig
}

func (k *KernelKsm) Gather(acc telegraf.Accumulator) error {
	stats := map[string]interface{}{
		"full_scans":                         0,
		"max_page_sharing":                   0,
		"merge_across_nodes":                 0,
		"pages_shared":                       0,
		"pages_sharing":                      0,
		"pages_to_scan":                      0,
		"pages_unshared":                     0,
		"pages_volatile":                     0,
		"run":                                0,
		"sleep_millisecs":                    0,
		"stable_node_chains":                 0,
		"stable_node_chains_prune_millisecs": 0,
		"stable_node_dups":                   0,
		"use_zero_pages":                     0,
	}

	// these exist in very recent Linux versions only, but useful to include if there.
	extraStats := map[string]interface{}{
		"general_profit": 0,
	}

	if _, err := os.Stat(k.sysPath); os.IsNotExist(err) {
		// ksm probably not included in the kernel, bail out early
		return fmt.Errorf("kernel_ksm: %s does not exist. Is KSM included in the kernel?", k.sysPath)
	}

	for f := range stats {
		data, err := k.getProcValue(filepath.Join(k.sysPath, f))
		if err != nil {
			return err
		}

		m, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return err
		}

		stats[f] = m
	}

	for f := range extraStats {
		data, err := k.getProcValue(filepath.Join(k.sysPath, f))
		if err != nil {
			// if an extraStats metric doesn't exist in our kernel version, ignore it.
			continue
		}

		m, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return err
		}

		stats[f] = m
	}

	acc.AddCounter("kernel_ksm", stats, map[string]string{})

	return nil
}

func (k *KernelKsm) getProcValue(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("kernel_ksm: %s does not exist", path)
	} else if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(data), nil
}

func init() {
	inputs.Add("kernel_ksm", func() telegraf.Input {
		return &KernelKsm{
			sysPath: "/sys/kernel/mm/ksm",
		}
	})
}
