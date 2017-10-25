// +build linux

package pidstat

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	firstTimestamp time.Time
	execCommand    = exec.Command // execCommand is used to mock commands in tests.
)

// ---- TELEGRAF INTERFACE

type Pidstat struct {

	// Processes which to monitor. None=all
	Programs []string

	// DeviceTags adds the possibility to add additional tags for devices.
	Interval int

	PerPid     bool
	PerCommand bool
}

func (*Pidstat) Description() string {
	return "Pidstat metrics collector"
}

var sampleConfig = `
  ## Gather metrics per pid or per command name
  ## metrics pidstat_pid and pidstat_command respectively
  # PerPid = true
  # PerCommand = true
  #
  ## Which command names to track. None = all
  # Programs = []
  ## Programs = [ "ngnix", "sh" ]
`

func (*Pidstat) SampleConfig() string {
	return sampleConfig
}

func (p *Pidstat) Gather(acc telegraf.Accumulator) error {
	if p.Interval == 0 {
		if firstTimestamp.IsZero() {
			firstTimestamp = time.Now()
		} else {
			p.Interval = int(time.Since(firstTimestamp).Seconds() + 0.5)
		}
	}
	stats_pids := make(map[string]Pidstat_record)
	stats_commands := make(map[string]Pidstat_record)
	err := collect(1, p.Programs, stats_pids, stats_commands)
	if err != nil {
		return fmt.Errorf("collect returned error: %s", err)
	}

	ts := time.Now().Add(time.Duration(p.Interval) * time.Second)

	if p.PerPid {
		feed_accumulator(acc, stats_pids, ts)
	}
	if p.PerCommand {
		feed_accumulator(acc, stats_commands, ts)
	}

	return nil
}

func init() {
	p := Pidstat{}
	inputs.Add("pidstat", func() telegraf.Input {
		return &p
	})
}

// ---- ROW PARSING

func tokenize_sanitize(src string) []string {

	src_split := strings.Split(src, " ")
	tokens := make([]string, 0)

	for _, f := range src_split {
		if len(f) > 0 {
			tokens = append(tokens, f)
		}
	}

	return tokens
}

func tokenize_sanitize_header(src string) []string {

	src_split := strings.Split(src, " ")
	tokens := make([]string, 0)

	for _, f := range src_split {
		if len(f) > 0 {
			tokens = append(tokens, escape(f))
		}
	}

	return tokens
}

func parse_general_info(src []string) []string {
	out := make([]string, 0)
	split := strings.Split(src[0], " ")
	for _, s := range split {
		if len(s) != 0 {
			out = append(out, strings.TrimSpace(s))
		}
	}
	for i := 1; i < len(src); i++ {
		out = append(out, strings.TrimSpace(src[i]))
	}
	return out
}

func parse_row(row []string, header []string) map[string]string {
	out := make(map[string]string, 0)

	if len(row) > len(header) {
		i := 0
		for i, h := range header {
			out[h] = row[i]
		}
		for i = len(header); i < len(row); i++ {
			out[header[len(header)-1]] += " " + row[i]
		}
	} else {
		for i, r := range row {
			out[header[i]] = r
		}
	}

	return out
}

// escape removes % and / chars in field names
func escape(dirty string) string {
	var fieldEscaper = strings.NewReplacer(
		`%`, "pct_",
		`/`, "_per_",
	)
	return fieldEscaper.Replace(dirty)
}

type Pidstat_record struct {
	measurement string
	fields      map[string]interface{}
	tags        map[string]string
}

// ---- /ROW PARSING
// ---- UTIL

func copy_map(src map[string]string) map[string]interface{} {
	out := make(map[string]interface{})
	for i, s := range src {
		out[i] = s
	}
	return out
}

func copy_map_string(src map[string]string) map[string]string {
	out := make(map[string]string)
	for i, s := range src {
		out[i] = s
	}
	return out
}

// ---- /UTIL
// ---- PRINTING

func print_stats(m map[string]Pidstat_record) {
	for k, v := range m {
		fmt.Printf("%s:", k)
		fmt.Printf("\t%s\n", v.measurement)
		fmt.Printf("\tfields:\n")
		for k_f, v_f := range v.fields {
			fmt.Printf("\t\t%s: %s\n", k_f, v_f)
		}
		fmt.Printf("\ttags:\n")
		for k_t, v_t := range v.tags {
			fmt.Printf("\t\t%s: %s\n", k_t, v_t)
		}
	}
}

func array_boxes(src []string) {
	for i, value := range src {
		fmt.Printf("\t%s: |%s|\n", i, value)
	}
}

// ---- /PRINTING
// ---- DATASTRUCTURE FEEDING

func update_pidstats(stats map[string]Pidstat_record,
	values map[string]interface{}, tags map[string]string) {

	for i, v := range values {
		values[i], _ = strconv.ParseFloat(v.(string), 64)
	}

	if rec, ok := stats[tags["PID"]]; ok {
		for k, v := range values {
			rec.fields[k] = v
		}
	} else {
		stats[tags["PID"]] = Pidstat_record{"pidstat_pid", values, tags}
	}
}

