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
)

// default executable path & flags
const (
	OSXCMD = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	FLAG   = "-I"
)

func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	cmd := exec.Command(OSXCMD, FLAG)
	wireless, err := internal.StdOutputTimeout(cmd, 2*time.Second)
	if err != nil {
		return err
	}
	fields, tags := w.loadMacWirelessTable(wireless)
	acc.AddFields("wireless", fields, tags)
	return nil
}

func (w *Wireless) loadMacWirelessTable(table []byte) (map[string]interface{}, map[string]string) {
	lines := strings.Split(string(table), "\n")
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for _, line := range lines {
		fm := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(fm) < 2 || fm[1] == "" {
			continue
		}
		name := strings.Replace(strings.Trim(strings.TrimSpace(fm[0]), ":"), " ", "_", -1)
		v := strings.TrimSpace(fm[1])
		switch name {
		case "channel", "BSSID", "SSID":
			tags[name] = "\"" + strings.Replace(v, " ", "_", -1) + "\""
		default:
			if val, err := strconv.Atoi(v); err == nil {
				// it's a number
				fields[name] = int64(val)
			} else {
				// it's a string
				tags[name] = "\"" + strings.Replace(v, " ", "_", -1) + "\""
			}
		}
	}
	return fields, tags
}
