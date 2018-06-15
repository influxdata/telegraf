// +build linux

package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Meminfo struct {
	statFile string
}

func (k *Meminfo) Description() string {
	return "Get memory statistics from /proc/meminfo"
}

func (k *Meminfo) SampleConfig() string {
	return ""
}

func (k *Meminfo) Gather(acc telegraf.Accumulator) error {
	data, err := k.getMeminfo()
	if err != nil {
		return err
	}

	data = bytes.TrimRight(data, "\n")
	// Get rid of the :'s
	data = bytes.Replace(data, []byte(":"), []byte(""), -1)
	// Get rid of the kB's
	data = bytes.Replace(data, []byte("kB"), []byte(""), -1)
	// Change ('s to _'s
	data = bytes.Replace(data, []byte("("), []byte("_"), -1)
	// Get rid of the )'s
	data = bytes.Replace(data, []byte(")"), []byte(""), -1)

	dataFields := bytes.Fields(data)

	fields := make(map[string]interface{})

	for i, field := range dataFields {

		// dataFields is an array of {"stat1_name", "stat1_value", "stat2_name",
		// "stat2_value", ...}
		// We only want the even number index as that contain the stat name.
		if i%2 == 0 {
			// Convert the stat value into an integer.
			m, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}

			fields[string(field)] = int64(m)
		}
	}

	acc.AddFields("meminfo", fields, map[string]string{})
	return nil
}

func (k *Meminfo) getMeminfo() ([]byte, error) {
	if _, err := os.Stat(k.statFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("meminfo: %s does not exist!", k.statFile)
	} else if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(k.statFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func init() {
	inputs.Add("meminfo", func() telegraf.Input {
		return &Meminfo{
			statFile: "/proc/meminfo",
		}
	})
}
