// Package neptuneapex implements an input plugin for the Neptune Apex
// aquarium controller.
package neptuneapex

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Measurement is constant across all metrics.
const Measurement = "neptune_apex"

type xmlReply struct {
	SoftwareVersion string   `xml:"software,attr"`
	HardwareVersion string   `xml:"hardware,attr"`
	Hostname        string   `xml:"hostname"`
	Serial          string   `xml:"serial"`
	Timezone        float64  `xml:"timezone"`
	Date            string   `xml:"date"`
	PowerFailed     string   `xml:"power>failed"`
	PowerRestored   string   `xml:"power>restored"`
	Probe           []probe  `xml:"probes>probe"`
	Outlet          []outlet `xml:"outlets>outlet"`
}

type probe struct {
	Name  string  `xml:"name"`
	Value string  `xml:"value"`
	Type  *string `xml:"type"`
}

type outlet struct {
	Name     string  `xml:"name"`
	OutputID string  `xml:"outputID"`
	State    string  `xml:"state"`
	DeviceID string  `xml:"deviceID"`
	Xstatus  *string `xml:"xstatus"`
}

// NeptuneApex implements telegraf.Input.
type NeptuneApex struct {
	Servers         []string
	ResponseTimeout internal.Duration
	httpClient      *http.Client
}

// Description implements telegraf.Input.Description
func (*NeptuneApex) Description() string {
	return "Neptune Apex data collector"
}

// SampleConfig implements telegraf.Input.SampleConfig
func (*NeptuneApex) SampleConfig() string {
	return `
  ## The Neptune Apex plugin reads the publicly available status.xml data from a local Apex.
  ## Measurements will be logged under "apex".

  ## The base URL of the local Apex(es). If you specify more than one server, they will
  ## be differentiated by the "source" tag.
  servers = [
    "http://apex.local",
  ]

  ## The response_timeout specifies how long to wait for a reply from the Apex.
  #response_timeout = "5s"
`
}

// Gather implements telegraf.Input.Gather
func (n *NeptuneApex) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, server := range n.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(n.gatherServer(acc, server))
		}(server)
	}
	wg.Wait()
	return nil
}

func (n *NeptuneApex) gatherServer(
	acc telegraf.Accumulator, server string) error {
	resp, err := n.sendRequest(server)
	if err != nil {
		return err
	}
	return n.parseXML(acc, resp)
}

