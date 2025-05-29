//go:build linux && amd64

package rapl

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const intelRAPLDir = "/sys/devices/virtual/powercap/intel-rapl"
const powerZonePrefix = "intel-rapl"

//go:embed sample.conf
var sampleConfig string

type RAPL struct {
	Log telegraf.Logger `toml:"-"`

	rapl           intelRAPL
	powerZones     []powerZone
	powerZoneNames []string
}

func (o *RAPL) Gather(acc telegraf.Accumulator) error {
	for i, zone := range o.powerZones {
		energy, err := o.rapl.readEnergy(&zone)
		if err != nil {
			return err
		}
		fields := map[string]interface{}{
			"energy_joules": float64(energy) / 1e6,
		}
		tags := map[string]string{
			"name":       o.powerZoneNames[i],
			"power_zone": zone.name(),
		}
		acc.AddFields("rapl", fields, tags)
	}
	return nil
}

func (*RAPL) SampleConfig() string {
	return sampleConfig
}

func (o *RAPL) Start(_ telegraf.Accumulator) error {
	o.rapl = intelRAPL{dir: intelRAPLDir}
	// Reads and caches the power zones.
	zones, err := o.rapl.powerZones()
	if err != nil {
		return err
	}
	o.powerZones = zones
	// Reads and caches the power zone names.
	o.powerZoneNames = make([]string, 0)
	for _, zone := range o.powerZones {
		name, err := o.rapl.readName(&zone)
		if err != nil {
			return err
		}
		o.powerZoneNames = append(o.powerZoneNames, name)
	}
	return nil
}

func (*RAPL) Stop() {
}

func init() {
	inputs.Add("rapl", func() telegraf.Input {
		return &RAPL{}
	})
}

// Represents a RAPL power zone.
type powerZone struct {
	ids []int
}

// Parses a directory name into a power zone.
func parsePowerZone(name string) (*powerZone, error) {
	if !strings.HasPrefix(name, powerZonePrefix) {
		return nil, fmt.Errorf("invalid power zone %s", name)
	}
	ids := make([]int, 0)
	for _, s := range strings.Split(name, ":")[1:] {
		id, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid power zone %s: %w", name, err)
		}
		ids = append(ids, id)
	}
	return &powerZone{ids: ids}, nil
}

// Returns the directory name of a power zone.
// Example: `intel-rapl:0:0`.
func (o *powerZone) name() string {
	parts := make([]string, len(o.ids)+1)
	parts[0] = powerZonePrefix
	for i, id := range o.ids {
		parts[i+1] = strconv.Itoa(id)
	}
	return strings.Join(parts, ":")
}

// Returns the complete relative path to a power zone.
// Example: `intel-rapl:0/intel-rapl:0:0`.
func (o *powerZone) path() string {
	parts := make([]string, len(o.ids))
	for i := range len(o.ids) {
		zone := &powerZone{o.ids[:i+1]}
		parts[i] = zone.name()
	}
	return filepath.Join(parts...)
}

// Represents an Intel RAPL interface.
type intelRAPL struct {
	dir string
}

// Returns the power zones in an Intel RAPL interface.
func (o *intelRAPL) powerZones() ([]powerZone, error) {
	zones := make([]powerZone, 0)
	todo := []string{o.dir}
	for len(todo) > 0 {
		path := todo[0]
		todo = todo[1:]
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			name := file.Name()
			zone, err := parsePowerZone(name)
			if err != nil {
				continue
			}
			zonePath := filepath.Join(path, name)
			if zonePath != filepath.Join(o.dir, zone.path()) {
				return nil, fmt.Errorf("incorrectly nested power zone %s", zonePath)
			}
			zones = append(zones, *zone)
			todo = append(todo, zonePath)
		}
	}
	return zones, nil
}

// Reads the human-readable name of a power zone.
func (o *intelRAPL) readName(zone *powerZone) (string, error) {
	b, err := os.ReadFile(filepath.Join(o.dir, zone.path(), "name"))
	if err != nil {
		return "", fmt.Errorf("could not read name: %w", err)
	}
	return strings.TrimSuffix(string(b), "\n"), nil
}

// Reads the energy of a power zone in micro joules (uJ).
func (o *intelRAPL) readEnergy(zone *powerZone) (uint64, error) {
	b, err := os.ReadFile(filepath.Join(o.dir, zone.path(), "energy_uj"))
	if err != nil {
		return 0, fmt.Errorf("could not read energy: %w", err)
	}
	energy, err := strconv.ParseUint(strings.TrimSuffix(string(b), "\n"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not read energy: %w", err)
	}
	return energy, nil
}
