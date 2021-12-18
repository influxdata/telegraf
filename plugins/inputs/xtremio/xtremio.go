package xtremio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type XtremIO struct {
	Username   string   `toml:"username"`
	Password   string   `toml:"password"`
	URL        string   `toml:"url"`
	Collectors []string `toml:"collectors"`
	cookie     *http.Cookie
	client     *http.Client
	tls.ClientConfig
	Log telegraf.Logger `toml:"-"`
}

type BBU struct {
	Content struct {
		Serial       string `json:"serial-number"`
		Guid         string `json:"guid"`
		PowerFeed    string `json:"power-feed"`
		Name         string `json:"Name"`
		ModelName    string `json:"model-name"`
		BBUPower     int    `json:"power"`
		BBUDailyTemp int    `json:"avg-daily-temp"`
		BBUEnabled   string `json:"enabled-state"`
		BBUNeedBat   string `json:"ups-need-battery-replacement"`
		BBULowBat    string `json:"is-low-battery-no-input"`
	}
}

type Clusters struct {
	Content struct {
		HardwarePlatform   string  `json:"hardware-platform"`
		LicenseId          string  `json:"license-id"`
		Guid               string  `json:"guid"`
		Name               string  `json:"name"`
		SerialNumber       string  `json:"sys-psnt-serial-number"`
		CompressionFactor  float64 `json:"compression-factor"`
		MemoryUsed         int     `json:"total-memory-in-use-in-percent"`
		ReadIops           int     `json:"rd-iops,string"`
		WriteIops          int     `json:"wr-iops,string"`
		NumVolumes         int     `json:"num-of-vols"`
		FreeSSDSpace       int     `json:"free-ud-ssd-space-in-percent"`
		NumSSDs            int     `json:"num-of-ssds"`
		DataReductionRatio float64 `json:"data-reduction-ratio"`
	}
}

type SSD struct {
	Content struct {
		ModelName       string `json:"model-name"`
		FirmwareVersion string `json:"fw-version"`
		SSDuid          string `json:"ssd-uid"`
		Guid            string `json:"guid"`
		SysName         string `json:"sys-name"`
		SerialNumber    string `json:"serial-number"`
		Size            int    `json:"ssd-size,string"`
		SpaceUsed       int    `json:"ssd-space-in-use,string"`
		WriteIops       int    `json:"wr-iops,string"`
		ReadIops        int    `json:"rd-iops,string"`
		WriteBandwidth  int    `json:"wr-bw,string"`
		ReadBandwidth   int    `json:"rd-bw,string"`
		NumBadSectors   int    `json:"num-bad-sectors"`
	}
}

type Volumes struct {
	Content struct {
		Guid               string  `json:"guid"`
		SysName            string  `json:"sys-name"`
		Name               string  `json:"name"`
		ReadIops           int     `json:"rd-iops,string"`
		WriteIops          int     `json:"wr-iops,string"`
		ReadLatency        int     `json:"rd-latency,string"`
		WriteLatency       int     `json:"wr-latency,string"`
		DataReductionRatio float64 `json:"data-reduction-ratio,string"`
		ProvisionedSpace   int     `json:"vol-size,string"`
		UsedSpace          int     `json:"logical-space-in-use,string"`
	}
}

type XMS struct {
	Content struct {
		Guid            string  `json:"guid"`
		Name            string  `json:"name"`
		Version         string  `json:"version"`
		IP              string  `json:"xms-ip"`
		WriteIops       int     `json:"wr-iops,string"`
		ReadIops        int     `json:"rd-iops,string"`
		EfficiencyRatio float64 `json:"overall-efficiency-ratio,string"`
		SpaceUsed       int     `json:"ssd-space-in-use,string"`
		RamUsage        int     `json:"ram-usage,string"`
		RamTotal        int     `json:"ram-total,string"`
		CpuUsage        float64 `json:"cpu"`
		WriteLatency    int     `json:"wr-latency,string"`
		ReadLatency     int     `json:"rd-latency,string"`
		NumAccounts     int     `json:"num-of-user-accounts"`
	}
}

type HREF struct {
	Href string `json:"href"`
}

type CollectorResponse struct {
	BBUs     []HREF `json:"bbus"`
	Clusters []HREF `json:"clusters"`
	SSDs     []HREF `json:"ssds"`
	Volumes  []HREF `json:"volumes"`
	XMS      []HREF `json:"xmss"`
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
	if xio.Username == "" {
		return errors.New("username cannot be empty")
	}
	if xio.Password == "" {
		return errors.New("password cannot be empty")
	}
	if xio.URL == "" {
		return errors.New("url cannot be empty")
	}

	availableCollectors := []string{"bbus", "clusters", "ssds", "volumes", "xms"}
	if len(xio.Collectors) == 0 {
		xio.Collectors = availableCollectors
	}

	for _, collector := range xio.Collectors {
		if !choice.Contains(collector, availableCollectors) {
			return fmt.Errorf("specified collector %q isn't supported", collector)
		}
	}

	tlsCfg, err := xio.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	xio.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	return nil
}