// parseXML is strict on the input and does not do best-effort parsing.
//This is because of the life-support nature of the Neptune Apex.
func (n *NeptuneApex) parseXML(acc telegraf.Accumulator, data []byte) error {
	r := xmlReply{}
	err := xml.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("unable to unmarshal XML: %v\nXML DATA: %q",
			err, data)
	}

	var reportTime time.Time
	var powerFailed, powerRestored int64
	if reportTime, err = parseTime(r.Date, r.Timezone); err != nil {
		return err
	}
	if val, err := parseTime(r.PowerFailed, r.Timezone); err != nil {
		return err
	} else {
		powerFailed = val.UnixNano()
	}
	if val, err := parseTime(r.PowerRestored, r.Timezone); err != nil {
		return err
	} else {
		powerRestored = val.UnixNano()
	}

	mainFields := map[string]interface{}{
		"serial":         r.Serial,
		"power_failed":   powerFailed,
		"power_restored": powerRestored,
	}
	acc.AddFields(Measurement, mainFields,
		map[string]string{
			"source":   r.Hostname,
			"type":     "controller",
			"software": r.SoftwareVersion,
			"hardware": r.HardwareVersion,
		},
		reportTime)

	// Outlets.
	for _, o := range r.Outlet {
		tags := map[string]string{
			"source":    r.Hostname,
			"output_id": o.OutputID,
			"device_id": o.DeviceID,
			"name":      o.Name,
			"type":      "output",
			"software":  r.SoftwareVersion,
			"hardware":  r.HardwareVersion,
		}
		fields := map[string]interface{}{
			"state": o.State,
		}
		// Find Amp and Watt probes and add them as fields.
		// Remove the redundant probe.
		if pos := findProbe(fmt.Sprintf("%sW", o.Name), r.Probe); pos > -1 {
			value, err := strconv.ParseFloat(
				strings.TrimSpace(r.Probe[pos].Value), 64)
			if err != nil {
				acc.AddError(
					fmt.Errorf(
						"cannot convert string value %q to float64: %v",
						r.Probe[pos].Value, err))
				continue // Skip the whole outlet.
			}
			fields["watt"] = value
			r.Probe[pos] = r.Probe[len(r.Probe)-1]
			r.Probe = r.Probe[:len(r.Probe)-1]
		}
		if pos := findProbe(fmt.Sprintf("%sA", o.Name), r.Probe); pos > -1 {
			value, err := strconv.ParseFloat(
				strings.TrimSpace(r.Probe[pos].Value), 64)
			if err != nil {
				acc.AddError(
					fmt.Errorf(
						"cannot convert string value %q to float64: %v",
						r.Probe[pos].Value, err))
				break // // Skip the whole outlet.
			}
			fields["amp"] = value
			r.Probe[pos] = r.Probe[len(r.Probe)-1]
			r.Probe = r.Probe[:len(r.Probe)-1]
		}
		if o.Xstatus != nil {
			fields["xstatus"] = *o.Xstatus
		}
		// Try to determine outlet type. Focus on accuracy, leaving the
		//outlet_type "unknown" when ambiguous. 24v and vortech cannot be
		// determined.
		switch {
		case strings.HasPrefix(o.DeviceID, "base_Var"):
			tags["output_type"] = "variable"
		case o.DeviceID == "base_Alarm":
			fallthrough
		case o.DeviceID == "base_Warn":
			fallthrough
		case strings.HasPrefix(o.DeviceID, "base_email"):
			tags["output_type"] = "alert"
		case fields["watt"] != nil || fields["amp"] != nil:
			tags["output_type"] = "outlet"
		case strings.HasPrefix(o.DeviceID, "Cntl_"):
			tags["output_type"] = "virtual"
		default:
			tags["output_type"] = "unknown"
		}

		acc.AddFields(Measurement, fields, tags, reportTime)
	}

	// Probes.
	for _, p := range r.Probe {
		value, err := strconv.ParseFloat(strings.TrimSpace(p.Value), 64)
		if err != nil {
			acc.AddError(fmt.Errorf(
				"cannot convert string value %q to float64: %v",
				p.Value, err))
			continue
		}
		fields := map[string]interface{}{
			"value": value,
		}
		tags := map[string]string{
			"source":   r.Hostname,
			"type":     "probe",
			"name":     p.Name,
			"software": r.SoftwareVersion,
			"hardware": r.HardwareVersion,
		}
		if p.Type != nil {
			tags["probe_type"] = *p.Type
		}
		acc.AddFields(Measurement, fields, tags, reportTime)
	}

	return nil
}

func findProbe(probe string, probes []probe) int {
	for i, p := range probes {
		if p.Name == probe {
			return i
		}
	}
	return -1
}

// parseTime takes a Neptune Apex date/time string with a timezone and
// returns a time.Time struct.
func parseTime(val string, tz float64) (time.Time, error) {
	// Magic time constant from https://golang.org/pkg/time/#Parse
	const TimeLayout = "01/02/2006 15:04:05 -0700"

	// Timezone offset needs to be explicit
	sign := '+'
	if tz < 0 {
		sign = '-'
	}

	// Build a time string with the timezone in a format Go can parse.
	tzs := fmt.Sprintf("%c%04d", sign, int(math.Abs(tz))*100)
	ts := fmt.Sprintf("%s %s", val, tzs)
	t, err := time.Parse(TimeLayout, ts)
	if err != nil {
		return time.Now(), fmt.Errorf("unable to parse %q (%v)", ts, err)
	}
	return t, nil
}

func (n *NeptuneApex) sendRequest(server string) ([]byte, error) {
	url := fmt.Sprintf("%s/cgi-bin/status.xml", server)
	resp, err := n.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"response from server URL %q returned %d (%s), expected %d (%s)",
			url, resp.StatusCode, http.StatusText(resp.StatusCode),
			http.StatusOK, http.StatusText(http.StatusOK))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read output from %q: %v", url, err)
	}
	return body, nil
}

func init() {
	inputs.Add("neptune_apex", func() telegraf.Input {
		return &NeptuneApex{
			httpClient: &http.Client{
				Timeout: 5 * time.Second,
			},
		}
	})
}
