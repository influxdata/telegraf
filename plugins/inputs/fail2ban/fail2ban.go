package fail2ban

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.
)

type Fail2ban struct {
	path    string
	UseSudo bool
}

var metricsTargets = []struct {
	target string
	field  string
}{
	{
		target: "Currently failed:",
		field:  "failed",
	},
	{
		target: "Currently banned:",
		field:  "banned",
	},
}

func (f *Fail2ban) Gather(acc telegraf.Accumulator) error {
	if len(f.path) == 0 {
		return errors.New("fail2ban-client not found: verify that fail2ban is installed and that fail2ban-client is in your PATH")
	}

	name := f.path
	var arg []string

	if f.UseSudo {
		name = "sudo"
		arg = append(arg, f.path)
	}

	args := append(arg, "status")

	cmd := execCommand(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	lines := strings.Split(string(out), "\n")
	const targetString = "Jail list:"
	var jails []string
	for _, line := range lines {
		idx := strings.LastIndex(line, targetString)
		if idx < 0 {
			// not target line, skip.
			continue
		}
		jails = strings.Split(strings.TrimSpace(line[idx+len(targetString):]), ", ")
		break
	}

	for _, jail := range jails {
		fields := make(map[string]interface{})
		args := append(arg, "status", jail)
		cmd := execCommand(name, args...)
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			key, value := extractCount(line)
			if key != "" {
				fields[key] = value
			}
		}
		acc.AddFields("fail2ban", fields, map[string]string{"jail": jail})
	}
	return nil
}

func extractCount(line string) (string, int) {
	for _, metricsTarget := range metricsTargets {
		idx := strings.LastIndex(line, metricsTarget.target)
		if idx < 0 {
			continue
		}
		ban := strings.TrimSpace(line[idx+len(metricsTarget.target):])
		banCount, err := strconv.Atoi(ban)
		if err != nil {
			return "", -1
		}
		return metricsTarget.field, banCount
	}
	return "", -1
}

func init() {
	f := Fail2ban{}
	path, _ := exec.LookPath("fail2ban-client")
	if len(path) > 0 {
		f.path = path
	}
	inputs.Add("fail2ban", func() telegraf.Input {
		f := f
		return &f
	})
}