func (xio *XtremIO) Gather(acc telegraf.Accumulator) error {
	if err := xio.authenticate(); err != nil {
		return err
	}
	if xio.cookie == nil {
		return errors.New("no authentication cookie set")
	}

	var wg sync.WaitGroup
	for _, collector := range xio.Collectors {
		wg.Add(1)
		go func(collector string) {
			defer wg.Done()

			resp, err := xio.call(collector)
			if err != nil {
				acc.AddError(err)
				return
			}

			data := CollectorResponse{}
			err = json.Unmarshal([]byte(resp), &data)
			if err != nil {
				acc.AddError(err)
			}

			var arr []HREF
			switch collector {
			case "bbus":
				arr = data.BBUs
			case "clusters":
				arr = data.Clusters
			case "ssds":
				arr = data.SSDs
			case "volumes":
				arr = data.Volumes
			case "xms":
				arr = data.XMS
			}

			for _, item := range arr {
				itemSplit := strings.Split(item.Href, "/")
				if len(itemSplit) < 1 {
					continue
				}
				url := collector + "/" + itemSplit[len(itemSplit)-1]

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
					acc.AddError(fmt.Errorf("specified collector %q isn't supported", collector))
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
		return
	}

	data := BBU{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"serial_number": data.Content.Serial,
		"guid":          data.Content.Guid,
		"power_feed":    data.Content.PowerFeed,
		"name":          data.Content.Name,
		"model_name":    data.Content.ModelName,
	}
	fields := make(map[string]interface{})
	fields["bbus_power"] = data.Content.BBUPower
	fields["bbus_average_daily_temp"] = data.Content.BBUDailyTemp
	fields["bbus_enabled"] = (data.Content.BBUEnabled == "enabled")
	fields["bbus_ups_need_battery_replacement"] = (data.Content.BBUNeedBat == "true")
	fields["bbus_ups_low_battery_no_input"] = (data.Content.BBULowBat == "true")

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherClusters(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
		return
	}

	data := Clusters{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"hardware_platform":      data.Content.HardwarePlatform,
		"license_id":             data.Content.LicenseId,
		"guid":                   data.Content.Guid,
		"name":                   data.Content.Name,
		"sys_psnt_serial_number": data.Content.SerialNumber,
	}
	fields := make(map[string]interface{})
	fields["clusters_compression_factor"] = data.Content.CompressionFactor
	fields["clusters_percent_memory_in_use"] = data.Content.MemoryUsed
	fields["clusters_read_iops"] = data.Content.ReadIops
	fields["clusters_write_iops"] = data.Content.WriteIops
	fields["clusters_number_of_volumes"] = data.Content.NumVolumes
	fields["clusters_free_ssd_space_in_percent"] = data.Content.FreeSSDSpace
	fields["clusters_ssd_num"] = data.Content.NumSSDs
	fields["clusters_data_reduction_ratio"] = data.Content.DataReductionRatio

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherSSDs(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
		return
	}

	data := SSD{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"model_name":       data.Content.ModelName,
		"firmware_version": data.Content.FirmwareVersion,
		"ssd_uid":          data.Content.SSDuid,
		"guid":             data.Content.Guid,
		"sys_name":         data.Content.SysName,
		"serial_number":    data.Content.SerialNumber,
	}
	fields := make(map[string]interface{})
	fields["ssds_ssd_size"] = data.Content.Size
	fields["ssds_ssd_space_in_use"] = data.Content.SpaceUsed
	fields["ssds_write_iops"] = data.Content.WriteIops
	fields["ssds_read_iops"] = data.Content.ReadIops
	fields["ssds_write_bandwidth"] = data.Content.WriteBandwidth
	fields["ssds_read_bandwidth"] = data.Content.ReadBandwidth
	fields["ssds_num_bad_sectors"] = data.Content.NumBadSectors

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherVolumes(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
		return
	}

	data := Volumes{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"guid":     data.Content.Guid,
		"sys_name": data.Content.SysName,
		"name":     data.Content.Name,
	}
	fields := make(map[string]interface{})
	fields["volumes_read_iops"] = data.Content.ReadIops
	fields["volumes_write_iops"] = data.Content.WriteIops
	fields["volumes_read_latency"] = data.Content.ReadLatency
	fields["volumes_write_latency"] = data.Content.WriteLatency
	fields["volumes_data_reduction_ratio"] = data.Content.DataReductionRatio
	fields["volumes_provisioned_space"] = data.Content.ProvisionedSpace
	fields["volumes_used_space"] = data.Content.UsedSpace

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) gatherXMS(acc telegraf.Accumulator, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := xio.call(url)
	if err != nil {
		acc.AddError(err)
		return
	}

	data := XMS{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"guid":    data.Content.Guid,
		"name":    data.Content.Name,
		"version": data.Content.Version,
		"xms_ip":  data.Content.IP,
	}
	fields := make(map[string]interface{})
	fields["xms_write_iops"] = data.Content.WriteIops
	fields["xms_read_iops"] = data.Content.ReadIops
	fields["xms_overall_efficiency_ratio"] = data.Content.EfficiencyRatio
	fields["xms_ssd_space_in_use"] = data.Content.SpaceUsed
	fields["xms_ram_in_use"] = data.Content.RamUsage
	fields["xms_ram_total"] = data.Content.RamTotal
	fields["xms_cpu_usage_total"] = data.Content.CpuUsage
	fields["xms_write_latency"] = data.Content.WriteLatency
	fields["xms_read_latency"] = data.Content.ReadLatency
	fields["xms_user_accounts_count"] = data.Content.NumAccounts

	acc.AddFields("xio", fields, tags)
}

func (xio *XtremIO) call(endpoint string) (string, error) {
	req, err := http.NewRequest("GET", xio.URL+"/api/json/v3/types/"+endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(xio.cookie)
	resp, err := xio.client.Do(req)
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
	req, err := http.NewRequest("GET", xio.URL+"/api/json/v3/commands/login?password="+xio.Password+"&user="+xio.Username, nil)
	if err != nil {
		return err
	}
	resp, err := xio.client.Do(req)
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
		return &XtremIO{}
	})
}
