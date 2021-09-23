//go:build linux
// +build linux

package dmesg

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (

type DmesgConf struct {
	Filters []string `toml:"filters"`
}

func (k *DmesgConf) Description() string {
	return "Return counts of specific regexes against dmesg"
}

var dmesgSampleConfig = `
	## some basic dmesg regexes
	# filters = ["oom_reaper|Out of memory", "Power-on or device reset occurred", "I/O error", "MCE MEMORY"]
`

func (k *DmesgConf) SampleConfig() string {
	return dmesgSampleConfig
}

func (k *DmesgConf) Gather(acc telegraf.Accumulator) error {
}

func init() {
	inputs.Add("dmesg", func() telegraf.Input { return &DmesgConf{} })
}
