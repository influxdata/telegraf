package xtremio

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tidwall/gjson"
)

type XtremIO struct {
	Username   string   `toml:"username"`
	Password   string   `toml:"password"`
	URL        string   `toml:"url"`
	Collectors []string `toml:"collectors"`
	Cookie     *http.Cookie
	Log        telegraf.Logger
}

const sampleConfig = `
  ## XtremIO Username
  username = "" # required
  ## XtremIO Password
  password = "" # required
  ## XtremIO User Interface Endpoint
  url = "https://xtremio.example.com/" # required
  ## Metrics to collect from the XtremIO
  collectors = ["bbus","clusters","ssds","volumes","xms"]
`

// Description will appear directly above the plugin definition in the config file
func (xio *XtremIO) Description() string {
	return `Gathers Metrics From a Dell EMC XtremIO Storage Array's V3 API`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (xio *XtremIO) SampleConfig() string {
	return sampleConfig
}

func (xio *XtremIO) Init() error {
	if xio.Username == "" {
		return fmt.Errorf("Username cannot be empty")
	}
	if xio.Password == "" {
		return fmt.Errorf("Password cannot be empty")
	}
	if xio.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if xio.Collectors == nil {
		xio.Collectors = []string{"bbus", "clusters", "ssds", "volumes", "xms"}
	}
	xio.Cookie = nil

	return nil
}

func (xio *XtremIO) Gather(acc telegraf.Accumulator) error {
	// Due to self signed certificates in many orgs, we don't verify the cert
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	err := xio.Authenticate()
	if err != nil {
		return err
	}
	if xio.Cookie == nil {
		return fmt.Errorf("no authentication cookie set")
	}

	availableCollectors := []string{"bbus", "clusters", "ssds", "volumes", "xms"}
	var wg sync.WaitGroup
	for _, collector := range xio.Collectors {
		wg.Add(1)
		go func(collector string) {
			defer wg.Done()
			if !contains(availableCollectors, collector) {
				acc.AddError(fmt.Errorf("Specified Collector Isn't Supported: " + collector))
				return
			}

			resp, err := xio.Call(collector)
			if err != nil {
				acc.AddError(err)
				return
			}
			var arr []gjson.Result
			if collector == "xms" {
				arr = gjson.Get(resp, "xmss").Array()
			} else {
				arr = gjson.Get(resp, collector).Array()
			}

			for _, item := range arr {
				itemSplit := strings.Split(gjson.Get(item.Raw, "href").Str, "/")
				url := ""
				if len(itemSplit) > 0 {
					url = collector + "/" + itemSplit[len(itemSplit)-1]
				} else {
					continue
				}

				if collector == "bbus" {
					wg.Add(1)
					go xio.GatherBBUs(&wg, url, acc)
				}
				if collector == "clusters" {
					wg.Add(1)
					go xio.GatherClusters(&wg, url, acc)
				}
				if collector == "ssds" {
					wg.Add(1)
					go xio.GatherSSDs(&wg, url, acc)
				}
				if collector == "volumes" {
					wg.Add(1)
					go xio.GatherVolumes(&wg, url, acc)
				}
				if collector == "xms" {
					wg.Add(1)
					go xio.GatherXMS(&wg, url, acc)
				}
			}
		}(collector)
	}

	wg.Wait()

	xio.ResetCookie()

	return nil
}

func (xio *XtremIO) GatherBBUs(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
	defer wg.Done()
	resp, _ := xio.Call(url)
	tags := map[string]string{
		"serial_number": gjson.Get(resp, "content.serial-number").Str,
		"guid":          gjson.Get(resp, "content.guid").Str,
		"power_feed":    gjson.Get(resp, "content.power-feed").Str,
		"name":          gjson.Get(resp, "content.name").Str,
		"model_name":    gjson.Get(resp, "content.model-name").Str,
	}
	fields := make(map[string]interface{})
	fields["bbus_power"], _ = strconv.Atoi(gjson.Get(resp, "content.power").Raw)
	fields["bbus_average_daily_temp"], _ = strconv.Atoi(gjson.Get(resp, "content.avg-daily-temp").Raw)

	fields["bbus_enabled"] = 0
	if gjson.Get(resp, "content.enabled-state").Str == "enabled" {
		fields["bbus_enabled"] = 1
	}

	fields["ups_need_battery_replacement"] = 0
	if gjson.Get(resp, "content.ups-need-battery-replacement").Str == "true" {
		fields["ups_need_battery_replacement"] = 1
	}

	fields["bbus_ups_low_battery_no_input"] = 0
	if gjson.Get(resp, "content.is-low-battery-no-input").Str == "true" {
		fields["bbus_ups_low_battery_no_input"] = 1
	}

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) GatherClusters(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
	defer wg.Done()
	resp, _ := xio.Call(url)
	tags := map[string]string{
		"hardware_platform":      gjson.Get(resp, "content.hardware-platform").Str,
		"license_id":             gjson.Get(resp, "content.license-id").Str,
		"guid":                   gjson.Get(resp, "content.guid").Str,
		"name":                   gjson.Get(resp, "content.name").Str,
		"sys_psnt_serial_number": gjson.Get(resp, "content.sys-psnt-serial-number").Str,
	}
	fields := make(map[string]interface{})
	fields["clusters_compression_factor"], _ = strconv.ParseFloat(gjson.Get(resp, "content.compression-factor").Raw, 64)
	fields["clusters_percent_memory_in_use"], _ = strconv.Atoi(gjson.Get(resp, "content.total-memory-in-use-in-percent").Raw)
	fields["clusters_read_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	fields["clusters_write_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	fields["clusters_number_of_volumes"], _ = strconv.Atoi(gjson.Get(resp, "content.num-of-vols").Raw)
	fields["clusters_free_ssd_space_in_percent"], _ = strconv.Atoi(gjson.Get(resp, "content.free-ud-ssd-space-in-percent").Raw)
	fields["clusters_ssd_num"], _ = strconv.Atoi(gjson.Get(resp, "content.num-of-ssds").Raw)
	fields["clusters_data_reduction_ratio"], _ = strconv.ParseFloat(gjson.Get(resp, "content.data-reduction-ratio").Raw, 64)

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) GatherSSDs(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
	defer wg.Done()
	resp, _ := xio.Call(url)
	tags := map[string]string{
		"model_name":       gjson.Get(resp, "content.model-name").Str,
		"firmware_version": gjson.Get(resp, "content.fw-version").Str,
		"ssd_uid":          gjson.Get(resp, "content.ssd-uid").Str,
		"guid":             gjson.Get(resp, "content.guid").Str,
		"sys_name":         gjson.Get(resp, "content.sys-name").Str,
		"serial_number":    gjson.Get(resp, "content.serial-number").Str,
	}
	fields := make(map[string]interface{})
	fields["ssds_ssd_size"], _ = strconv.Atoi(gjson.Get(resp, "content.ssd-size").Str)
	fields["ssds_ssd_space_in_use"], _ = strconv.Atoi(gjson.Get(resp, "content.ssd-space-in-use").Str)
	fields["ssds_write_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	fields["ssds_read_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	fields["ssds_write_bandwidth"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-bw").Str)
	fields["ssds_read_bandwidth"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-bw").Str)
	fields["ssds_num_bad_sectors"], _ = strconv.Atoi(gjson.Get(resp, "content.num-bad-sectors").Raw)

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) GatherVolumes(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
	defer wg.Done()
	resp, _ := xio.Call(url)
	tags := map[string]string{
		"guid":     gjson.Get(resp, "content.guid").Str,
		"sys_name": gjson.Get(resp, "content.sys-name").Str,
		"name":     gjson.Get(resp, "content.name").Str,
	}
	fields := make(map[string]interface{})
	fields["volumes_read_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	fields["volumes_write_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	fields["volumes_read_latency"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-latency").Str)
	fields["volumes_write_latency"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-latency").Str)
	fields["volumes_data_reduction_ratio"], _ = strconv.ParseFloat(gjson.Get(resp, "content.data-reduction-ratio").Str, 64)
	fields["volumes_provisioned_space"], _ = strconv.Atoi(gjson.Get(resp, "content.vol-size").Str)
	fields["volumes_used_space"], _ = strconv.Atoi(gjson.Get(resp, "content.logical-space-in-use").Str)

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) GatherXMS(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
	defer wg.Done()
	resp, _ := xio.Call(url)
	tags := map[string]string{
		"guid":    gjson.Get(resp, "content.guid").Str,
		"name":    gjson.Get(resp, "content.name").Str,
		"version": gjson.Get(resp, "content.version").Str,
		"xms_ip":  gjson.Get(resp, "content.xms-ip").Str,
	}
	fields := make(map[string]interface{})
	fields["xms_write_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	fields["xms_read_iops"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	fields["xms_overall_efficiency_ratio"], _ = strconv.ParseFloat(gjson.Get(resp, "content.overall-efficiency-ratio").Str, 64)
	fields["xms_ssd_space_in_use"], _ = strconv.Atoi(gjson.Get(resp, "content.ssd-space-in-use").Str)
	fields["xms_ram_in_use"], _ = strconv.Atoi(gjson.Get(resp, "content.ram-usage").Str)
	fields["xms_ram_total"], _ = strconv.Atoi(gjson.Get(resp, "content.ram-total").Str)
	fields["xms_cpu_usage_total"], _ = strconv.ParseFloat(gjson.Get(resp, "content.cpu").Raw, 64)
	fields["xms_write_latency"], _ = strconv.Atoi(gjson.Get(resp, "content.wr-latency").Str)
	fields["xms_read_latency"], _ = strconv.Atoi(gjson.Get(resp, "content.rd-latency").Str)
	fields["xms_user_accounts_count"], _ = strconv.Atoi(gjson.Get(resp, "content.num-of-user-accounts").Raw)

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) Call(endpoint string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", xio.URL+"/api/json/v3/types/"+endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(xio.Cookie)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (xio *XtremIO) Authenticate() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", xio.URL+"/api/json/v3/commands/login?password="+xio.Password+"&user="+xio.Username, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessid" {
			xio.Cookie = cookie
		}
	}
	return nil
}

func (xio *XtremIO) ResetCookie() {
	xio.Cookie = nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func init() {
	inputs.Add("xtremio", func() telegraf.Input {
		return &XtremIO{}
	})
}
