//go:generate ../../../tools/readme_config_includer/generator
package smartctl

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// execCommand is used to mock commands in tests.
var execCommand = exec.Command

type Smartctl struct {
	Path           string          `toml:"path"`
	NoCheck        string          `toml:"no_check"`
	UseSudo        bool            `toml:"use_sudo"`
	Timeout        config.Duration `toml:"timeout"`
	DevicesInclude []string        `toml:"devices_include"`
	DevicesExclude []string        `toml:"devices_exclude"`
	Log            telegraf.Logger `toml:"-"`

	deviceFilter filter.Filter
}

func (*Smartctl) SampleConfig() string {
	return sampleConfig
}

func (s *Smartctl) Init() error {
	if s.Path == "" {
		s.Path = "/usr/sbin/smartctl"
	}

	switch s.NoCheck {
	case "never", "sleep", "standby", "idle":
	case "":
		s.NoCheck = "standby"
	default:
		return fmt.Errorf("invalid no_check value: %s", s.NoCheck)
	}

	if s.Timeout == 0 {
		s.Timeout = config.Duration(time.Second * 30)
	}

	if len(s.DevicesInclude) != 0 && len(s.DevicesExclude) != 0 {
		return errors.New("cannot specify both devices_include and devices_exclude")
	}

	var err error
	s.deviceFilter, err = filter.NewIncludeExcludeFilter(s.DevicesInclude, s.DevicesExclude)
	if err != nil {
		return err
	}

	return nil
}

func (s *Smartctl) Gather(acc telegraf.Accumulator) error {
	devices, err := s.scan()
	if err != nil {
		return fmt.Errorf("Error scanning system: %w", err)
	}

	for _, device := range devices {
		if err := s.scanDevice(acc, device.Name, device.Type); err != nil {
			return fmt.Errorf("Error getting device %s: %w", device, err)
		}
	}

	return nil
}

func init() {
	// Set LC_NUMERIC to uniform numeric output from cli tools
	_ = os.Setenv("LC_NUMERIC", "en_US.UTF-8")
	inputs.Add("smartctl", func() telegraf.Input {
		return &Smartctl{
			Timeout: config.Duration(time.Second * 30),
		}
	})
}
