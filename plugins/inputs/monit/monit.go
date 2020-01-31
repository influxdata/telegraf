package monit

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/net/html/charset"
)

const (
	fileSystem string = "0"
	directory         = "1"
	file              = "2"
	process           = "3"
	remoteHost        = "4"
	system            = "5"
	fifo              = "6"
	program           = "7"
	network           = "8"
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
	Uptime        int64  `xml:"uptime"`
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
	Memory           Memory  `xml:"memory"`
	CPU              CPU     `xml:"cpu"`
	System           System  `xml:"system"`
	Size             int64   `xml:"size"`
	Mode             int     `xml:"mode"`
	Program          Program `xml:"program"`
	Block            Block   `xml:"block"`
	Inode            Inode   `xml:"inode"`
	Pid              int64   `xml:"pid"`
	ParentPid        int64   `xml:"ppid"`
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
	Hostname   string `xml:"hostname"`
	PortNumber int64  `xml:"portnumber"`
	Request    string `xml:"request"`
	Protocol   string `xml:"protocol"`
	Type       string `xml:"type"`
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
	Started int64 `xml:"started"`
	Status  int   `xml:"status"`
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
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	client   http.Client
	tls.ClientConfig
	Timeout internal.Duration `toml:"timeout"`
}

type Messagebody struct {
	Metrics []string `json:"metrics"`
}

func (m *Monit) Description() string {
	return "Read metrics and status information about processes managed by Monit"
}

var sampleConfig = `
  ## Monit HTTPD address
  address = "http://127.0.0.1:2812"

  ## Username and Password for Monit
  # username = ""
  # password = ""

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

func (m *Monit) Init() error {
	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	m.client = http.Client{
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
	if len(m.Username) > 0 || len(m.Password) > 0 {
		req.SetBasicAuth(m.Username, m.Password)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		var status Status
		decoder := xml.NewDecoder(resp.Body)
		decoder.CharsetReader = charset.NewReaderLabel
		if err := decoder.Decode(&status); err != nil {
			return fmt.Errorf("error parsing input: %v", err)
		}

		tags := map[string]string{
			"version":       status.Server.Version,
			"source":        status.Server.LocalHostname,
			"platform_name": status.Platform.Name,
		}

		for _, service := range status.Services {
			fields := make(map[string]interface{})
			tags["status"] = serviceStatus(service)
			fields["status_code"] = service.Status
			tags["pending_action"] = pendingAction(service)
			fields["pending_action_code"] = service.PendingAction
			tags["monitoring_status"] = monitoringStatus(service)
			fields["monitoring_status_code"] = service.MonitoringStatus
			tags["monitoring_mode"] = monitoringMode(service)
			fields["monitoring_mode_code"] = service.MonitorMode
			tags["service"] = service.Name
			if service.Type == fileSystem {
				fields["mode"] = service.Mode
				fields["block_percent"] = service.Block.Percent
				fields["block_usage"] = service.Block.Usage
				fields["block_total"] = service.Block.Total
				fields["inode_percent"] = service.Inode.Percent
				fields["inode_usage"] = service.Inode.Usage
				fields["inode_total"] = service.Inode.Total
				acc.AddFields("monit_filesystem", fields, tags)
			} else if service.Type == directory {
				fields["mode"] = service.Mode
				acc.AddFields("monit_directory", fields, tags)
			} else if service.Type == file {
				fields["size"] = service.Size
				fields["mode"] = service.Mode
				acc.AddFields("monit_file", fields, tags)
			} else if service.Type == process {
				fields["cpu_percent"] = service.CPU.Percent
				fields["cpu_percent_total"] = service.CPU.PercentTotal
				fields["mem_kb"] = service.Memory.Kilobyte
				fields["mem_kb_total"] = service.Memory.KilobyteTotal
				fields["mem_percent"] = service.Memory.Percent
				fields["mem_percent_total"] = service.Memory.PercentTotal
				fields["pid"] = service.Pid
				fields["parent_pid"] = service.ParentPid
				fields["threads"] = service.Threads
				fields["children"] = service.Children
				acc.AddFields("monit_process", fields, tags)
			} else if service.Type == remoteHost {
				fields["remote_hostname"] = service.Port.Hostname
				fields["port_number"] = service.Port.PortNumber
				fields["request"] = service.Port.Request
				fields["protocol"] = service.Port.Protocol
				fields["type"] = service.Port.Type
				acc.AddFields("monit_remote_host", fields, tags)
			} else if service.Type == system {
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
				acc.AddFields("monit_system", fields, tags)
			} else if service.Type == fifo {
				fields["mode"] = service.Mode
				acc.AddFields("monit_fifo", fields, tags)
			} else if service.Type == program {
				fields["program_started"] = service.Program.Started * 10000000
				fields["program_status"] = service.Program.Status
				acc.AddFields("monit_program", fields, tags)
			} else if service.Type == network {
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
				acc.AddFields("monit_network", fields, tags)
			}
		}
	} else {
		return fmt.Errorf("received status code %d (%s), expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))

	}
	return nil
}

func linkMode(s Service) string {
	if s.Link.Duplex == 1 {
		return "duplex"
	} else if s.Link.Duplex == 0 {
		return "simplex"
	} else {
		return "unknown"
	}
}

func serviceStatus(s Service) string {
	if s.Status == 0 {
		return "running"
	} else {
		return "failure"
	}
}

func pendingAction(s Service) string {
	if s.PendingAction > 0 {
		if s.PendingAction >= len(pendingActions) {
			return "unknown"
		}
		return pendingActions[s.PendingAction-1]
	} else {
		return "none"
	}
}

func monitoringMode(s Service) string {
	switch s.MonitorMode {
	case 0:
		return "active"
	case 1:
		return "passive"
	}
	return "unknown"
}

func monitoringStatus(s Service) string {
	switch s.MonitoringStatus {
	case 1:
		return "monitored"
	case 2:
		return "initializing"
	case 4:
		return "waiting"
	}
	return "not_monitored"
}

func init() {
	inputs.Add("monit", func() telegraf.Input {
		return &Monit{}
	})
}
