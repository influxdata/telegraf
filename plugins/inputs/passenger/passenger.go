package passenger

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/net/html/charset"
)

type passenger struct {
	Command string
}

func (p *passenger) parseCommand() (string, []string) {
	var arguments []string
	if !strings.Contains(p.Command, " ") {
		return p.Command, arguments
	}

	arguments = strings.Split(p.Command, " ")
	if len(arguments) == 1 {
		return arguments[0], arguments[1:]
	}

	return arguments[0], arguments[1:]
}

type info struct {
	Passenger_version  string `xml:"passenger_version"`
	Process_count      int    `xml:"process_count"`
	Capacity_used      int    `xml:"capacity_used"`
	Get_wait_list_size int    `xml:"get_wait_list_size"`
	Max                int    `xml:"max"`
	Supergroups        struct {
		Supergroup []struct {
			Name               string `xml:"name"`
			Get_wait_list_size int    `xml:"get_wait_list_size"`
			Capacity_used      int    `xml:"capacity_used"`
			Group              []struct {
				Name                    string `xml:"name"`
				AppRoot                 string `xml:"app_root"`
				AppType                 string `xml:"app_type"`
				Enabled_process_count   int    `xml:"enabled_process_count"`
				Disabling_process_count int    `xml:"disabling_process_count"`
				Disabled_process_count  int    `xml:"disabled_process_count"`
				Capacity_used           int    `xml:"capacity_used"`
				Get_wait_list_size      int    `xml:"get_wait_list_size"`
				Processes_being_spawned int    `xml:"processes_being_spawned"`
				Processes               struct {
					Process []*process `xml:"process"`
				} `xml:"processes"`
			} `xml:"group"`
		} `xml:"supergroup"`
	} `xml:"supergroups"`
}

type process struct {
	Pid                   int    `xml:"pid"`
	Concurrency           int    `xml:"concurrency"`
	Sessions              int    `xml:"sessions"`
	Busyness              int    `xml:"busyness"`
	Processed             int    `xml:"processed"`
	Spawner_creation_time int64  `xml:"spawner_creation_time"`
	Spawn_start_time      int64  `xml:"spawn_start_time"`
	Spawn_end_time        int64  `xml:"spawn_end_time"`
	Last_used             int64  `xml:"last_used"`
	Uptime                string `xml:"uptime"`
	Code_revision         string `xml:"code_revision"`
	Life_status           string `xml:"life_status"`
	Enabled               string `xml:"enabled"`
	Has_metrics           bool   `xml:"has_metrics"`
	Cpu                   int64  `xml:"cpu"`
	Rss                   int64  `xml:"rss"`
	Pss                   int64  `xml:"pss"`
	Private_dirty         int64  `xml:"private_dirty"`
	Swap                  int64  `xml:"swap"`
	Real_memory           int64  `xml:"real_memory"`
	Vmsize                int64  `xml:"vmsize"`
	Process_group_id      string `xml:"process_group_id"`
}

func (p *process) getUptime() int64 {
	if p.Uptime == "" {
		return 0
	}

	timeSlice := strings.Split(p.Uptime, " ")
	var uptime int64
	uptime = 0
	for _, v := range timeSlice {
		switch {
		case strings.HasSuffix(v, "d"):
			iValue := strings.TrimSuffix(v, "d")
			value, err := strconv.ParseInt(iValue, 10, 64)
			if err == nil {
				uptime += value * (24 * 60 * 60)
			}
		case strings.HasSuffix(v, "h"):
			iValue := strings.TrimSuffix(v, "h")
			value, err := strconv.ParseInt(iValue, 10, 64)
			if err == nil {
				uptime += value * (60 * 60)
			}
		case strings.HasSuffix(v, "m"):
			iValue := strings.TrimSuffix(v, "m")
			value, err := strconv.ParseInt(iValue, 10, 64)
			if err == nil {
				uptime += value * 60
			}
		case strings.HasSuffix(v, "s"):
			iValue := strings.TrimSuffix(v, "s")
			value, err := strconv.ParseInt(iValue, 10, 64)
			if err == nil {
				uptime += value
			}
		}
	}

	return uptime
}

var sampleConfig = `
  ## Path of passenger-status.
  ##
  ## Plugin gather metric via parsing XML output of passenger-status
  ## More information about the tool:
  ##   https://www.phusionpassenger.com/library/admin/apache/overall_status_report.html
  ##
  ## If no path is specified, then the plugin simply execute passenger-status
  ## hopefully it can be found in your PATH
  command = "passenger-status -v --show=xml"
`

func (r *passenger) SampleConfig() string {
	return sampleConfig
}

func (r *passenger) Description() string {
	return "Read metrics of passenger using passenger-status"
}

func (g *passenger) Gather(acc telegraf.Accumulator) error {
	if g.Command == "" {
		g.Command = "passenger-status -v --show=xml"
	}

	cmd, args := g.parseCommand()
	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return err
	}

	if err = importMetric(out, acc); err != nil {
		return err
	}

	return nil
}

func importMetric(stat []byte, acc telegraf.Accumulator) error {
	var p info

	decoder := xml.NewDecoder(bytes.NewReader(stat))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&p); err != nil {
		return fmt.Errorf("Cannot parse input with error: %v\n", err)
	}

	tags := map[string]string{
		"passenger_version": p.Passenger_version,
	}
	fields := map[string]interface{}{
		"process_count":      p.Process_count,
		"max":                p.Max,
		"capacity_used":      p.Capacity_used,
		"get_wait_list_size": p.Get_wait_list_size,
	}
	acc.AddFields("passenger", fields, tags)

	for _, sg := range p.Supergroups.Supergroup {
		tags := map[string]string{
			"name": sg.Name,
		}
		fields := map[string]interface{}{
			"get_wait_list_size": sg.Get_wait_list_size,
			"capacity_used":      sg.Capacity_used,
		}
		acc.AddFields("passenger_supergroup", fields, tags)

		for _, group := range sg.Group {
			tags := map[string]string{
				"name":     group.Name,
				"app_root": group.AppRoot,
				"app_type": group.AppType,
			}
			fields := map[string]interface{}{
				"get_wait_list_size":      group.Get_wait_list_size,
				"capacity_used":           group.Capacity_used,
				"processes_being_spawned": group.Processes_being_spawned,
			}
			acc.AddFields("passenger_group", fields, tags)

			for _, process := range group.Processes.Process {
				tags := map[string]string{
					"group_name":       group.Name,
					"app_root":         group.AppRoot,
					"supergroup_name":  sg.Name,
					"pid":              fmt.Sprintf("%d", process.Pid),
					"code_revision":    process.Code_revision,
					"life_status":      process.Life_status,
					"process_group_id": process.Process_group_id,
				}
				fields := map[string]interface{}{
					"concurrency":           process.Concurrency,
					"sessions":              process.Sessions,
					"busyness":              process.Busyness,
					"processed":             process.Processed,
					"spawner_creation_time": process.Spawner_creation_time,
					"spawn_start_time":      process.Spawn_start_time,
					"spawn_end_time":        process.Spawn_end_time,
					"last_used":             process.Last_used,
					"uptime":                process.getUptime(),
					"cpu":                   process.Cpu,
					"rss":                   process.Rss,
					"pss":                   process.Pss,
					"private_dirty":         process.Private_dirty,
					"swap":                  process.Swap,
					"real_memory":           process.Real_memory,
					"vmsize":                process.Vmsize,
				}
				acc.AddFields("passenger_process", fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("passenger", func() telegraf.Input {
		return &passenger{}
	})
}
