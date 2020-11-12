package proxmox

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## API connection configuration. The API token was introduced in Proxmox v6.2. Required permissions for user and token: PVEAuditor role on /.
  base_url = "https://localhost:8006/api2/json"
  api_token = "USER@REALM!TOKENID=UUID"
  ## Node name, defaults to OS hostname
  # node_name = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
`

func (px *Proxmox) SampleConfig() string {
	return sampleConfig
}

func (px *Proxmox) Description() string {
	return "Provides metrics from Proxmox nodes (Proxmox Virtual Environment > 6.2)."
}

func (px *Proxmox) Gather(acc telegraf.Accumulator) error {
	err := getNodeSearchDomain(px)
	if err != nil {
		return err
	}

	gatherLxcData(px, acc)
	gatherQemuData(px, acc)

	return nil
}

func (px *Proxmox) Init() error {
	// Set hostname as default node name for backwards compatibility
	if px.NodeName == "" {
		hostname, _ := os.Hostname()
		px.NodeName = hostname
	}

	tlsCfg, err := px.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	px.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: px.ResponseTimeout.Duration,
	}

	return nil
}

func init() {
	inputs.Add("proxmox", func() telegraf.Input {
		return &Proxmox{
			requestFunction: performRequest,
		}
	})
}

func getNodeSearchDomain(px *Proxmox) error {
	apiUrl := "/nodes/" + px.NodeName + "/dns"
	jsonData, err := px.requestFunction(px, apiUrl, http.MethodGet, nil)
	if err != nil {
		return err
	}

	var nodeDns NodeDns
	err = json.Unmarshal(jsonData, &nodeDns)
	if err != nil {
		return err
	}

	if nodeDns.Data.Searchdomain == "" {
		return errors.New("search domain is not set")
	}
	px.nodeSearchDomain = nodeDns.Data.Searchdomain

	return nil
}

func performRequest(px *Proxmox, apiUrl string, method string, data url.Values) ([]byte, error) {
	request, err := http.NewRequest(method, px.BaseURL+apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "PVEAPIToken="+px.APIToken)

	resp, err := px.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func gatherLxcData(px *Proxmox, acc telegraf.Accumulator) {
	gatherVmData(px, acc, LXC)
}

func gatherQemuData(px *Proxmox, acc telegraf.Accumulator) {
	gatherVmData(px, acc, QEMU)
}

func gatherVmData(px *Proxmox, acc telegraf.Accumulator, rt ResourceType) {
	vmStats, err := getVmStats(px, rt)
	if err != nil {
		px.Log.Error("Error getting VM stats: %v", err)
		return
	}

	// For each VM add metrics to Accumulator
	for _, vmStat := range vmStats.Data {
		vmConfig, err := getVmConfig(px, vmStat.ID, rt)
		if err != nil {
			px.Log.Errorf("Error getting VM config: %v", err)
			return
		}

		if vmConfig.Data.Template == 1 {
			px.Log.Debugf("Ignoring template VM %s (%s)", vmStat.ID, vmStat.Name)
			continue
		}

		tags := getTags(px, vmStat.Name, vmConfig, rt)
		currentVMStatus, err := getCurrentVMStatus(px, rt, vmStat.ID)
		if err != nil {
			px.Log.Errorf("Error getting VM curent VM status: %v", err)
			return
		}

		fields, err := getFields(currentVMStatus)
		if err != nil {
			px.Log.Errorf("Error getting VM measurements: %v", err)
			return
		}

		acc.AddFields("proxmox", fields, tags)
	}
}

func getCurrentVMStatus(px *Proxmox, rt ResourceType, id string) (VmStat, error) {
	apiUrl := "/nodes/" + px.NodeName + "/" + string(rt) + "/" + id + "/status/current"

	jsonData, err := px.requestFunction(px, apiUrl, http.MethodGet, nil)
	if err != nil {
		return VmStat{}, err
	}

	var currentVmStatus VmCurrentStats
	err = json.Unmarshal(jsonData, &currentVmStatus)
	if err != nil {
		return VmStat{}, err
	}

	return currentVmStatus.Data, nil
}

func getVmStats(px *Proxmox, rt ResourceType) (VmStats, error) {
	apiUrl := "/nodes/" + px.NodeName + "/" + string(rt)
	jsonData, err := px.requestFunction(px, apiUrl, http.MethodGet, nil)
	if err != nil {
		return VmStats{}, err
	}

	var vmStats VmStats
	err = json.Unmarshal(jsonData, &vmStats)
	if err != nil {
		return VmStats{}, err
	}

	return vmStats, nil
}

func getVmConfig(px *Proxmox, vmId string, rt ResourceType) (VmConfig, error) {
	apiUrl := "/nodes/" + px.NodeName + "/" + string(rt) + "/" + vmId + "/config"
	jsonData, err := px.requestFunction(px, apiUrl, http.MethodGet, nil)
	if err != nil {
		return VmConfig{}, err
	}

	var vmConfig VmConfig
	err = json.Unmarshal(jsonData, &vmConfig)
	if err != nil {
		return VmConfig{}, err
	}

	return vmConfig, nil
}

func getFields(vmStat VmStat) (map[string]interface{}, error) {
	mem_total, mem_used, mem_free, mem_used_percentage := getByteMetrics(vmStat.TotalMem, vmStat.UsedMem)
	swap_total, swap_used, swap_free, swap_used_percentage := getByteMetrics(vmStat.TotalSwap, vmStat.UsedSwap)
	disk_total, disk_used, disk_free, disk_used_percentage := getByteMetrics(vmStat.TotalDisk, vmStat.UsedDisk)

	return map[string]interface{}{
		"status":               vmStat.Status,
		"uptime":               jsonNumberToInt64(vmStat.Uptime),
		"cpuload":              jsonNumberToFloat64(vmStat.CpuLoad),
		"mem_used":             mem_used,
		"mem_total":            mem_total,
		"mem_free":             mem_free,
		"mem_used_percentage":  mem_used_percentage,
		"swap_used":            swap_used,
		"swap_total":           swap_total,
		"swap_free":            swap_free,
		"swap_used_percentage": swap_used_percentage,
		"disk_used":            disk_used,
		"disk_total":           disk_total,
		"disk_free":            disk_free,
		"disk_used_percentage": disk_used_percentage,
	}, nil
}

func getByteMetrics(total json.Number, used json.Number) (int64, int64, int64, float64) {
	int64Total := jsonNumberToInt64(total)
	int64Used := jsonNumberToInt64(used)
	int64Free := int64Total - int64Used
	usedPercentage := 0.0
	if int64Total != 0 {
		usedPercentage = float64(int64Used) * 100 / float64(int64Total)
	}

	return int64Total, int64Used, int64Free, usedPercentage
}

func jsonNumberToInt64(value json.Number) int64 {
	int64Value, err := value.Int64()
	if err != nil {
		return 0
	}

	return int64Value
}

func jsonNumberToFloat64(value json.Number) float64 {
	float64Value, err := value.Float64()
	if err != nil {
		return 0
	}

	return float64Value
}

func getTags(px *Proxmox, name string, vmConfig VmConfig, rt ResourceType) map[string]string {
	domain := vmConfig.Data.Searchdomain
	if len(domain) == 0 {
		domain = px.nodeSearchDomain
	}

	hostname := vmConfig.Data.Hostname
	if len(hostname) == 0 {
		hostname = name
	}
	fqdn := hostname + "." + domain

	return map[string]string{
		"node_fqdn": px.NodeName + "." + px.nodeSearchDomain,
		"vm_name":   name,
		"vm_fqdn":   fqdn,
		"vm_type":   string(rt),
	}
}
