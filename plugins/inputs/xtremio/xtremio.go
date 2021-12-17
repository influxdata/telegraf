package xtremio

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type XtremIO struct {
	username   string   `toml:"username"`
	password   string   `toml:"password"`
	url        string   `toml:"url"`
	collectors []string `toml:"collectors"`
	cookie     *http.Cookie
	tls.ClientConfig
	Log telegraf.Logger `toml:"-"`
}

const sampleConfig = `
  ## XtremIO User Interface Endpoint
  url = "https://xtremio.example.com/" # required

  ## Credentials
  username = "user1"
  password = "pass123"

  ## Metrics to collect from the XtremIO
  # collectors = ["bbus","clusters","ssds","volumes","xms"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
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
	if xio.username == "" {
		return fmt.Errorf("Username cannot be empty")
	}
	if xio.password == "" {
		return fmt.Errorf("Password cannot be empty")
	}
	if xio.url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(xio.collectors) == 0 {
		xio.collectors = []string{"bbus", "clusters", "ssds", "volumes", "xms"}
	}

	availableCollectors := []string{"bbus", "clusters", "ssds", "volumes", "xms"}
	for _, collector := range xio.collectors {
		if !choice.Contains(collector, availableCollectors) {
			return fmt.Errorf("Specified Collector Isn't Supported: " + collector)
		}
	}

	tlsCfg, err := xio.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = tlsCfg

	return nil
}

func (xio *XtremIO) Gather(acc telegraf.Accumulator) error {
	var err = xio.authenticate()
	if err != nil {
		return err
	}
	if xio.cookie == nil {
		return fmt.Errorf("no authentication cookie set")
	}

	var wg sync.WaitGroup
	for _, collector := range xio.collectors {
		wg.Add(1)
		go func(collector string) {
			defer wg.Done()

			resp, err := xio.call(collector)
			if err != nil {
				acc.AddError(err)
				return
			}

			// Due to an inconsistency in the XtremIO API, the XMS endpoint
			// returns a json array with XMSS as the result. Which is why this
			// if statement here exists
			var arr []gjson.Result
			if collector == "xms" {
				arr = gjson.Get(resp, "xmss").Array()
			} else {
				arr = gjson.Get(resp, collector).Array()
			}

			for _, item := range arr {
				itemSplit := strings.Split(gjson.Get(item.Raw, "href").Str, "/")
				url := ""
				if len(itemSplit) < 1 {
					continue
				}
				url = collector + "/" + itemSplit[len(itemSplit)-1]

				// Each collector is ran in a goroutine so they can be run in parallel.
				// Each collector does an initial query to build out the subqueries it
				// needs to run, which are started here in nested goroutines. A future
				// refactor opportunity would be for the intial collector goroutines to 
				// return the results while exiting the goroutine, and then a series of
				// goroutines can be kicked off for the subqueries. That way there is no
				// nesting of goroutines.
				switch collector {
				case "bbus":
					wg.Add(1)
					go xio.gatherBBUs(acc, url, &wg)
				case "clusters":
					wg.Add(1)
					go xio.gatherClusters(acc, url, &wg)
				case "ssds":
					wg.Add(1)
					go xio.gatherSSDs(acc, url, &wg)
				case "volumes":
					wg.Add(1)
					go xio.gatherVolumes(acc, url, &wg)
				case "xms":
					wg.Add(1)
					go xio.gatherXMS(acc, url, &wg)
				default:
					acc.AddError(fmt.Errorf("Specified Collector Isn't Supported: " + collector))
				}
			}
		}(collector)
	}
	wg.Wait()

	xio.cookie = nil

	return nil
}

func (xio *XtremIO) gatherBBUs(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
	}
	tags := map[string]string{
		"serial_number": gjson.Get(resp, "content.serial-number").Str,
		"guid":          gjson.Get(resp, "content.guid").Str,
		"power_feed":    gjson.Get(resp, "content.power-feed").Str,
		"name":          gjson.Get(resp, "content.name").Str,
		"model_name":    gjson.Get(resp, "content.model-name").Str,
	}
	fields := make(map[string]interface{})
	fields["bbus_power"], err = strconv.Atoi(gjson.Get(resp, "content.power").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["bbus_average_daily_temp"], err = strconv.Atoi(gjson.Get(resp, "content.avg-daily-temp").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["bbus_enabled"] = (gjson.Get(resp, "content.enabled-state").Str == "enabled")
	fields["bbus_ups_need_battery_replacement"] = (gjson.Get(resp, "content.ups-need-battery-replacement").Str == "true")
	fields["bbus_ups_low_battery_no_input"] = (gjson.Get(resp, "content.is-low-battery-no-input").Str == "true")

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherClusters(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
	}
	tags := map[string]string{
		"hardware_platform":      gjson.Get(resp, "content.hardware-platform").Str,
		"license_id":             gjson.Get(resp, "content.license-id").Str,
		"guid":                   gjson.Get(resp, "content.guid").Str,
		"name":                   gjson.Get(resp, "content.name").Str,
		"sys_psnt_serial_number": gjson.Get(resp, "content.sys-psnt-serial-number").Str,
	}
	fields := make(map[string]interface{})
	fields["clusters_compression_factor"], err = strconv.ParseFloat(gjson.Get(resp, "content.compression-factor").Raw, 64)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_percent_memory_in_use"], err = strconv.Atoi(gjson.Get(resp, "content.total-memory-in-use-in-percent").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_read_iops"], err = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_write_iops"], err = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_number_of_volumes"], err = strconv.Atoi(gjson.Get(resp, "content.num-of-vols").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_free_ssd_space_in_percent"], err = strconv.Atoi(gjson.Get(resp, "content.free-ud-ssd-space-in-percent").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_ssd_num"], err = strconv.Atoi(gjson.Get(resp, "content.num-of-ssds").Raw)
	if err != nil {
		acc.AddError(err)
	}

	fields["clusters_data_reduction_ratio"], err = strconv.ParseFloat(gjson.Get(resp, "content.data-reduction-ratio").Raw, 64)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherSSDs(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
	}
	tags := map[string]string{
		"model_name":       gjson.Get(resp, "content.model-name").Str,
		"firmware_version": gjson.Get(resp, "content.fw-version").Str,
		"ssd_uid":          gjson.Get(resp, "content.ssd-uid").Str,
		"guid":             gjson.Get(resp, "content.guid").Str,
		"sys_name":         gjson.Get(resp, "content.sys-name").Str,
		"serial_number":    gjson.Get(resp, "content.serial-number").Str,
	}
	fields := make(map[string]interface{})
	fields["ssds_ssd_size"], err = strconv.Atoi(gjson.Get(resp, "content.ssd-size").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_ssd_space_in_use"], err = strconv.Atoi(gjson.Get(resp, "content.ssd-space-in-use").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_write_iops"], err = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_read_iops"], err = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_write_bandwidth"], err = strconv.Atoi(gjson.Get(resp, "content.wr-bw").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_read_bandwidth"], err = strconv.Atoi(gjson.Get(resp, "content.rd-bw").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["ssds_num_bad_sectors"], err = strconv.Atoi(gjson.Get(resp, "content.num-bad-sectors").Raw)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherVolumes(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
	}
	tags := map[string]string{
		"guid":     gjson.Get(resp, "content.guid").Str,
		"sys_name": gjson.Get(resp, "content.sys-name").Str,
		"name":     gjson.Get(resp, "content.name").Str,
	}
	fields := make(map[string]interface{})
	fields["volumes_read_iops"], err = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_write_iops"], err = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_read_latency"], err = strconv.Atoi(gjson.Get(resp, "content.rd-latency").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_write_latency"], err = strconv.Atoi(gjson.Get(resp, "content.wr-latency").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_data_reduction_ratio"], err = strconv.ParseFloat(gjson.Get(resp, "content.data-reduction-ratio").Str, 64)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_provisioned_space"], err = strconv.Atoi(gjson.Get(resp, "content.vol-size").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["volumes_used_space"], err = strconv.Atoi(gjson.Get(resp, "content.logical-space-in-use").Str)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherXMS(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
	}
	tags := map[string]string{
		"guid":    gjson.Get(resp, "content.guid").Str,
		"name":    gjson.Get(resp, "content.name").Str,
		"version": gjson.Get(resp, "content.version").Str,
		"xms_ip":  gjson.Get(resp, "content.xms-ip").Str,
	}
	fields := make(map[string]interface{})
	fields["xms_write_iops"], err = strconv.Atoi(gjson.Get(resp, "content.wr-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_read_iops"], err = strconv.Atoi(gjson.Get(resp, "content.rd-iops").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_overall_efficiency_ratio"], err = strconv.ParseFloat(gjson.Get(resp, "content.overall-efficiency-ratio").Str, 64)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_ssd_space_in_use"], err = strconv.Atoi(gjson.Get(resp, "content.ssd-space-in-use").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_ram_in_use"], err = strconv.Atoi(gjson.Get(resp, "content.ram-usage").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_ram_total"], err = strconv.Atoi(gjson.Get(resp, "content.ram-total").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_cpu_usage_total"], err = strconv.ParseFloat(gjson.Get(resp, "content.cpu").Raw, 64)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_write_latency"], err = strconv.Atoi(gjson.Get(resp, "content.wr-latency").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_read_latency"], err = strconv.Atoi(gjson.Get(resp, "content.rd-latency").Str)
	if err != nil {
		acc.AddError(err)
	}

	fields["xms_user_accounts_count"], err = strconv.Atoi(gjson.Get(resp, "content.num-of-user-accounts").Raw)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) call(endpoint string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", xio.url+"/api/json/v3/types/"+endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(xio.cookie)
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

func (xio *XtremIO) authenticate() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", xio.url+"/api/json/v3/commands/login?password="+xio.password+"&user="+xio.username, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessid" {
			xio.cookie = cookie
		}
	}
	return nil
}

func init() {
	inputs.Add("xtremio", func() telegraf.Input {
		return &XtremIO{cookie: nil}
	})
}
