package monit

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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
	Status           int    `xml:"status"`
	MonitoringStatus int    `xml:"monitor"`
	PendingAction    int    `xml:"pendingaction"`
	Uptime           int    `xml:"uptime"`
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
		Kilobyte float64 `xml:"kilobyte"`
	} `xml:"memory"`
	Swap struct {
		Percent  float64 `xml:"percent"`
		Kilobyte float64 `xml:"kilobyte"`
	} `xml:"swap"`
}

type Monit struct {
	Address string
}

func (m *Monit) Description() string {
	return "Read metrics and status information about processes managed by Monit"
}

var sampleConfig = `
  ## Monit
  url = "http://127.0.0.1:2812"
  basic_auth_username = ""
  basic_auth_password = ""
`

func (m *Monit) SampleConfig() string {
	return sampleConfig
}

func (m *Monit) Gather(acc telegraf.Accumulator) error {
	return nil
}

func serviceStatus(s Service) string {
	var status string

	if s.MonitoringStatus == 0 || s.MonitoringStatus == 2 {
		status = monitoringStatus(s)
	} else if s.Status == 1 {
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
