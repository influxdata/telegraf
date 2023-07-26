//go:generate ../../../tools/readme_config_includer/generator
package nvidia_smi

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/schema_v11"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "nvidia_smi"

// NvidiaSMI holds the methods for this plugin
type NvidiaSMI struct {
	BinPath string
	Timeout config.Duration
}

func (*NvidiaSMI) SampleConfig() string {
	return sampleConfig
}

func (smi *NvidiaSMI) Init() error {
	if _, err := os.Stat(smi.BinPath); os.IsNotExist(err) {
		binPath, err := exec.LookPath("nvidia-smi")
		// fail-fast
		if err != nil {
			return fmt.Errorf("nvidia-smi not found in %q and not in PATH; please make sure nvidia-smi is installed and/or is in PATH", smi.BinPath)
		}
		smi.BinPath = binPath
	}

	return nil
}

// Gather implements the telegraf interface
func (smi *NvidiaSMI) Gather(acc telegraf.Accumulator) error {
	data, err := smi.pollSMI()
	if err != nil {
		return err
	}

	return smi.parse(acc, data)
}

func (smi *NvidiaSMI) parse(acc telegraf.Accumulator, data []byte) error {
	return schema_v11.Parse(acc, data)
}

func (smi *NvidiaSMI) pollSMI() ([]byte, error) {
	// Construct and execute metrics query
	ret, err := internal.CombinedOutputTimeout(exec.Command(smi.BinPath, "-q", "-x"), time.Duration(smi.Timeout))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func init() {
	inputs.Add("nvidia_smi", func() telegraf.Input {
		return &NvidiaSMI{
			BinPath: "/usr/bin/nvidia-smi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
