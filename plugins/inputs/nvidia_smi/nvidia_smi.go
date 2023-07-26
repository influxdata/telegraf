//go:generate ../../../tools/readme_config_includer/generator
package nvidia_smi

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/schema_v11"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/schema_v12"
)

//go:embed sample.conf
var sampleConfig string

// NvidiaSMI holds the methods for this plugin
type NvidiaSMI struct {
	BinPath string          `toml:"bin_path"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`
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
	schema := "v11"

	buf := bytes.NewBuffer(data)
	decoder := xml.NewDecoder(buf)
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("reading token failed: %w", err)
		}
		d, ok := token.(xml.Directive)
		if !ok {
			continue
		}
		directive := string(d)
		if !strings.HasPrefix(directive, "DOCTYPE") {
			continue
		}
		parts := strings.Split(directive, " ")
		s := strings.Trim(parts[len(parts)-1], "\" ")
		if strings.HasPrefix(s, "nvsmi_device_") && strings.HasSuffix(s, ".dtd") {
			schema = strings.TrimSuffix(strings.TrimPrefix(s, "nvsmi_device_"), ".dtd")
		} else {
			smi.Log.Debugf("Cannot find schema version in %q", directive)
		}
		break
	}
	smi.Log.Debugf("Using schema version in %s", schema)

	switch schema {
	case "v10", "v11":
		return schema_v11.Parse(acc, data)
	case "v12":
		return schema_v12.Parse(acc, data)
	}
	return fmt.Errorf("unknown schema %q", schema)
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
