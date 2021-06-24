// +build !windows

package varnish_cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

const (
	reloadPrefix         = "VBE.reload_"
	measurementNamespace = "varnish"
)

var (
	defaultBinary  = "/usr/bin/varnishstat"
	defaultTimeout = config.Duration(time.Second)
	//// REGEXPS for parsing backend and server from
	// (prefix:)<uuid>.<name>
	regexBackendUUID = regexp.MustCompile(`([[0-9A-Za-z]{8}-[0-9A-Za-z]{4}-[0-9A-Za-z]{4}-[89ABab][0-9A-Za-z]{3}-[0-9A-Za-z]{12})(.*)`)
	// <name>(<ip>,(<something>),<port>)
	regexBackendParen = regexp.MustCompile(`(.*)\((.*)\)`)
)

func extractMeasurement(vName string) string {
	vNameLower := strings.ToLower(vName)
	return measurementNamespace + "_" + strings.Split(vNameLower, ".")[0]
}

func trimMeasurementPrefix(name string) string {
	index := strings.Index(name, ".")
	if index > 0 {
		return name[index+1:]
	}
	return name
}

type runner func(cmdName string, args []string, useSudo bool, instanceName string, timeout config.Duration) (*bytes.Buffer, error)

// VarnishCache is used to store configuration values
type VarnishCache struct {
	Binary       string
	Args         []string
	UseSudo      bool
	InstanceName string
	Timeout      config.Duration
	run          runner
	Log          telegraf.Logger
}

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"

  ## Optional command line arguments	
  # args = ["-j"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually append after -n in varnish cli
  # instance_name = instanceName

  ## Timeout for varnishstat command
  # timeout = "1s"
