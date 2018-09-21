// +build linux

package wireless

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default host proc path
const defaultHostProc = "/proc"

// env host proc variable name
const envProc = "HOST_PROC"

var newLineByte = []byte("\n")

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
	HostProc string `toml:"host_proc"`
}

var sampleConfig = `
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"
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
	// load proc path, get default value if config value and env variable are empty
	w.loadPath()

	wirelessPath := path.Join(w.HostProc, "net", "wireless")
	table, err := ioutil.ReadFile(wirelessPath)
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
	lines := bytes.Split(table, newLineByte)

	// iterate over interfaces
	for i := 2; i < len(lines); i = i + 1 {
		if len(lines[i]) == 0 {
			continue
		}
		values := []int64{}
		fields := strings.Fields(string(lines[i]))
		for k := 1; k < len(fields); k = k + 1 {
			v, err := strconv.ParseInt(strings.Trim(fields[k], "."), 10, 64)
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

// loadPath can be used to read path firstly from config
// if it is empty then try read from env variable
func (w *Wireless) loadPath() {
	if w.HostProc == "" {
		w.HostProc = proc(envProc, defaultHostProc)
	}
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// return default path
	return path
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
