// +build linux

package wireless

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default file paths
const (
	NET_WIRELESS = "/net/wireless"
	NET_PROC     = "/proc"
)

// env variable names
const (
	ENV_WIRELESS = "PROC_NET_WIRELESS"
	ENV_ROOT     = "PROC_ROOT"
)

var (
	newLineByte = []byte("\n")
)

type wirelessInterface struct {
	Interface string
	Status    int64
	Link      int64
	Level     int64
	Noise     int64
	Nwid      int64
	Crypt     int64
	Frag      int64
	Retry     int64
	Misc      int64
	Beacon    int64
}

// Wireless is used to store configuration values.
type Wireless struct {
	ProcNetWireless string `toml:"proc_net_wireless"`
}

var sampleConfig = `
  ## file paths for proc files. If empty default paths will be used:
  ##    /proc/net/wireless
  proc_net_wireless = "/proc/net/wireless"
`

// Description returns information about the plugin.
func (w *Wireless) Description() string {
	return "Monitor wifi signal strength and quality"
}

// SampleConfig displays configuration instructions.
func (w *Wireless) SampleConfig() string {
	return sampleConfig
}

// Gather collects the wireless information.
func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	// load paths, get from env if config values are empty
	w.loadPaths()

	table, err := ioutil.ReadFile(w.ProcNetWireless)
	if err != nil {
		return err
	}

	interfaces, err := loadWirelessTable(table)
	if err != nil {
		return err
	}
	for _, w := range interfaces {
		tags := map[string]string{
			"interface": w.Interface,
		}
		fieldsG := map[string]interface{}{
			"status": w.Status,
			"link":   w.Link,
			"level":  w.Level,
			"noise":  w.Noise,
		}
		fieldsC := map[string]interface{}{
			"nwid":   w.Nwid,
			"crypt":  w.Crypt,
			"frag":   w.Frag,
			"retry":  w.Retry,
			"misc":   w.Misc,
			"beacon": w.Beacon,
		}
		acc.AddGauge("wireless", fieldsG, tags)
		acc.AddCounter("wireless", fieldsC, tags)
	}

	return nil
}

func loadWirelessTable(table []byte) ([]*wirelessInterface, error) {
	var w []*wirelessInterface

	// split the lines by newline
	lines := bytes.Split(table, newLineByte)

	// iterate over intefaces
	for i := 2; i < len(lines); i = i + 1 {
		if len(lines[i]) == 0 {
			continue
		}
		fields := strings.Fields(string(lines[i]))
		var values []int64
		for k := 1; k < len(fields); k = k + 1 {
			v, err := strconv.ParseInt(fields[i], ".")
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		w = append(w, &wirelessInterface{
			Interface: strings.Trim(fields[0], ":"),
			Status:    values[0],
			Link:      values[1],
			Level:     values[2],
			Noise:     values[3],
			Nwid:      values[4],
			Crypt:     values[5],
			Frag:      values[6],
			Retry:     values[7],
			Misc:      values[8],
			Beacon:    values[9],
		})
	}
	return w, nil
}

// loadPaths can be used to read paths firstly from config
// if it is empty then try read from env variables
func (w *Wireless) loadPaths() {
	if w.ProcNetWireless == "" {
		w.ProcNetWireless = proc(ENV_WIRELESS, NET_WIRELESS)
	}
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// try to read root path, or use default root path
	root := os.Getenv(ENV_ROOT)
	if root == "" {
		root = NET_PROC
	}
	return root + path
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
