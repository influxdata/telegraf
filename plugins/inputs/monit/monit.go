package monit

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/net/html/charset"
)

var pendingActions = []string{"ignore", "alert", "restart", "stop", "exec", "unmonitor", "start", "monitor"}

type Status struct {
	Server   Server    `xml:"server"`
	Platform Platform  `xml:"platform"`
	Services []Service `xml:"service"`
}

type Server struct {
	ID      string `xml:"id"`
	Version string `xml:"version"`
	Uptime  int    `xml:"uptime"`
	Poll    int    `xml:"poll"`
}

type Platform struct {
	Name    string `xml:"name"`
	Release string `xml:"release"`
	Version string `xml:"version"`
	Machine string `xml:"machine"`
	CPU     int    `xml:"cpu"`
	Memory  int    `xml:"memory"`
	Swap    int    `xml:"swap"`
}

type Service struct {
	Type             string `xml:"type,attr"`
	Name             string `xml:"name"`
	Status           int64  `xml:"status"`
	MonitoringStatus int64  `xml:"monitor"`
	PendingAction    int    `xml:"pendingaction"`
	Uptime           int64  `xml:"uptime"`
	Memory           Memory `xml:"memory"`
	CPU              CPU    `xml:"cpu"`
	System           System `xml:"system"`
}

type Memory struct {
	Percent       float64 `xml:"percent"`
	PercentTotal  float64 `xml:"percenttotal"`
	Kilobyte      int64   `xml:"kilobyte"`
	KilobyteTotal int64   `xml:"kilobytetotal"`
}

type CPU struct {
	Percent      float64 `xml:"percent"`
	PercentTotal float64 `xml:"percenttotal"`
}

type System struct {
	Load struct {
		Avg01 float64 `xml:"avg01"`
		Avg05 float64 `xml:"avg05"`
		Avg15 float64 `xml:"avg15"`
	} `xml:"load"`
	CPU struct {
		User   float64 `xml:"user"`
		System float64 `xml:"system"`
		Wait   float64 `xml:"wait"`
	} `xml:"cpu"`
	Memory struct {
		Percent  float64 `xml:"percent"`
		Kilobyte int64   `xml:"kilobyte"`
	} `xml:"memory"`
	Swap struct {
		Percent  float64 `xml:"percent"`
		Kilobyte float64 `xml:"kilobyte"`
	} `xml:"swap"`
}

type Monit struct {
	Address           string
	BasicAuthUsername string
	BasicAuthPassword string
}

func (m *Monit) Description() string {
	return "Read metrics and status information about processes managed by Monit"
}

var sampleConfig = `
  ## Monit
  address = "http://127.0.0.1:2812"
  basic_auth_username = ""
  basic_auth_password = ""
`

func (m *Monit) SampleConfig() string {
	return sampleConfig
}

func (m *Monit) Gather(acc telegraf.Accumulator) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_status?format=xml", m.Address), nil)
	if err != nil {
		return err
	}
	if len(m.BasicAuthUsername) > 0 || len(m.BasicAuthPassword) > 0 {
		req.SetBasicAuth(m.BasicAuthUsername, m.BasicAuthPassword)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var status Status
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&status); err != nil {
		return fmt.Errorf("Cannot parse input with error: %v\n", err)
	}

	// Prepare tags
	tags := map[string]string{"address": m.Address, "version": status.Server.Version}

	for _, service := range status.Services {
		// Prepare fields
		fields := make(map[string]interface{})
		fields["status_code"] = service.Status
		fields["status"] = serviceStatus(service)
		fields["monitoring_status_code"] = service.MonitoringStatus
		fields["monitoring_status"] = monitoringStatus(service)

		if service.Type == "3" {
			fields["cpu_percent"] = service.CPU.Percent
			fields["cpu_percent_total"] = service.CPU.PercentTotal
			fields["mem_kb"] = service.Memory.Kilobyte
			fields["mem_kb_total"] = service.Memory.KilobyteTotal
			fields["mem_percent"] = service.Memory.Percent
			fields["mem_percent_total"] = service.Memory.PercentTotal
			fields["service_uptime"] = service.Uptime
		} else if service.Type == "5" {
			fields["cpu_system"] = service.System.CPU.System
			fields["cpu_user"] = service.System.CPU.User
			fields["cpu_wait"] = service.System.CPU.Wait
			fields["cpu_load_avg_1m"] = service.System.Load.Avg01
			fields["cpu_load_avg_5m"] = service.System.Load.Avg05
			fields["cpu_load_avg_15m"] = service.System.Load.Avg15
			fields["mem_kb"] = service.System.Memory.Kilobyte
			fields["mem_percent"] = service.System.Memory.Percent
			fields["swap_kb"] = service.System.Swap.Kilobyte
			fields["swap_percent"] = service.System.Swap.Percent
		}

		tags["service"] = service.Name
		tags["service_type"] = service.Type
		acc.AddFields("monit", fields, tags)
	}

	return nil
}

func serviceStatus(s Service) string {
	var status string

	if s.MonitoringStatus == 0 || s.MonitoringStatus == 2 {
		status = monitoringStatus(s)
	} else if s.Status == 0 {
		status = "Running"
	} else {
		status = "Failure"
	}

	if s.PendingAction > 0 {
		if s.PendingAction >= len(pendingActions) {
			return fmt.Sprintf("%s - pending", status)
		}
		status = fmt.Sprintf("%s - %s pending", status, pendingActions[s.PendingAction])
	}
	return status
}

func monitoringStatus(s Service) string {
	switch s.MonitoringStatus {
	case 1:
		return "Running"
	case 2:
		return "Initializing"
	case 4:
		return "Waiting"
	}
	return "Not monitored"
}

func init() {
	inputs.Add("monit", func() telegraf.Input {
		return &Monit{}
	})
}
