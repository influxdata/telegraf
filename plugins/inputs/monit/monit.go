// Copyright 2020, Verizon
// Licensed under the terms of the MIT License. See LICENSE file in project root for terms.

package monit

import (
	"encoding/xml"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/net/html/charset"
	"net/http"
)

var pendingActions = []string{"ignore", "alert", "restart", "stop", "exec", "unmonitor", "start", "monitor"}

type Status struct {
	Server   Server    `xml:"server"`
	Platform Platform  `xml:"platform"`
	Services []Service `xml:"service"`
}

type Server struct {
	ID            string `xml:"id"`
	Version       string `xml:"version"`
	Uptime        int    `xml:"uptime"`
	Poll          int    `xml:"poll"`
	LocalHostname string `xml:"localhostname"`
	StartDelay    int    `xml:"startdelay"`
	ControlFile   string `xml:"controlfile"`
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
	Type             string  `xml:"type,attr"`
	Name             string  `xml:"name"`
	Status           int     `xml:"status"`
	MonitoringStatus int     `xml:"monitor"`
	MonitorMode      int     `xml:"monitormode"`
	PendingAction    int     `xml:"pendingaction"`
	Uptime           int64   `xml:"uptime"`
	Memory           Memory  `xml:"memory"`
	CPU              CPU     `xml:"cpu"`
	System           System  `xml:"system"`
	Size             int64   `xml:"size"`
	Mode             int     `xml:"mode"`
	Program          Program `xml:"program"`
	Block            Block   `xml:"block"`
	Inode            Inode   `xml:"inode"`
	Pid              int64   `xml:"pid"`
	ParentPid        int64   `xml"ppid"`
	Threads          int     `xml:"threads"`
	Children         int     `xml:"children"`
	Port             Port    `xml:"port"`
	Link             Link    `xml:"link"`
}

type Link struct {
	State    int      `xml:"state"`
	Speed    int64    `xml:"speed"`
	Duplex   int      `xml:"duplex"`
	Download Download `xml:"download"`
	Upload   Upload   `xml:"upload"`
}

type Download struct {
	Packets struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"packets"`
	Bytes struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"bytes"`
	Errors struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"errors"`
}

type Upload struct {
	Packets struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"packets"`
	Bytes struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"bytes"`
	Errors struct {
		Now   int64 `xml:"now"`
		Total int64 `xml:"total"`
	} `xml:"errors"`
}

type Port struct {
	Hostname     string  `xml:"hostname"`
	PortNumber   int64   `xml:"portnumber"`
	Request      string  `xml:"request"`
	Protocol     string  `xml:"protocol"`
	Type         string  `xml:"type"`
	ResponseTime float64 `xml:"responsetime"`
}

type Block struct {
	Percent float64 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

type Inode struct {
	Percent float64 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

type Program struct {
	Started int64  `xml:"started"`
	Status  int    `xml:"status"`
	Output  string `xml:"output"`
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
	parser            parsers.Parser
	client            *http.Client
	tls.ClientConfig
	Timeout internal.Duration `toml:"timeout"`
}

func (m *Monit) Description() string {
	return "Read metrics and status information about processes managed by Monit"
}

var sampleConfig = `
  ## Monit
  address = "http://127.0.0.1:2812"

  ## Username and Password for Monit
  basic_auth_username = ""
  basic_auth_password = ""

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (m *Monit) SampleConfig() string {
	return sampleConfig
}

func (m *Monit) SetParser(parser parsers.Parser) {
	m.parser = parser
}

func (m *Monit) Init() error {
	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	m.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: m.Timeout.Duration,
	}
	return nil
}

