package mac_wireless

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// default file paths
const (
	CMD = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I"
)

type Wireless struct {
	CMD       string `toml:"cmd"`
	DumpZeros bool   `toml:"dump_zeros"`
}

var sampleConfig = `
  ## command to get wireless info. If empty default will be used:
  ##  
  ## This can also be overridden with env variable, see README.
  cmd = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I"
  ## dump metrics with 0 values too
  dump_zeros       = true
`

func (ns *Wireless) Description() string {
	return "Collect wireless interface link quality metrics"
}

func (ns *Wireless) SampleConfig() string {
	return sampleConfig
}
func exe_cmd(cmd string, wg *sync.WaitGroup) ([]byte, error) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	out, err := exec.Command(head, parts...).Output()
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return out, err
}

func (ns *Wireless) Gather(acc telegraf.Accumulator) error {
	// load paths, get from env if config values are empty
	ns.loadPath()

	// collect wireless data
	wg := new(sync.WaitGroup)
	wg.Add(3)
	wireless, err := exe_cmd(CMD, wg)
	if err != nil {
		return err
	}
	err = ns.gatherWireless(wireless, acc)
	if err != nil {
		return err
	}
	return nil
}

func (ns *Wireless) gatherWireless(data []byte, acc telegraf.Accumulator) error {
	metrics, tags, err := loadWirelessTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	acc.AddFields("wireless", metrics, tags)
	return nil
}

func loadWirelessTable(table []byte, dumpZeros bool) (map[string]interface{}, map[string]string, error) {
	metrics := map[string]interface{}{}
	tags := map[string]string{}
	myLines := strings.Split(string(table), "\n")
	for x := 0; x < len(myLines)-1; x++ {
		f := strings.SplitN(myLines[x], ":", 2)
		f[0] = strings.Trim(f[0], " ")
		f[1] = strings.Trim(f[1], " ")
		if f[0] == "BSSID" {
			tags[strings.Replace(strings.Trim(f[0], " "), " ", "_", -1)] = strings.Replace(strings.Trim(string(f[1]), " "), " ", "_", -1)
			continue
		}
		n, err := strconv.ParseInt(strings.Trim(f[1], " "), 10, 64)
		if err != nil {
			tags[strings.Replace(strings.Trim(f[0], " "), " ", "_", -1)] = strings.Replace(strings.Trim(f[1], " "), " ", "_", -1)
			continue
		}
		if n == 0 {
			if dumpZeros {
				continue
			}
		}
		metrics[strings.Trim(f[0], " ")] = n

	}
	tags["interface"] = "airport"
	return metrics, tags, nil

}

// loadPath can be used to read path firstly from config
// if it is empty then try read from env variables
func (ns *Wireless) loadPath() {
	if ns.CMD == "" {
		ns.CMD = proc(CMD, "")
	}
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// try to read root path, or use default root path
	root := os.Getenv(CMD)
	if root == "" {
		root = CMD
	}
	return root
}

func init() {
	// this only works on Mac OS X, so if we're not running on Mac, punt.
	if runtime.GOOS != "darwin" {
		return
	}
	inputs.Add("mac_wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
