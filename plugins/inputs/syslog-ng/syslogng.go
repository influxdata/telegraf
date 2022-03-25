package syslogng

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type runner func(cmdName string, timeout config.Duration, useSudo bool) (*bytes.Buffer, error)

type SyslogNg struct {
	Binary  string
	Timeout config.Duration
	UseSudo bool

	run runner
}

var defaultBinary = "/usr/local/sbin/syslog-ng-ctl"
var defaultTimeout = config.Duration(time.Second)

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  # use_sudo = false

  ## The default location of the syslog-ng-ctl binary can be overridden with:
  # binary = "/usr/local/sbin/syslog-ng-ctl"

  ## The default timeout of 1s can be overridden with:
  # timeout = "1s"
`

func SyslogNgCtlRunner(cmdName string, timeout config.Duration, useSudo bool) (*bytes.Buffer, error) {
	var out bytes.Buffer
	cmdArgs := []string{"stats"}

	cmd := exec.Command(cmdName, cmdArgs...)
	if useSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Duration(timeout))

	if err != nil {
		return &out, fmt.Errorf("error running syslog-ng-ctl: %w (%s, %v)", err, cmdName, cmdArgs)
	}

	return &out, nil
}

// Description displays what this plugin is about
func (s *SyslogNg) Description() string {
	return "A plugin to collect stats from the syslog-ng log daemon"
}

// SampleConfig displays configuration instructions
func (s *SyslogNg) SampleConfig() string {
	return sampleConfig
}

// Gather collects stats from syslog-ng-ctl and adds them to the Accumulator
func (s *SyslogNg) Gather(acc telegraf.Accumulator) error {
	out, err := s.run(s.Binary, s.Timeout, s.UseSudo)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %w", err)
	}

	reader := csv.NewReader(out)
	reader.Comma = ';'

	rows := []map[string]string{}
	header := []string{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("cannot read metrics: %w", err)
		}

		if len(header) == 0 {
			header = record
		} else {
			row := map[string]string{}
			for i := range header {
				row[header[i]] = record[i]
			}
			rows = append(rows, row)
		}
	}

	for _, row := range rows {
		number, _ := strconv.Atoi(row["Number"])

		switch row["Type"] {
		case "processed", "dropped", "queued", "suppressed", "discarded", "memory_usage", "matched", "not_matched", "written":
			tags := map[string]string{
				"type":        row["Type"],
				"source_name": row["SourceName"],
			}
			fields := map[string]interface{}{
				"number": number,
			}

			acc.AddFields("syslog-ng", fields, tags)
		default:
		}
	}

	return nil
}

func init() {
	inputs.Add("syslog-ng", func() telegraf.Input {
		return &SyslogNg{
			Binary:  defaultBinary,
			Timeout: defaultTimeout,
			UseSudo: false,

			run: SyslogNgCtlRunner,
		}
	})
}
