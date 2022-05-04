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
	Username   string          `toml:"username"`
	Password   string          `toml:"password"`
	URL        string          `toml:"url"`
	Collectors []string        `toml:"collectors"`
	Log        telegraf.Logger `toml:"-"`
	tls.ClientConfig

	cookie *http.Cookie
	client *http.Client
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

	// At the beginning of every collection, we re-authenticate.
	// We reset this cookie so we don't accidentally use an
	// expired cookie, we can just check if it's nil and know
	// that we either need to re-authenticate or that the
	// authentication failed to set the cookie.
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
		"guid":          data.Content.GUID,
		"power_feed":    data.Content.PowerFeed,
		"name":          data.Content.Name,
		"model_name":    data.Content.ModelName,
	}
	fields := map[string]interface{}{
		"bbus_power":                        data.Content.BBUPower,
		"bbus_average_daily_temp":           data.Content.BBUDailyTemp,
		"bbus_enabled":                      (data.Content.BBUEnabled == "enabled"),
		"bbus_ups_need_battery_replacement": data.Content.BBUNeedBat,
		"bbus_ups_low_battery_no_input":     data.Content.BBULowBat,
	}

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
		"license_id":             data.Content.LicenseID,
		"guid":                   data.Content.GUID,
		"name":                   data.Content.Name,
		"sys_psnt_serial_number": data.Content.SerialNumber,
	}
	fields := map[string]interface{}{
		"clusters_compression_factor":        data.Content.CompressionFactor,
		"clusters_percent_memory_in_use":     data.Content.MemoryUsed,
		"clusters_read_iops":                 data.Content.ReadIops,
		"clusters_write_iops":                data.Content.WriteIops,
		"clusters_number_of_volumes":         data.Content.NumVolumes,
		"clusters_free_ssd_space_in_percent": data.Content.FreeSSDSpace,
		"clusters_ssd_num":                   data.Content.NumSSDs,
		"clusters_data_reduction_ratio":      data.Content.DataReductionRatio,
	}

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
		"guid":             data.Content.GUID,
		"sys_name":         data.Content.SysName,
		"serial_number":    data.Content.SerialNumber,
	}
	fields := map[string]interface{}{
		"ssds_ssd_size":         data.Content.Size,
		"ssds_ssd_space_in_use": data.Content.SpaceUsed,
		"ssds_write_iops":       data.Content.WriteIops,
		"ssds_read_iops":        data.Content.ReadIops,
		"ssds_write_bandwidth":  data.Content.WriteBandwidth,
		"ssds_read_bandwidth":   data.Content.ReadBandwidth,
		"ssds_num_bad_sectors":  data.Content.NumBadSectors,
	}

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
		"guid":     data.Content.GUID,
		"sys_name": data.Content.SysName,
		"name":     data.Content.Name,
	}
	fields := map[string]interface{}{
		"volumes_read_iops":            data.Content.ReadIops,
		"volumes_write_iops":           data.Content.WriteIops,
		"volumes_read_latency":         data.Content.ReadLatency,
		"volumes_write_latency":        data.Content.WriteLatency,
		"volumes_data_reduction_ratio": data.Content.DataReductionRatio,
		"volumes_provisioned_space":    data.Content.ProvisionedSpace,
		"volumes_used_space":           data.Content.UsedSpace,
	}

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
		"guid":    data.Content.GUID,
		"name":    data.Content.Name,
		"version": data.Content.Version,
		"xms_ip":  data.Content.IP,
	}
	fields := map[string]interface{}{
		"xms_write_iops":               data.Content.WriteIops,
		"xms_read_iops":                data.Content.ReadIops,
		"xms_overall_efficiency_ratio": data.Content.EfficiencyRatio,
		"xms_ssd_space_in_use":         data.Content.SpaceUsed,
		"xms_ram_in_use":               data.Content.RAMUsage,
		"xms_ram_total":                data.Content.RAMTotal,
		"xms_cpu_usage_total":          data.Content.CPUUsage,
		"xms_write_latency":            data.Content.WriteLatency,
		"xms_read_latency":             data.Content.ReadLatency,
		"xms_user_accounts_count":      data.Content.NumAccounts,
	}

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
	req, err := http.NewRequest("GET", xio.URL+"/api/json/v3/commands/login", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(xio.Username, xio.Password)
	resp, err := xio.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessid" {
			xio.cookie = cookie
			break
		}
	}
	return nil
}

func init() {
	inputs.Add("xtremio", func() telegraf.Input {
		return &XtremIO{}
	})
}