func update_cmdstats(stats map[string]Pidstat_record,
	values map[string]interface{}, tags map[string]string) {
	for i, v := range values {
		values[i], _ = strconv.ParseFloat(v.(string), 64)
	}

	if rec, ok := stats[tags["Command"]]; ok {
		for k, v := range values {
			if val, ok_rec := rec.fields[k]; ok_rec {
				//old_float, _ := strconv.ParseFloat(val.(string), 64)
				//new_float, _ := strconv.ParseFloat(v.(string), 64)
				//rec.fields[k] = strconv.FormatFloat( old_float + new_float, 'f', 6, 64 );
				rec.fields[k] = val.(float64) + v.(float64)
			} else {
				rec.fields[k] = v

			}
		}
	} else {
		stats[tags["Command"]] = Pidstat_record{"pidstat_command", values, tags}
	}
}

func feed_accumulator(acc telegraf.Accumulator, stats map[string]Pidstat_record,
	ts time.Time) {
	for _, rec := range stats {
		acc.AddFields(rec.measurement, rec.fields, rec.tags, ts)
	}
}

func stats_from_string(inp []byte, stat_pids map[string]Pidstat_record, stat_commands map[string]Pidstat_record) error {

	r := bytes.NewReader(inp)
	//fmt.Printf("%s", inp)

	csvreader := csv.NewReader(r)
	csvreader.Comma = '\t'
	//csvreader.FieldsPerRecord = num_fields[i]

	//general info
	//os, version, sys-name, date, arch, cores
	record, err := csvreader.Read()

	if err != nil {
		return fmt.Errorf("csvread failed to read general pidstat info (first row): %s from source: %s", err, inp)
	}

	tags_base := make(map[string]string)
	header_general := []string{"os", "os_ver", "sys_name", "date", "arch", "cores"}
	general_info := parse_general_info(record)

	for i, t := range header_general {
		tags_base[t] = general_info[i]
	}

	//table header
	record, err = csvreader.Read()

	//if err != nil {
	//return fmt.Errorf( "csvread failed to read table header (second row): %s", err )
	//}

	header := tokenize_sanitize_header(record[0])
	if len(header) >= 4 {
		header[0] = "time"
		header[1] = "part_of_day"
		header[2] = "UID"
		header[3] = "PID"
	}

	for {
		row, err := csvreader.Read()
		if err == io.EOF {
			return nil
		}

		if row != nil {
			//fmt.Println("nil row")
			row = tokenize_sanitize(row[0])

			if row != nil {
				//fmt.Println("nil tokenized row")

				parsed_row := parse_row(row, header)

				tags := copy_map_string(tags_base)
				tags["PID"] = parsed_row["PID"]
				tags["UID"] = parsed_row["UID"]
				tags["Command"] = parsed_row["Command"]

				delete(parsed_row, "PID")
				delete(parsed_row, "UID")
				delete(parsed_row, "Command")
				delete(parsed_row, "time")
				delete(parsed_row, "part_of_day")

				delete(tags, "date")

				update_pidstats(stat_pids, copy_map(parsed_row), copy_map_string(tags))

				delete(tags, "PID")

				update_cmdstats(stat_commands, copy_map(parsed_row), copy_map_string(tags))
			}
		}
	}
	return nil
}

// ---- /DATASTRUCTURE FEEDING
// ---- MAIN

func collect(timeout int, Programs []string, stat_pids map[string]Pidstat_record, stat_commands map[string]Pidstat_record) error {

	extra_args := []string{"-d", "-l"}

	if len(Programs) > 0 {
		pstr := ""
		for _, p := range Programs {
			pstr += p + "|"
		}
		pstr = pstr[:len(pstr)-1]
		extra_args = append(extra_args, "-C")
		extra_args = append(extra_args, pstr)
	}

	extra_args[0] = "-d"
	cpu := execCommand("pidstat", extra_args...)
	extra_args[0] = "-r"
	mem := execCommand("pidstat", extra_args...)
	extra_args[0] = "-u"
	stack := execCommand("pidstat", extra_args...)
	extra_args[0] = "-w"
	rw := execCommand("pidstat", extra_args...)
	extra_args[0] = "-v"
	context := execCommand("pidstat", extra_args...)
	extra_args[0] = "-s"
	thread := execCommand("pidstat", extra_args...)

	commands := []*exec.Cmd{cpu, mem, stack, rw, context, thread}

	//num_fields := []int{ 10, 10, 7, 8, 6, 6 }

	for _, c := range commands {

		out, err := internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout))

		if err != nil {
			return fmt.Errorf("CombinedOutputTimeout returned error: %s", err)
		}
		err = stats_from_string(out, stat_pids, stat_commands)
		if err != nil {
			return fmt.Errorf("Failed to parse stats from string %s", err)
		}

	}

	return nil
}
