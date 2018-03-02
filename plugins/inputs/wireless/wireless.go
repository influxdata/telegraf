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

var (
	newLineByte     = []byte("\n")
	wirelessHeaders = map[int]string{
		0:  "interface",
		1:  "status",
		2:  "link",
		3:  "level",
		4:  "noise",
		5:  "nwid",
		6:  "crypt",
		7:  "frag",
		8:  "retry",
		9:  "misc",
		10: "missed_beacon",
	}
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

// Wireless is used to store configuration values.
type Wireless struct {
	ProcNetWireless string `toml:"proc_net_wireless"`
	DumpZeros       bool   `toml:"dump_zeros"`
}

var sampleConfig = `
  ## file paths for proc files. If empty default paths will be used:
  ##    /proc/net/wireless
  proc_net_wireless = "/proc/net/wireless"

  ## dump metrics with 0 values too
  dump_zeros       = false
`

// Desciption returns information about the plugin.
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

	wireless, err := ioutil.ReadFile(w.ProcNetWireless)
	if err != nil {
		return err
	}

	// collect wireless data
	err = w.gatherWireless(wireless, acc)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wireless) gatherWireless(data []byte, acc telegraf.Accumulator) error {
	metrics, tags, err := loadWirelessTable(data, w.DumpZeros)
	if err != nil {
		return err
	}
	acc.AddFields("wireless", metrics, tags)
	return nil
}

// loadPaths can be used to read paths firstly from config
// if it is empty then try read from env variables
func (w *Wireless) loadPaths() {
	if w.ProcNetWireless == "" {
		w.ProcNetWireless = proc(ENV_WIRELESS, NET_WIRELESS)
	}
}

func loadWirelessTable(table []byte, dumpZeros bool) (map[string]interface{}, map[string]string, error) {
	entries := map[string]interface{}{}
	tags := map[string]string{}
	// split the lines by newline
	lines := bytes.Split(table, newLineByte)
	var value int64
	var err error
	// iterate over intefaces
	for i := 2; i < len(lines); i = i + 1 {
		if len(lines[i]) == 0 {
			continue
		}
		fields := strings.Fields(string(lines[i]))
		for j := 0; j < len(fields); j = j + 1 {
			// parse interface
			if j == 0 {
				tags[wirelessHeaders[j]] = strings.Trim(fields[j], ":")
				continue
			}
			// parse value
			value, err = strconv.ParseInt(strings.Trim(fields[j], "."), 10, 64)
			if err != nil {
				continue
			}
			// value is zero
			if value == 0 && dumpZeros {
				continue
			}
			// the value is not zero, so parse it
			entries[wirelessHeaders[j]] = value
		}
	}
	return entries, tags, nil
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
