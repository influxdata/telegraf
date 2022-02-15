package wireless_mac

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// default file paths
const (
	OSXCMD       = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I"
	NET_WIRELESS = "/net/wireless"
)

type Wireless_Mac struct {
	DumpZeros   bool   `toml:"dump_zeros"`
	NetWireless string `toml:"wireless_cmd"`
	HostProc string          `toml:"host_proc"`
	Log      telegraf.Logger `toml:"-"`
}

// WirelessData struct to hold the tags, headers (measurements) and data
type Wireless_MacData struct {
	Headers []string
	Data    [][]int64
	Tags    []string
}


func exe_cmd(cmd string, wg *sync.WaitGroup) ([]byte, error) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]
	out, err := exec.Command(head, parts...).Output()
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return out, err
}

func (ns *Wireless_Mac) Gather(acc telegraf.Accumulator) error {
	if runtime.GOOS == "darwin" {
		// collect MAC OS wireless data
		wg := new(sync.WaitGroup)
		wg.Add(3)
		wireless, err := exe_cmd(OSXCMD, wg)
		if err != nil {
			return err
		}
		err = ns.gatherMacWireless(wireless, acc)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("OS Not Supported")
}

func (ns *Wireless_Mac) gatherMacWireless(data []byte, acc telegraf.Accumulator) error {
	metrics, tags, err := loadMacWirelessTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	acc.AddFields("wireless_mac", metrics, tags)
	return nil
}

func loadMacWirelessTable(table []byte, dumpZeros bool) (map[string]interface{}, map[string]string, error) {
	metrics := map[string]interface{}{}
	tags := map[string]string{}
	myLines := strings.Split(string(table), "\n")
	for _, line := range myLines {
		if len(line) == 0 {
			continue
		}
		f1 := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
		f2 := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		if f1 == "BSSID" {
			tags[strings.Replace(strings.Trim(f1, " "), " ", "_", -1)] = strings.Replace(strings.Trim(string(f2), " "), " ", "_", -1)
			continue
		}
		n, err := strconv.ParseInt(strings.Trim(f2, " "), 10, 64)
		if err != nil {
			tags[strings.Replace(strings.Trim(f1, " "), " ", "_", -1)] = strings.Replace(strings.Trim(f2, " "), " ", "_", -1)
			continue
		}
		if n == 0 {
			if dumpZeros {
				continue
			}
		}
		metrics[strings.Trim(f1, " ")] = n

	}
	tags["interface"] = "airport"
	return metrics, tags, nil

}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// try to read root path, or use default root path
	root := os.Getenv(OSXCMD)
	if root == "" {
		root = OSXCMD
	}
	return root
}

func init() {
	inputs.Add("wireless_mac", func() telegraf.Input {
		return &Wireless_Mac{}
	})
}
