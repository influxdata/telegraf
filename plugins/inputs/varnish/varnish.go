//go:build !windows
// +build !windows

package varnish

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	measurementNamespace = "varnish"
	defaultStats         = []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"}
	defaultStatBinary    = "/usr/bin/varnishstat"
	defaultAdmBinary     = "/usr/bin/varnishadm"
	defaultTimeout       = config.Duration(time.Second)

	//vcl name and backend restriction regexp [A-Za-z][A-Za-z0-9_-]*
	defaultRegexps = []*regexp.Regexp{
		//dynamic backends
		//VBE.VCL_xxxx_xxx_VOD_SHIELD_Vxxxxxxxxxxxxx_xxxxxxxxxxxxx.goto.000007c8.(xx.xx.xxx.xx).(http://xxxxxxx-xxxxx-xxxxx-xxxxxx-xx-xxxx-x-xxxx.xx-xx-xxxx-x.amazonaws.com:80).(ttl:5.000000).fail_eaddrnotavail
		regexp.MustCompile(`^VBE\.(?P<_vcl>[\w\-]*)\.goto\.[[:alnum:]]+\.\((?P<backend>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\)\.\((?P<server>.*)\)\.\(ttl:\d*\.\d*.*\)`),

		//VBE.reload_20210622_153544_23757.default.unhealthy
		regexp.MustCompile(`^VBE\.(?P<_vcl>[\w\-]*)\.(?P<backend>[\w\-]*)\.([\w\-]*)`),

		//KVSTORE values
		regexp.MustCompile(`^KVSTORE\.(?P<id>[\w\-]*)\.(?P<_vcl>[\w\-]*)\.([\w\-]*)`),

		//XCNT.abc1234.XXX+_YYYY.cr.pass.val
		regexp.MustCompile(`^XCNT\.(?P<_vcl>[\w\-]*)(\.)*(?P<group>[\w\-.+]*)\.(?P<_field>[\w\-.+]*)\.val`),

		//generic metric like MSE_STORE.store-1-1.g_aio_running_bytes_write
		regexp.MustCompile(`([\w\-]*)\.(?P<_field>[\w\-.]*)`),
	}
)

type runner func(cmdName string, useSudo bool, args []string, timeout config.Duration) (*bytes.Buffer, error)

// Varnish is used to store configuration values
type Varnish struct {
	Stats         []string
	Binary        string
	BinaryArgs    []string
	AdmBinary     string
	AdmBinaryArgs []string
	UseSudo       bool
	InstanceName  string
	Timeout       config.Duration
	Regexps       []string
	MetricVersion int

	filter          filter.Filter
	run             runner
	admRun          runner
	regexpsCompiled []*regexp.Regexp
}

// Shell out to varnish cli and return the output
func varnishRunner(cmdName string, useSudo bool, cmdArgs []string, timeout config.Duration) (*bytes.Buffer, error) {
	cmd := exec.Command(cmdName, cmdArgs...)

	if useSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmdArgs = append([]string{"-n"}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	err := internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running %s %v - %s", cmdName, cmdArgs, err)
	}

	return &out, nil
}

func (s *Varnish) Init() error {
	var customRegexps []*regexp.Regexp
	for _, re := range s.Regexps {
		compiled, err := regexp.Compile(re)
		if err != nil {
			return fmt.Errorf("error parsing regexp: %s", err)
		}
		customRegexps = append(customRegexps, compiled)
	}
	s.regexpsCompiled = append(customRegexps, s.regexpsCompiled...)
	return nil
}

