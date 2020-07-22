package proxmox

import (
	"encoding/json"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Proxmox struct {
	BaseUrl         string            `toml:"base_url"`
	ApiToken        string            `toml:"api_token"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	tls.ClientConfig

	hostname         string
	nodeSearchDomain string
}

type VmStats struct {
	Data []VmStat `json:"data"`
}

type VmStat struct {
	Id        string      `json:"vmid"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	UsedMem   json.Number `json:"mem"`
	TotalMem  json.Number `json:"maxmem"`
	UsedDisk  json.Number `json:"disk"`
	TotalDisk json.Number `json:"maxdisk"`
	UsedSwap  json.Number `json:"swap"`
	TotalSwap json.Number `json:"maxswap"`
	Uptime    json.Number `json:"uptime"`
	CpuLoad   json.Number `json:"cpu"`
}

type VmConfig struct {
	Data struct {
		Searchdomain string `json:"searchdomain"`
		Hostname     string `json:"hostname"`
	} `json:"data"`
}

type NodeDns struct {
	Data struct {
		Searchdomain string `json:"search"`
	} `json:"data"`
}
