package htcondor

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type HTCondor struct {
}

// SampleConfig returns sample configuration options.
func (htc *HTCondor) SampleConfig() string {
	return ""
}

// Description returns a short description of the plugin.
func (htc *HTCondor) Description() string {
	return "Gather outputs from condor_q command"
}

const measurement = "htcondor"

var condorOutputRegex = regexp.MustCompile(`(?m)(?P<jobs>\d+\s*jobs);\s*(?P<completed>\d+\s*completed),\s*(?P<removed>\d+\s*removed),\s*(?P<idle>\d+\s*idle),\s*(?P<running>\d+\s*running),\s*(?P<held>\d+\s*held),\s*(?P<suspended>\d+\s*suspended)`)

// Gather outputs.
func (htc *HTCondor) Gather(acc telegraf.Accumulator) error {
	c := exec.Command("condor_q")
	out, err := c.Output()

	// htcondor notfound
	if err != nil {
		return err
	}

	var regexGroupMatch = condorOutputRegex.FindAllStringSubmatch(string(out), -1)
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for i := 1; i < len(regexGroupMatch[0]); i++ {
		var matched = strings.Split(regexGroupMatch[0][i], " ") // "1 jobs" --> ["1", "jobs"]
		var fieldKey = matched[1]
		var fieldvalue, _ = strconv.ParseInt(matched[0], 10, 64)
		fields[fieldKey] = fieldvalue
	}
	acc.AddFields(measurement, fields, tags)

	return nil
}

func init() {
	inputs.Add("htcondor", func() telegraf.Input { return &HTCondor{} })
}