`

func (s *VarnishCache) Description() string {
	return "A plugin to collect stats from Varnish HTTP Cache"
}

// SampleConfig displays configuration instructions
func (s *VarnishCache) SampleConfig() string {
	return sampleConfig
}

// Shell out to varnish_stat and return the output
func varnishRunner(
	cmdName string,
	args []string,
	useSudo bool,
	instanceName string,
	timeout config.Duration,
) (*bytes.Buffer, error) {
	var cmdArgs []string

	//custom args override
	if args != nil {
		cmdArgs = append(args, cmdArgs...)
	} else {
		// json output
		cmdArgs = []string{"-j"}
	}

	if instanceName != "" {
		cmdArgs = append(cmdArgs, []string{"-n", instanceName}...)
	}

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
		return &out, fmt.Errorf("error running varnishstat: %s", err)
	}

	return &out, nil
}

// Gather collects the configured stats from varnish_stat and adds them to the
// Accumulator
// The prefix of each stat (eg MAIN, MEMPOOL, LCK, etc) will be used as a
// measurement name, string after last "." parsed as a field, middle part is parsed into tags.
func (s *VarnishCache) Gather(acc telegraf.Accumulator) error {

	out, err := s.run(s.Binary, s.Args, s.UseSudo, s.InstanceName, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}
	rootJSON := make(map[string]interface{})
	dec := json.NewDecoder(out)
	dec.UseNumber()
	if err := dec.Decode(&rootJSON); err != nil {
		return err
	}
	return s.processJSON(acc, rootJSON)
}

func getCountersJSON(rootJSON map[string]interface{}) (map[string]interface{}, error) {

	countersJSON := make(map[string]interface{})

	// schema version and counters structure added in 6.5+
	version := rootJSON["version"]
	if version != nil {
		versionNum, ok := version.(json.Number)
		if !ok {
			return nil, fmt.Errorf("invalid json schema version")
		}
		switch versionNum.String() {
		case "1":
			countersJSON = rootJSON["counters"].(map[string]interface{})
		default:
			return nil, fmt.Errorf("unsupported json stats version: %s", versionNum)
		}
	} else {
		countersJSON = rootJSON
	}
	return countersJSON, nil

}

// Adds varnish stats json into into accumulator
func (s *VarnishCache) processJSON(acc telegraf.Accumulator, rootJSON map[string]interface{}) error {

	countersJSON, err := getCountersJSON(rootJSON)
	if err != nil {
		acc.AddError(err)
	}

	timestamp := time.Now()
	//find the most recent VBE.reload_
	recentVbeReloadPrefix := findActiveReloadPrefix(countersJSON)

	for vFieldName, raw := range countersJSON {
		if vFieldName == "timestamp" {
			continue
		}

		//skip old reload_ prefixes
		if isReloadPrefixNotActive(vFieldName, recentVbeReloadPrefix) {
			continue
		}

		data, ok := raw.(map[string]interface{})

		if !ok {
			acc.AddError(fmt.Errorf("W: unexpected data from json: %s: %#v\n", vFieldName, raw))
			continue
		}
		var (
			//measurement = extractMeasurement(vFieldName)
			vValue float64
			//iValue      uint64
			vErr error
		)

		flag, _ := data["flag"]

		// parse value
		if value, ok := data["value"]; ok {
			if number, ok := value.(json.Number); ok {
				if vValue, vErr = number.Float64(); vErr != nil {
					vErr = fmt.Errorf("%s value float64 error: %s", vFieldName, vErr)
				}
				//Happy health probes TODO
				if flag == "b" {
					if _, vErr = strconv.ParseUint(number.String(), 10, 64); vErr != nil {
						vErr = fmt.Errorf("%s value uint64 error: %s", vFieldName, vErr)
					}
				}
			} else {
				vErr = fmt.Errorf("%s value it not a float64", vFieldName)
			}
		}

		if vErr != nil {
			acc.AddError(vErr)
			continue
		}

		measurement, field, tags := createMetric(vFieldName)
		fields := make(map[string]interface{})
		fields[field] = vValue
		switch flag {
		case "c", "a":
			acc.AddCounter(measurement, fields, tags, timestamp)
		case "g":
			acc.AddGauge(measurement, fields, tags, timestamp)
		default:
			acc.AddGauge(measurement, fields, tags, timestamp)
		}
	}
	return nil
}

// converts varnish metrics name into field and list of tags
func createMetric(vName string) (measurement string, name string, tags map[string]string) {
	measurement = extractMeasurement(vName)
	var innerPart string
	if strings.Count(vName, ".") > 1 {
		innerPart = trimMeasurementPrefix(strings.ToLower(vName))
		innerPart = innerPart[0:strings.LastIndex(innerPart, ".")]
	}
	name = trimMeasurementPrefix(strings.ToLower(vName))
	if len(innerPart) > 0 {
		name = strings.Replace(name, strings.ToLower(innerPart)+".", "", -1)
	}
	name = strings.Replace(name, ".", "_", -1)
	tags = make(map[string]string)
	if len(name) > 0 {
		if isVBE := strings.HasPrefix(vName, "VBE."); isVBE {
			if hits := regexBackendUUID.FindAllStringSubmatch(innerPart, -1); len(hits) > 0 && len(hits[0]) >= 3 {
				tags["backend"] = cleanBackendName(hits[0][2])
				tags["server"] = hits[0][1]
			} else if hits := regexBackendParen.FindAllStringSubmatch(innerPart, -1); len(hits) > 0 && len(hits[0]) >= 3 {
				tags["backend"] = cleanBackendName(hits[0][1])
				tags["server"] = strings.Replace(hits[0][2], ",,", ":", 1)
			}
			if len(tags) == 0 {
				tags["backend"] = cleanBackendName(innerPart)
			}
		}
		if len(tags) == 0 && innerPart != "" {
			tags["id"] = innerPart
		}
	}
	return measurement, name, tags
}

func cleanBackendName(name string) string {
	name = strings.Trim(name, ".")
	for _, prefix := range []string{"boot.", "root:"} {
		if strings.HasPrefix(name, prefix) {
			name = name[len(prefix):]
		}
	}
	// reload_20191014_091124_78599.<name>
	if strings.HasPrefix(name, "reload_") {
		dot := strings.Index(name, ".")
		if dot != -1 {
			name = name[dot+1:]
		}
	}
	return name
}

// Find the most recent 'VBE.reload_' prefix using string compare
// 'VBE.reload_20210623_170621_31083'
func findActiveReloadPrefix(json map[string]interface{}) string {
	var prefix string
	for vName := range json {
		if strings.HasPrefix(vName, reloadPrefix) && strings.HasSuffix(vName, ".happy") {
			dotAfterPrefixIndex := len(reloadPrefix) + strings.Index(vName[len(reloadPrefix):], ".")
			vbeReloadPrefix := vName[:dotAfterPrefixIndex]
			if strings.Compare(vbeReloadPrefix, prefix) > 0 {
				prefix = vbeReloadPrefix
			}
		}
	}
	return prefix
}

// Returns true if the given 'VBE.' metric contains outdated reload_ prefix
func isReloadPrefixNotActive(vName string, activeReloadPrefix string) bool {
	return activeReloadPrefix != "" && strings.HasPrefix(vName, "VBE.") &&
		!strings.HasPrefix(vName, activeReloadPrefix)
}

func init() {
	inputs.Add("varnish_cache", func() telegraf.Input {
		return &VarnishCache{
			run:          varnishRunner,
			Binary:       defaultBinary,
			Args:         nil,
			UseSudo:      false,
			InstanceName: "",
			Timeout:      defaultTimeout,
		}
	})
}