func (m *Monit) Gather(acc telegraf.Accumulator) error {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_status?format=xml", m.Address), nil)
	if err != nil {
		return err
	}
	if len(m.BasicAuthUsername) > 0 || len(m.BasicAuthPassword) > 0 {
		req.SetBasicAuth(m.BasicAuthUsername, m.BasicAuthPassword)
	}

	resp, err := m.client.Do(req)
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

	tags := map[string]string{"address": m.Address, "version": status.Server.Version, "hostname": status.Server.LocalHostname, "platform_name": status.Platform.Name}

	for _, service := range status.Services {
		fields := make(map[string]interface{})
		fields["status_code"] = service.Status
		fields["status"] = serviceStatus(service)
		fields["monitoring_status_code"] = service.MonitoringStatus
		fields["monitoring_status"] = monitoringStatus(service)
		fields["monitoring_mode_status"] = monitoringMode(service)
		fields["monitoring_mode_code"] = service.MonitorMode
		tags["service"] = service.Name
		tags["service_type"] = service.Type
		if service.Type == "0" {
			fields["mode"] = service.Mode
			fields["block_percent"] = service.Block.Percent
			fields["block_usage"] = service.Block.Usage
			fields["block_total"] = service.Block.Total
			fields["inode_percent"] = service.Inode.Percent
			fields["inode_usage"] = service.Inode.Usage
			fields["inode_total"] = service.Inode.Total
			acc.AddFields("filesystem", fields, tags)
		} else if service.Type == "1" {
			fields["permissions"] = service.Mode
			acc.AddFields("directory", fields, tags)
		} else if service.Type == "2" {
			fields["size"] = service.Size
			fields["permissions"] = service.Mode
			acc.AddFields("file", fields, tags)
		} else if service.Type == "3" {
			fields["cpu_percent"] = service.CPU.Percent
			fields["cpu_percent_total"] = service.CPU.PercentTotal
			fields["mem_kb"] = service.Memory.Kilobyte
			fields["mem_kb_total"] = service.Memory.KilobyteTotal
			fields["mem_percent"] = service.Memory.Percent
			fields["mem_percent_total"] = service.Memory.PercentTotal
			fields["service_uptime"] = service.Uptime
			fields["pid"] = service.Pid
			fields["parent_pid"] = service.ParentPid
			fields["threads"] = service.Threads
			fields["children"] = service.Children
			acc.AddFields("process", fields, tags)
		} else if service.Type == "4" {
			fields["hostname"] = service.Port.Hostname
			fields["port_number"] = service.Port.PortNumber
			fields["request"] = service.Port.Request
			fields["response_time"] = service.Port.ResponseTime
			fields["protocol"] = service.Port.Protocol
			fields["type"] = service.Port.Type
			acc.AddFields("remote_host", fields, tags)
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
			acc.AddFields("system", fields, tags)
		} else if service.Type == "6" {
			fields["permissions"] = service.Mode
			acc.AddFields("fifo", fields, tags)
		} else if service.Type == "7" {
			fields["last_started_time"] = service.Program.Started * 1000
			fields["program_status"] = service.Program.Status
			fields["output"] = service.Program.Output
			acc.AddFields("program", fields, tags)
		} else if service.Type == "8" {
			fields["link_state"] = service.Link.State
			fields["link_speed"] = service.Link.Speed
			fields["link_mode"] = linkMode(service)
			fields["download_packets_now"] = service.Link.Download.Packets.Now
			fields["download_packets_total"] = service.Link.Download.Packets.Total
			fields["download_bytes_now"] = service.Link.Download.Bytes.Now
			fields["download_bytes_total"] = service.Link.Download.Bytes.Total
			fields["download_errors_now"] = service.Link.Download.Errors.Now
			fields["download_errors_total"] = service.Link.Download.Errors.Total
			fields["upload_packets_now"] = service.Link.Upload.Packets.Now
			fields["upload_packets_total"] = service.Link.Upload.Packets.Total
			fields["upload_bytes_now"] = service.Link.Upload.Bytes.Now
			fields["upload_bytes_total"] = service.Link.Upload.Bytes.Total
			fields["upload_errors_now"] = service.Link.Upload.Errors.Now
			fields["upload_errors_total"] = service.Link.Upload.Errors.Total
			acc.AddFields("network", fields, tags)
		}
	}
	return nil
}

func linkMode(s Service) string {
	if s.Link.Duplex == 1 {
		return "Duplex Mode"
	} else {
		return "Simplex Mode"
	}
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

func monitoringMode(s Service) string {
	switch s.MonitorMode {
	case 0:
		return "Monitoring in Active mode"
	case 1:
		return "Monitoring in Passive mode"
	}
	return "Unknown Mode"
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
