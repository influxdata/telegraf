package ipset

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ipsets is a telegraf plugin to gather packets and bytes counters from ipset
type Ipset struct {
	IncludeUnmatchedSets bool
	UseSudo              bool
	Timeout              config.Duration
	lister               setLister
}

type setLister func(Timeout config.Duration, UseSudo bool) (*bytes.Buffer, error)

const measurement = "ipset"

var defaultTimeout = config.Duration(time.Second)

func (i *Ipset) Init() error {
	_, err := exec.LookPath("ipset")
	if err != nil {
		return err
	}

	return nil
}

func (i *Ipset) Gather(acc telegraf.Accumulator) error {
	out, e := i.lister(i.Timeout, i.UseSudo)
	if e != nil {
		acc.AddError(e)
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignore sets created without the "counters" option
		nocomment := strings.Split(line, "\"")[0]
		if !(strings.Contains(nocomment, "packets") &&
			strings.Contains(nocomment, "bytes")) {
			continue
		}

		data := strings.Fields(line)
		if len(data) < 7 {
			acc.AddError(fmt.Errorf("error parsing line (expected at least 7 fields): %s", line))
			continue
		}
		if data[0] == "add" && (data[4] != "0" || i.IncludeUnmatchedSets) {
			tags := map[string]string{
				"set":  data[1],
				"rule": data[2],
			}
			packetsTotal, err := strconv.ParseUint(data[4], 10, 64)
			if err != nil {
				acc.AddError(err)
			}
			bytesTotal, err := strconv.ParseUint(data[6], 10, 64)
			if err != nil {
				acc.AddError(err)
			}
			fields := map[string]interface{}{
				"packets_total": packetsTotal,
				"bytes_total":   bytesTotal,
			}
			acc.AddCounter(measurement, fields, tags)
		}
	}
	return nil
}

func setList(timeout config.Duration, useSudo bool) (*bytes.Buffer, error) {
	// Is ipset installed ?
	ipsetPath, err := exec.LookPath("ipset")
	if err != nil {
		return nil, err
	}
	var args []string
	cmdName := ipsetPath
	if useSudo {
		cmdName = "sudo"
		args = append(args, ipsetPath)
	}
	args = append(args, "save")

	cmd := exec.Command(cmdName, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running ipset save: %s", err)
	}

	return &out, nil
}

func init() {
	inputs.Add("ipset", func() telegraf.Input {
		return &Ipset{
			lister:  setList,
			Timeout: defaultTimeout,
		}
	})
}
