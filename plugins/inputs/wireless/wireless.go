package wireless

import (
	"errors"
	"io/ioutil"
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
	NET_PROC     = "/proc"
)

type Wireless struct {
	DumpZeros   bool   `toml:"dump_zeros"`
	NetWireless string `toml:"wireless_cmd"`
}

// WirelessDataa struct to hold the tags, headers (measurements) and data
type WirelessData struct {
	Headers []string
	Data    [][]int64
	Tags    []string
}

var sampleConfig = `
  [[inputs.wireless]]
  ## dump metrics with 0 values too
  dump_zeros       = true
`

func (ns *Wireless) Description() string {
	return "Collect wireless interface metrics"
}

func (ns *Wireless) SampleConfig() string {
	return sampleConfig
}
func exe_cmd(cmd string, wg *sync.WaitGroup) ([]byte, error) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]
	out, err := exec.Command(head, parts...).Output()
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return out, err
}

func (ns *Wireless) Gather(acc telegraf.Accumulator) error {
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
	} else if runtime.GOOS == "linux" {
		// collect wireless data
		wireless, err := ioutil.ReadFile(ns.NetWireless)
		if err != nil {
			return err
		}
		err = ns.gatherLinuxWireless(wireless, acc)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("OS Not Supported")
}

func (ns *Wireless) gatherMacWireless(data []byte, acc telegraf.Accumulator) error {
	metrics, tags, err := loadMacWirelessTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	acc.AddFields("wireless", metrics, tags)
	return nil
}

func loadMacWirelessTable(table []byte, dumpZeros bool) (map[string]interface{}, map[string]string, error) {
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

func (ns *Wireless) gatherLinuxWireless(data []byte, acc telegraf.Accumulator) error {
	wirelessData, err := loadLinuxWirelessTable(data, ns.DumpZeros)
	if err != nil {
		return err
	}
	// go through the WirelessData struct and create maps for
	// Telegraf to deal with, then addd the data to the
	// telegraf accumulator
	for x := 0; x < len(wirelessData.Tags); x++ {
		entries := map[string]interface{}{}
		tags := map[string]string{
			"interface": wirelessData.Tags[x],
		}
		if len(wirelessData.Headers) == len(wirelessData.Data[x]) && len(wirelessData.Data) == len(wirelessData.Tags) {
			for z := 0; z < len(wirelessData.Data[x]); z++ {
				entries[wirelessData.Headers[z]] = wirelessData.Data[x][z]
			}
			acc.AddFields("wireless", entries, tags)
		} else {
			return errors.New("Invalid field lengths returned.")
		}
	}
	return nil
}
func loadLinuxWirelessTable(table []byte, dumpZeros bool) (WirelessData, error) {
	wd := WirelessData{}
	var value int64
	var err error
	myLines := strings.Split(string(table), "\n")
	if len(myLines) < 2 {
		return WirelessData{}, errors.New("Error gathering Wireless Data")
	}
	// split on '|' and trim the spaces
	h1 := strings.Split(myLines[0], "|")
	h2 := strings.Split(myLines[1], "|")
	if len(h2) < len(h1) || len(h2) < 2 {
		return WirelessData{}, errors.New("Invalid header lengths returned.")
	}
	header_fields := make([]string, 11)
	header_count := 1
	// we'll collect the data and tags in here for now
	tags := make([]string, len(myLines)-2)
	data := make([][]int64, len(myLines)-2)
	// trim out all the spaces.
	for x := 0; x < len(h1); x++ {
		h1[x] = strings.Trim(h1[x], " ")
		h2[x] = strings.Trim(h2[x], " ")
	}
	// first 2 headers have a '-' in them, so join those and remove the '-'
	// also, ignore the first one, since it is the interface name
	header_fields[0] = strings.ToLower(strings.Replace(strings.Trim(h1[1], " ")+strings.Trim(h2[1], " "), "-", "", -1))
	// next headers are composed with sub-headers, so build those.
	for y := 2; y < len(h1)-2; y++ {
		tmpStr := strings.Split(h2[y], " ")
		for z := 0; z < len(tmpStr); z++ {
			if tmpStr[z] == "" {
				continue
			}
			header_fields[header_count] = strings.ToLower(strings.Replace(h1[y]+"_"+tmpStr[z], " ", "_", -1))
			header_count++
		}
	}
	// last 2 are simple multi-line headers, so join them
	for t := len(h1) - 2; t < len(h1); t++ {
		header_fields[header_count] = strings.ToLower(h1[t] + "_" + h2[t])
		header_count++
	}
	// now let's go through the data and save it for return.
	// if we're dumping zeros, we will also dump the header for the
	// zero data.
	for x := 2; x <= len(myLines)-1; x++ {
		data_count := 0
		metrics := strings.Fields(myLines[x])
		sub_data := make([]int64, len(metrics))
		for z := 0; z < len(metrics)-2; z++ {
			if strings.Index(metrics[z], ":") > 0 {
				tags[x-2] = metrics[z]
			} else {
				if metrics[z] == "0" {
					if dumpZeros {
						continue
					}
				}

				// clean up the string as they have extraneous characters in them
				value, err = strconv.ParseInt(strings.Replace(metrics[z], ".", "", -1), 10, 64)
				if err == nil {
					sub_data[data_count] = value
					data_count++
				}

			}
		}
		data[x-2] = sub_data
	}
	// Now fill out the Wireless struct and return it
	wd.Headers = header_fields
	wd.Tags = tags
	wd.Data = data
	return wd, nil
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
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
