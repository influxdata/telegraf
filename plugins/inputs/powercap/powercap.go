//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package powercap

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Powercap struct {
	Log telegraf.Logger `toml:"-"`

	previousReadings map[string]energyReading
}

type energyReading struct {
	microjoules uint64
	time        time.Time
}

func (*Powercap) SampleConfig() string {
	return sampleConfig
}

func (p *Powercap) Init() error {
	p.previousReadings = make(map[string]energyReading)
	return nil
}

func (p *Powercap) Gather(acc telegraf.Accumulator) error {
	energyFiles, err := filepath.Glob(filepath.Join(internal.GetSysPath(), "class", "powercap", "*", "energy_uj"))
	if err != nil {
		return fmt.Errorf("finding power zones failed: %w", err)
	}

	for _, energyFile := range energyFiles {
		if err := p.gatherZone(acc, filepath.Dir(energyFile)); err != nil {
			acc.AddError(err)
		}
	}
	return nil
}

func (p *Powercap) gatherZone(acc telegraf.Accumulator, zonePath string) error {
	zone := filepath.Base(zonePath)

	name, err := readString(filepath.Join(zonePath, "name"))
	if err != nil {
		return fmt.Errorf("reading name of zone %q failed: %w", zone, err)
	}

	microjoules, err := readUint(filepath.Join(zonePath, "energy_uj"))
	if err != nil {
		return fmt.Errorf("reading energy of zone %q failed: %w", zone, err)
	}
	current := energyReading{microjoules: microjoules, time: time.Now()}

	fields := map[string]any{"energy_uj": current.microjoules}
	if previous, found := p.previousReadings[zone]; found {
		wrapsAt, err := readUint(filepath.Join(zonePath, "max_energy_range_uj"))
		if err != nil {
			return fmt.Errorf("reading energy range of zone %q failed: %w", zone, err)
		}
		fields["power_watts"] = averageWatts(previous, current, wrapsAt)
	}
	p.previousReadings[zone] = current

	acc.AddFields("powercap", fields, map[string]string{"zone": zone, "name": name}, current.time)
	return nil
}

func averageWatts(earlier, later energyReading, wrapsAt uint64) float64 {
	consumed := later.microjoules - earlier.microjoules
	if later.microjoules < earlier.microjoules {
		consumed = wrapsAt - earlier.microjoules + later.microjoules
	}
	return float64(consumed) / 1e6 / later.time.Sub(earlier.time).Seconds()
}

func readString(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func readUint(path string) (uint64, error) {
	content, err := readString(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(content, 10, 64)
}

func init() {
	inputs.Add("powercap", func() telegraf.Input {
		return &Powercap{}
	})
}