// Gather collects the configured stats from varnish_stat and adds them to the
// Accumulator
//
// The prefix of each stat (eg MAIN, MEMPOOL, LCK, etc) will be used as a
// 'section' tag and all stats that share that prefix will be reported as fields
// with that tag
func (s *Varnish) Gather(acc telegraf.Accumulator) error {
	if s.filter == nil {
		var err error
		if len(s.Stats) == 0 {
			s.filter, err = filter.Compile(defaultStats)
		} else {
			// legacy support, change "all" -> "*":
			if s.Stats[0] == "all" {
				s.Stats[0] = "*"
			}
			s.filter, err = filter.Compile(s.Stats)
		}
		if err != nil {
			return err
		}
	}

	admArgs, statsArgs := s.prepareCmdArgs()

	statOut, err := s.run(s.Binary, s.UseSudo, statsArgs, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	if s.MetricVersion == 2 {
		//run varnishadm to get active vcl
		var activeVcl = "boot"
		if s.admRun != nil {
			admOut, err := s.admRun(s.AdmBinary, s.UseSudo, admArgs, s.Timeout)
			if err != nil {
				return fmt.Errorf("error gathering metrics: %s", err)
			}
			activeVcl, err = getActiveVCLJson(admOut)
			if err != nil {
				return fmt.Errorf("error gathering metrics: %s", err)
			}
		}
		return s.processMetricsV2(activeVcl, acc, statOut)
	}
	return s.processMetricsV1(acc, statOut)
}

// Prepare varnish cli tools arguments
func (s *Varnish) prepareCmdArgs() ([]string, []string) {
	//default varnishadm arguments
	admArgs := []string{"vcl.list", "-j"}

	//default varnish stats arguments
	statsArgs := []string{"-j"}
	if s.MetricVersion == 1 {
		statsArgs = []string{"-1"}
	}

	//add optional instance name
	if s.InstanceName != "" {
		statsArgs = append(statsArgs, []string{"-n", s.InstanceName}...)
		admArgs = append([]string{"-n", s.InstanceName}, admArgs...)
	}

	//override custom arguments
	if len(s.AdmBinaryArgs) > 0 {
		admArgs = s.AdmBinaryArgs
	}
	//override custom arguments
	if len(s.BinaryArgs) > 0 {
		statsArgs = s.BinaryArgs
	}
	return admArgs, statsArgs
}

func (s *Varnish) processMetricsV1(acc telegraf.Accumulator, out *bytes.Buffer) error {
	sectionMap := make(map[string]map[string]interface{})
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		cols := strings.Fields(scanner.Text())
		if len(cols) < 2 {
			continue
		}
		if !strings.Contains(cols[0], ".") {
			continue
		}

		stat := cols[0]
		value := cols[1]

		if s.filter != nil && !s.filter.Match(stat) {
			continue
		}

		parts := strings.SplitN(stat, ".", 2)
		section := parts[0]
		field := parts[1]

		// Init the section if necessary
		if _, ok := sectionMap[section]; !ok {
			sectionMap[section] = make(map[string]interface{})
		}

		var err error
		sectionMap[section][field], err = strconv.ParseUint(value, 10, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("expected a numeric value for %s = %v", stat, value))
		}
	}

	for section, fields := range sectionMap {
		tags := map[string]string{
			"section": section,
		}
		if len(fields) == 0 {
			continue
		}

		acc.AddFields("varnish", fields, tags)
	}
	return nil
}

