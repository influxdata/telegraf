package proxmox

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Proxmox struct {
	BaseURL         string          `toml:"base_url"`
	APIToken        string          `toml:"api_token"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	NodeName        string          `toml:"node_name"`

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

type VMStats struct {
	Data []VMStat `json:"data"`
}

type VMCurrentStats struct {
	Data VMStat `json:"data"`
}

type VMStat struct {
	ID        json.Number `json:"vmid"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	UsedMem   json.Number `json:"mem"`
	TotalMem  json.Number `json:"maxmem"`
	UsedDisk  json.Number `json:"disk"`
	TotalDisk json.Number `json:"maxdisk"`
	UsedSwap  json.Number `json:"swap"`
	TotalSwap json.Number `json:"maxswap"`
	Uptime    json.Number `json:"uptime"`
	CPULoad   json.Number `json:"cpu"`
}

type VMConfig struct {
	Data struct {
		Searchdomain string `json:"searchdomain"`
		Hostname     string `json:"hostname"`
		Template     int    `json:"template"`
	} `json:"data"`
}

type NodeDNS struct {
	Data struct {
		Searchdomain string `json:"search"`
	} `json:"data"`
}

type metrics struct {
	total          int64
	used           int64
	free           int64
	usedPercentage float64
}
