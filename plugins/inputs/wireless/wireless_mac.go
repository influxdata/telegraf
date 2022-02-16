//go:build darwin
// +build darwin

package wireless

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default executable path & flags
const (
	OSXCMD = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	FLAG = "-I"
)


func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	cmd := exec.Command(OSXCMD, FLAG)
	wireless, err := internal.StdOutputTimeout(cmd, 2*time.Second)// cmd.Output()
	if err != nil {
		return err
	}
	metrics, tags, err := w.loadMacWirelessTable(wireless)
	if err != nil {
		return err
	}
	acc.AddFields("wireless", metrics, tags)
	return nil
}

func (w *Wireless) loadMacWirelessTable(table []byte) (map[string]interface{}, map[string]string, error) {
	lines := strings.Split(string(table), "\n")
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for _, line := range lines {
		fm := strings.Split(strings.TrimSpace(line), ":")
		if len(fm) < 2 {
			continue
		}
		name := strings.Replace(strings.Trim(strings.TrimSpace(fm[0]), ":"), " ", "_", -1)
		v := strings.TrimSpace(fm[1])
		val, err := strconv.Atoi(v)
		if err == nil { // it's a number
				fields[name] = int64(val)
		} else { // it's a string
		if name == "channel" || name == "BSSID" || name == "SSID" {
				fields[name] = strings.Replace(v, " ", "_", -1)
			} else {
			tags[name] = strings.Replace(v, " ", "_", -1)
			}
		}
	}
	tags["interface"] = "airport"
	return fields, tags, nil
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