// metrics version 2 - parsing json
func (s *Varnish) processMetricsV2(activeVcl string, acc telegraf.Accumulator, out *bytes.Buffer) error {
	rootJSON := make(map[string]interface{})
	dec := json.NewDecoder(out)
	dec.UseNumber()
	if err := dec.Decode(&rootJSON); err != nil {
		return err
	}
	countersJSON := getCountersJSON(rootJSON)
	timestamp := time.Now()
	for fieldName, raw := range countersJSON {
		if fieldName == "timestamp" {
			continue
		}
		if s.filter != nil && !s.filter.Match(fieldName) {
			continue
		}
		data, ok := raw.(map[string]interface{})
		if !ok {
			acc.AddError(fmt.Errorf("unexpected data from json: %s: %#v", fieldName, raw))
			continue
		}

		var metricValue interface{}
		var parseError error
		flag := data["flag"]

		if value, ok := data["value"]; ok {
			if number, ok := value.(json.Number); ok {
				//parse bitmap value
				if flag == "b" {
					if metricValue, parseError = strconv.ParseUint(number.String(), 10, 64); parseError != nil {
						parseError = fmt.Errorf("%s value uint64 error: %s", fieldName, parseError)
					}
				} else if metricValue, parseError = number.Int64(); parseError != nil {
					//try parse float
					if metricValue, parseError = number.Float64(); parseError != nil {
						parseError = fmt.Errorf("stat %s value %v is not valid number: %s", fieldName, value, parseError)
					}
				}
			} else {
				metricValue = value
			}
		}

		if parseError != nil {
			acc.AddError(parseError)
			continue
		}

		metric := s.parseMetricV2(fieldName)
		if metric.vclName != "" && activeVcl != "" && metric.vclName != activeVcl {
			//skip not active vcl
			continue
		}

		fields := make(map[string]interface{})
		fields[metric.fieldName] = metricValue
		switch flag {
		case "c", "a":
			acc.AddCounter(metric.measurement, fields, metric.tags, timestamp)
		case "g":
			acc.AddGauge(metric.measurement, fields, metric.tags, timestamp)
		default:
			acc.AddGauge(metric.measurement, fields, metric.tags, timestamp)
		}
	}
	return nil
}

// Parse the output of "varnishadm vcl.list -j" and find active vcls
func getActiveVCLJson(out io.Reader) (string, error) {
	var output = ""
	if b, err := io.ReadAll(out); err == nil {
		output = string(b)
	}
	// workaround for non valid json in varnish 6.6.1 https://github.com/varnishcache/varnish-cache/issues/3687
	output = strings.TrimPrefix(output, "200")

	var jsonOut []interface{}
	err := json.Unmarshal([]byte(output), &jsonOut)
	if err != nil {
		return "", err
	}

	for _, item := range jsonOut {
		switch s := item.(type) {
		case []interface{}:
			command := s[0]
			if command != "vcl.list" {
				return "", fmt.Errorf("unsupported varnishadm command %v", jsonOut[1])
			}
		case map[string]interface{}:
			if s["status"] == "active" {
				return s["name"].(string), nil
			}
		default:
			//ignore
			continue
		}
	}
	return "", nil
}

// Gets the "counters" section from varnishstat json (there is change in schema structure in varnish 6.5+)
func getCountersJSON(rootJSON map[string]interface{}) map[string]interface{} {
	//version 1 contains "counters" wrapper
	if counters, exists := rootJSON["counters"]; exists {
		return counters.(map[string]interface{})
	}
	return rootJSON
}

// converts varnish metrics name into field and list of tags
func (s *Varnish) parseMetricV2(name string) (metric varnishMetric) {
	metric.measurement = measurementNamespace
	if strings.Count(name, ".") == 0 {
		return metric
	}
	metric.fieldName = name[strings.LastIndex(name, ".")+1:]
	var section = strings.Split(name, ".")[0]
	metric.tags = map[string]string{
		"section": section,
	}

	//parse name using regexpsCompiled
	for _, re := range s.regexpsCompiled {
		submatch := re.FindStringSubmatch(name)
		if len(submatch) < 1 {
			continue
		}
		for _, sub := range re.SubexpNames() {
			if sub == "" {
				continue
			}
			val := submatch[re.SubexpIndex(sub)]
			if sub == "_vcl" {
				metric.vclName = val
			} else if sub == "_field" {
				metric.fieldName = val
			} else if val != "" {
				metric.tags[sub] = val
			}
		}
		break
	}
	return metric
}

type varnishMetric struct {
	measurement string
	fieldName   string
	tags        map[string]string
	vclName     string
}

func init() {
	inputs.Add("varnish", func() telegraf.Input {
		return &Varnish{
			run:             varnishRunner,
			admRun:          varnishRunner,
			regexpsCompiled: defaultRegexps,
			Stats:           defaultStats,
			Binary:          defaultStatBinary,
			AdmBinary:       defaultAdmBinary,
			MetricVersion:   1,
			UseSudo:         false,
			InstanceName:    "",
			Timeout:         defaultTimeout,
			Regexps:         []string{},
		}
	})
}
