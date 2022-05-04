package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

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
		Timeout: time.Duration(px.ResponseTimeout),
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
	apiURL := "/nodes/" + px.NodeName + "/dns"
	jsonData, err := px.requestFunction(px, apiURL, http.MethodGet, nil)
	if err != nil {
		return err
	}

	var nodeDNS NodeDNS
	err = json.Unmarshal(jsonData, &nodeDNS)
	if err != nil {
		return err
	}

	if nodeDNS.Data.Searchdomain == "" {
		return errors.New("search domain is not set")
	}
	px.nodeSearchDomain = nodeDNS.Data.Searchdomain

	return nil
}

func performRequest(px *Proxmox, apiURL string, method string, data url.Values) ([]byte, error) {
	request, err := http.NewRequest(method, px.BaseURL+apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "PVEAPIToken="+px.APIToken)

	resp, err := px.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func gatherLxcData(px *Proxmox, acc telegraf.Accumulator) {
	gatherVMData(px, acc, LXC)
}

func gatherQemuData(px *Proxmox, acc telegraf.Accumulator) {
	gatherVMData(px, acc, QEMU)
}

func gatherVMData(px *Proxmox, acc telegraf.Accumulator, rt ResourceType) {
	vmStats, err := getVMStats(px, rt)
	if err != nil {
		px.Log.Error("Error getting VM stats: %v", err)
		return
	}

	// For each VM add metrics to Accumulator
	for _, vmStat := range vmStats.Data {
		vmConfig, err := getVMConfig(px, vmStat.ID, rt)
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

		fields := getFields(currentVMStatus)
		acc.AddFields("proxmox", fields, tags)
	}
}

func getCurrentVMStatus(px *Proxmox, rt ResourceType, id json.Number) (VMStat, error) {
	apiURL := "/nodes/" + px.NodeName + "/" + string(rt) + "/" + string(id) + "/status/current"

	jsonData, err := px.requestFunction(px, apiURL, http.MethodGet, nil)
	if err != nil {
		return VMStat{}, err
	}

	var currentVMStatus VMCurrentStats
	err = json.Unmarshal(jsonData, &currentVMStatus)
	if err != nil {
		return VMStat{}, err
	}

	return currentVMStatus.Data, nil
}

func getVMStats(px *Proxmox, rt ResourceType) (VMStats, error) {
	apiURL := "/nodes/" + px.NodeName + "/" + string(rt)
	jsonData, err := px.requestFunction(px, apiURL, http.MethodGet, nil)
	if err != nil {
		return VMStats{}, err
	}

	var vmStats VMStats
	err = json.Unmarshal(jsonData, &vmStats)
	if err != nil {
		return VMStats{}, err
	}

	return vmStats, nil
}

func getVMConfig(px *Proxmox, vmID json.Number, rt ResourceType) (VMConfig, error) {
	apiURL := "/nodes/" + px.NodeName + "/" + string(rt) + "/" + string(vmID) + "/config"
	jsonData, err := px.requestFunction(px, apiURL, http.MethodGet, nil)
	if err != nil {
		return VMConfig{}, err
	}

	var vmConfig VMConfig
	err = json.Unmarshal(jsonData, &vmConfig)
	if err != nil {
		return VMConfig{}, err
	}

	return vmConfig, nil
}

func getFields(vmStat VMStat) map[string]interface{} {
	memMetrics := getByteMetrics(vmStat.TotalMem, vmStat.UsedMem)
	swapMetrics := getByteMetrics(vmStat.TotalSwap, vmStat.UsedSwap)
	diskMetrics := getByteMetrics(vmStat.TotalDisk, vmStat.UsedDisk)

	return map[string]interface{}{
		"status":               vmStat.Status,
		"uptime":               jsonNumberToInt64(vmStat.Uptime),
		"cpuload":              jsonNumberToFloat64(vmStat.CPULoad),
		"mem_used":             memMetrics.used,
		"mem_total":            memMetrics.total,
		"mem_free":             memMetrics.free,
		"mem_used_percentage":  memMetrics.usedPercentage,
		"swap_used":            swapMetrics.used,
		"swap_total":           swapMetrics.total,
		"swap_free":            swapMetrics.free,
		"swap_used_percentage": swapMetrics.usedPercentage,
		"disk_used":            diskMetrics.used,
		"disk_total":           diskMetrics.total,
		"disk_free":            diskMetrics.free,
		"disk_used_percentage": diskMetrics.usedPercentage,
	}
}

func getByteMetrics(total json.Number, used json.Number) metrics {
	int64Total := jsonNumberToInt64(total)
	int64Used := jsonNumberToInt64(used)
	int64Free := int64Total - int64Used
	usedPercentage := 0.0
	if int64Total != 0 {
		usedPercentage = float64(int64Used) * 100 / float64(int64Total)
	}

	return metrics{
		total:          int64Total,
		used:           int64Used,
		free:           int64Free,
		usedPercentage: usedPercentage,
	}
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

func getTags(px *Proxmox, name string, vmConfig VMConfig, rt ResourceType) map[string]string {
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
