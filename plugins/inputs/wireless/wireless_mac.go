//go:build darwin
// +build darwin

package wireless

import (
	"errors"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// default executable path
const (
	OSXCMD = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I"
)

func (w *Wireless) exe_cmd(cmd string, wg *sync.WaitGroup) ([]byte, error) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]
	out, err := exec.Command(head, parts...).Output()
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return out, err
}

func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	if runtime.GOOS == "darwin" {
		// collect MAC OS wireless data
		wg := new(sync.WaitGroup)
		wg.Add(3)
		wireless, err := w.exe_cmd(OSXCMD, wg)
		if err != nil {
			return err
		}
		metrics, tags, err := w.loadMacWirelessTable(wireless, w.DumpZeros)
		if err != nil {
			return err
		}
		acc.AddGauge("wireless", metrics, tags)
		return nil
	}
	return errors.New("OS Not Supported")
}

func (w *Wireless) loadMacWirelessTable(table []byte, dumpZeros bool) (map[string]interface{}, map[string]string, error) {
	metrics := strings.Split(string(string(table)), "\n")
	tags := make(map[string]string)
	points := make(map[string]interface{})
	for x := 0; x < len(metrics); x++ {
		fm := strings.Split(strings.TrimSpace(metrics[x]), ":")
		if len(fm) > 1 {
			name := strings.Replace(strings.Trim(strings.TrimSpace(fm[0]), ":"), " ", "_", -1)
			v := strings.TrimSpace(fm[1])
			val, err := strconv.Atoi(v)
			if err == nil { // it's a number
				if !dumpZeros && val == 0 {
					continue
				}
				points[name] = int64(val)
			} else { // it's a string
				tags[name] = strings.Replace(v, " ", "_", -1)
			}
		}
	}
	tags["interface"] = "airport"
	return points, tags, nil
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
