// +build linux

package mdadm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	statusLineRE      = regexp.MustCompile(`(\d+) blocks .*\[(\d+)/(\d+)\] \[[U_]+\]`)
	recoveryLineRE    = regexp.MustCompile(`\((\d+)/\d+\)`)
	componentDeviceRE = regexp.MustCompile(`(.*)\[\d+\]`)
)

type mdadmStat struct {
	statFile string
}

func (k *mdadmStat) Description() string {
	return "Get md array statistics from /proc/mdstat"
}

func (k *mdadmStat) SampleConfig() string {
	return ""
}

func (k *mdadmStat) Gather(acc telegraf.Accumulator) error {
	data, err := k.getProcMdstat()
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	lines := srings.Split(string(data), "\n")
	for _, line := range lines {
		/*
			Skip lines that have no useful data in them
		*/
		if strings.TrimSpace(line) == "" || line[0] == ' ' ||
			strings.HasPrefix(line, "Personalities") ||
			strings.HasPrefix(line, "unused") {
			continue
		}

	}

	acc.AddFields("mdadm", fields, map[string]string{})
	return nil
}

func (k *mdadmStat) getProcMdstat() ([]byte, error) {
	if _, err := os.Stat(k.statFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("mdadm: %s does not exist", k.statFile)
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
	inputs.Add("mdadm", func() telegraf.Input {
		return &mdadmStat{
			statFile: "/proc/mdstat",
		}
	})
}
