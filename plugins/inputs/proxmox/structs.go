package proxmox

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Proxmox struct {
	BaseURL         string            `toml:"base_url"`
	APIToken        string            `toml:"api_token"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	NodeName        string            `toml:"node_name"`

	tls.ClientConfig

	httpClient       *http.Client
	nodeSearchDomain string

	requestFunction func(px *Proxmox, apiUrl string, method string, data url.Values) ([]byte, error)
	Log             telegraf.Logger `toml:"-"`
}

type ResourceType string

var (
	QEMU ResourceType = "qemu"
	LXC  ResourceType = "lxc"
)

type VmStats struct {
	Data []VmStat `json:"data"`
}

type VmCurrentStats struct {
	Data VmStat `json:"data"`
}

type VmStat struct {
	ID        string      `json:"vmid"`
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
		Template     int    `json:"template"`
	} `json:"data"`
}

type NodeDns struct {
	Data struct {
		Searchdomain string `json:"search"`
	} `json:"data"`
}
