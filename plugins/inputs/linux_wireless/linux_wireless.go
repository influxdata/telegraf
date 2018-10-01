package linux_wireless

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
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
	ENV_ROOT     = "PROC_ROOT"
	ENV_WIRELESS = "PROC_NET_WIRELESS"
)

type Wireless struct {
	ProcNetWireless string `toml:"proc_net_wireless"`
	DumpZeros       bool   `toml:"dump_zeros"`
}

// WirelessDataa struct to hold the tags, headers (measurements) and data
type WirelessData struct {
	Headers []string
	Data    [][]int64
	Tags    []string
}

var sampleConfig = `
  ## file path for proc file. If empty default path will be used:
  ##    /proc/net/wireless
  ## This can also be overridden with env variable, see README.
  proc_net_wireless = "/proc/net/wireless"
  ## dump metrics with 0 values too
  dump_zeros       = true
`

func (ns *Wireless) Description() string {
	return "Collect wireless interface link quality metrics"
}

func (ns *Wireless) SampleConfig() string {
	return sampleConfig
}

func (ns *Wireless) Gather(acc telegraf.Accumulator) error {
	// load paths, get from env if config values are empty
	ns.loadPath()
	// collect wireless data
	wireless, err := ioutil.ReadFile(ns.ProcNetWireless)
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
	wirelessData, err := loadWirelessTable(data, ns.DumpZeros)
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

func loadWirelessTable(table []byte, dumpZeros bool) (WirelessData, error) {
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
	header_fields[0] = strings.ToLower(strings.Replace(h1[1]+h2[1], "-", "", -1))
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
	for x := 2; x < len(myLines)-1; x++ {
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
		fmt.Println("Data ", x-2, ": ", data[x-2])
	}
	// Now fill out the Wireless struct and return it
	wd.Headers = header_fields
	wd.Tags = tags
	wd.Data = data
	return wd, nil
}

// loadPath can be used to read path firstly from config
// if it is empty then try read from env variables
func (ns *Wireless) loadPath() {
	if ns.ProcNetWireless == "" {
		ns.ProcNetWireless = proc(ENV_WIRELESS, NET_WIRELESS)
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
	// this only works on linux, so if we're not running on Linux, punt.
	if runtime.GOOS != "linux" {
		return
	}
	inputs.Add("linux_wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
