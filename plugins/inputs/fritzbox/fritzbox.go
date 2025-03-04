//go:generate ../../../tools/readme_config_includer/generator
package fritzbox

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tdrn-org/go-tr064"
	"github.com/tdrn-org/go-tr064/mesh"
	"github.com/tdrn-org/go-tr064/services/igddesc/igdicfg"
	"github.com/tdrn-org/go-tr064/services/tr64desc/deviceinfo"
	"github.com/tdrn-org/go-tr064/services/tr64desc/hosts"
	"github.com/tdrn-org/go-tr064/services/tr64desc/wancommonifconfig"
	"github.com/tdrn-org/go-tr064/services/tr64desc/wandslifconfig"
	"github.com/tdrn-org/go-tr064/services/tr64desc/wanpppconn"
	"github.com/tdrn-org/go-tr064/services/tr64desc/wlanconfig"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Fritzbox struct {
	URLs    []string        `toml:"urls"`
	Collect []string        `toml:"collect"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`
	tls.ClientConfig

	deviceClients   []*tr064.Client
	serviceHandlers map[string]serviceHandlerFunc
}

func (*Fritzbox) SampleConfig() string {
	return sampleConfig
}

func (f *Fritzbox) Init() error {
	// No need to run without any device URL
	if len(f.URLs) == 0 {
		return errors.New("no client URLs configured")
	}
	if f.Collect == nil {
		f.Collect = []string{"device", "wan", "ppp", "dsl", "wlan"}
	}
	if err := f.initDeviceClients(); err != nil {
		return fmt.Errorf("initializing clients failed: %w", err)
	}
	f.initServiceHandlers()
	return nil
}

func (f *Fritzbox) initDeviceClients() error {
	f.deviceClients = make([]*tr064.Client, 0, len(f.URLs))
	for _, rawUrl := range f.URLs {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			return fmt.Errorf("parsing client URL %q failed: %w", rawUrl, err)
		}
		client := tr064.NewClient(parsedUrl)
		client.Debug = f.Log.Level().Includes(telegraf.Trace)
		client.Timeout = time.Duration(f.Timeout)
		tlsConfig, err := f.TLSConfig()
		if err != nil {
			return err
		}
		client.TlsConfig = tlsConfig
		f.deviceClients = append(f.deviceClients, client)
	}
	return nil
}

func (f *Fritzbox) initServiceHandlers() {
	f.serviceHandlers = make(map[string]serviceHandlerFunc, len(f.Collect))
	if choice.Contains("device", f.Collect) {
		f.serviceHandlers[deviceinfo.ServiceShortType] = f.gatherDeviceInfo
	}
	if choice.Contains("wan", f.Collect) {
		f.serviceHandlers[wancommonifconfig.ServiceShortType] = f.gatherWanInfo
	}
	if choice.Contains("ppp", f.Collect) {
		f.serviceHandlers[wanpppconn.ServiceShortType] = f.gatherPppInfo
	}
	if choice.Contains("dsl", f.Collect) {
		f.serviceHandlers[wandslifconfig.ServiceShortType] = f.gatherDslInfo
	}
	if choice.Contains("wlan", f.Collect) {
		f.serviceHandlers[wlanconfig.ServiceShortType] = f.gatherWlanInfo
	}
	if choice.Contains("hosts", f.Collect) {
		f.serviceHandlers[hosts.ServiceShortType] = f.gatherHostsInfo
	}
}

func (f *Fritzbox) Gather(acc telegraf.Accumulator) error {
	var waitComplete sync.WaitGroup
	for _, deviceClient := range f.deviceClients {
		waitComplete.Add(1)
		go func() {
			defer waitComplete.Done()
			f.gatherDevice(acc, deviceClient)
		}()
	}
	waitComplete.Wait()
	return nil
}

func (f *Fritzbox) gatherDevice(acc telegraf.Accumulator, deviceClient *tr064.Client) {
	services, err := deviceClient.Services(tr064.DefaultServiceSpec)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, service := range services {
		serviceHandler := f.serviceHandlers[service.ShortType()]
		if serviceHandler == nil {
			continue
		}
		acc.AddError(serviceHandler(acc, deviceClient, service))
	}
}

type serviceHandlerFunc func(telegraf.Accumulator, *tr064.Client, tr064.ServiceDescriptor) error

func (f *Fritzbox) gatherDeviceInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := deviceinfo.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &deviceinfo.GetInfoResponse{}
	err := serviceClient.GetInfo(info)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
	}
	fields := map[string]interface{}{
		"uptime":           info.NewUpTime,
		"model_name":       info.NewModelName,
		"serial_number":    info.NewSerialNumber,
		"hardware_version": info.NewHardwareVersion,
		"software_version": info.NewSoftwareVersion,
	}
	acc.AddFields("fritzbox_device", fields, tags)
	return nil
}

func (f *Fritzbox) gatherWanInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wancommonifconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	commonLinkProperties := &wancommonifconfig.GetCommonLinkPropertiesResponse{}
	err := serviceClient.GetCommonLinkProperties(commonLinkProperties)
	if err != nil {
		return err
	}
	// Prefer igdicfg service over wancommonifconfig one for total bytes stats, because igdicfg supports uint64 counters
	igdServices, err := deviceClient.ServicesByType(tr064.IgdServiceSpec, igdicfg.ServiceShortType)
	if err != nil {
		return err
	}
	var totalBytesSent uint64 = 0
	var totalBytesReceived uint64 = 0
	if len(igdServices) > 0 {
		igdServiceClient := &igdicfg.ServiceClient{
			TR064Client: deviceClient,
			Service:     igdServices[0],
		}
		addonInfos := &igdicfg.GetAddonInfosResponse{}
		err = igdServiceClient.GetAddonInfos(addonInfos)
		if err != nil {
			return err
		}
		totalBytesSent, err = strconv.ParseUint(addonInfos.NewX_AVM_DE_TotalBytesSent64, 10, 64)
		if err != nil {
			return err
		}
		totalBytesReceived, err = strconv.ParseUint(addonInfos.NewX_AVM_DE_TotalBytesReceived64, 10, 64)
		if err != nil {
			return err
		}
	} else {
		// Fall back to wancommonifconfig service in case igdicfg is not available (only uint32 based)
		totalBytesSentResponse := &wancommonifconfig.GetTotalBytesSentResponse{}
		err = serviceClient.GetTotalBytesSent(totalBytesSentResponse)
		if err != nil {
			return err
		}
		totalBytesSent = uint64(totalBytesSentResponse.NewTotalBytesSent)
		totalBytesReceivedResponse := &wancommonifconfig.GetTotalBytesReceivedResponse{}
		err = serviceClient.GetTotalBytesReceived(totalBytesReceivedResponse)
		if err != nil {
			return err
		}
		totalBytesReceived = uint64(totalBytesReceivedResponse.NewTotalBytesReceived)
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
	}
	fields := map[string]interface{}{
		"layer1_upstream_max_bit_rate":   commonLinkProperties.NewLayer1UpstreamMaxBitRate,
		"layer1_downstream_max_bit_rate": commonLinkProperties.NewLayer1DownstreamMaxBitRate,
		"upstream_current_max_speed":     commonLinkProperties.NewX_AVM_DE_UpstreamCurrentMaxSpeed,
		"downstream_current_max_speed":   commonLinkProperties.NewX_AVM_DE_DownstreamCurrentMaxSpeed,
		"total_bytes_sent":               totalBytesSent,
		"total_bytes_received":           totalBytesReceived,
	}
	acc.AddFields("fritzbox_wan", fields, tags)
	return nil
}

func (f *Fritzbox) gatherPppInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wanpppconn.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wanpppconn.GetInfoResponse{}
	err := serviceClient.GetInfo(info)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
	}
	fields := map[string]interface{}{
		"uptime":                  info.NewUptime,
		"upstream_max_bit_rate":   info.NewUpstreamMaxBitRate,
		"downstream_max_bit_rate": info.NewDownstreamMaxBitRate,
	}
	acc.AddFields("fritzbox_ppp", fields, tags)
	return nil
}

func (f *Fritzbox) gatherDslInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wandslifconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wandslifconfig.GetInfoResponse{}
	err := serviceClient.GetInfo(info)
	if err != nil {
		return err
	}
	statisticsTotal := &wandslifconfig.GetStatisticsTotalResponse{}
	if info.NewStatus == "Up" {
		err = serviceClient.GetStatisticsTotal(statisticsTotal)
		if err != nil {
			return err
		}
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
		"status":  info.NewStatus,
	}
	fields := map[string]interface{}{
		"upstream_curr_rate":      info.NewUpstreamCurrRate,
		"downstream_curr_rate":    info.NewDownstreamCurrRate,
		"upstream_max_rate":       info.NewUpstreamMaxRate,
		"downstream_max_rate":     info.NewDownstreamMaxRate,
		"upstream_noise_margin":   info.NewUpstreamNoiseMargin,
		"downstream_noise_margin": info.NewDownstreamNoiseMargin,
		"upstream_attenuation":    info.NewUpstreamAttenuation,
		"downstream_attenuation":  info.NewDownstreamAttenuation,
		"upstream_power":          info.NewUpstreamPower,
		"downstream_power":        info.NewDownstreamPower,
		"receive_blocks":          statisticsTotal.NewReceiveBlocks,
		"transmit_blocks":         statisticsTotal.NewTransmitBlocks,
		"cell_delin":              statisticsTotal.NewCellDelin,
		"link_retrain":            statisticsTotal.NewLinkRetrain,
		"init_errors":             statisticsTotal.NewInitErrors,
		"init_timeouts":           statisticsTotal.NewInitTimeouts,
		"loss_of_framing":         statisticsTotal.NewLossOfFraming,
		"errored_secs":            statisticsTotal.NewErroredSecs,
		"severly_errored_secs":    statisticsTotal.NewSeverelyErroredSecs,
		"fec_errors":              statisticsTotal.NewFECErrors,
		"atuc_fec_errors":         statisticsTotal.NewATUCFECErrors,
		"hec_errors":              statisticsTotal.NewHECErrors,
		"atuc_hec_errors":         statisticsTotal.NewATUCHECErrors,
		"crc_errors":              statisticsTotal.NewCRCErrors,
		"atuc_crc_errors":         statisticsTotal.NewATUCCRCErrors,
	}
	acc.AddFields("fritzbox_dsl", fields, tags)
	return nil
}

func (f *Fritzbox) gatherWlanInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wlanconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wlanconfig.GetInfoResponse{}
	err := serviceClient.GetInfo(info)
	if err != nil {
		return err
	}
	totalAssociations := &wlanconfig.GetTotalAssociationsResponse{}
	if info.NewStatus == "Up" {
		err = serviceClient.GetTotalAssociations(totalAssociations)
		if err != nil {
			return err
		}
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
		"status":  info.NewStatus,
		"ssid":    info.NewSSID,
		"channel": strconv.Itoa(int(info.NewChannel)),
		"band":    f.wlanBandFromInfo(info),
	}
	fields := map[string]interface{}{
		"total_associations": totalAssociations.NewTotalAssociations,
	}
	acc.AddGauge("fritzbox_wlan", fields, tags)
	return nil
}

func (f *Fritzbox) wlanBandFromInfo(info *wlanconfig.GetInfoResponse) string {
	band := info.NewX_AVM_DE_FrequencyBand
	if band != "" {
		return band
	}
	if 1 <= info.NewChannel && info.NewChannel <= 14 {
		return "2400"
	}
	return "5000"
}

func (f *Fritzbox) gatherHostsInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := hosts.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	connections, err := f.fetchHostsConnections(&serviceClient)
	if err != nil {
		return err
	}
	for _, connection := range connections {
		// Ignore ephemeral UUID style device names
		_, err = uuid.Parse(connection.RightDeviceName)
		if err == nil {
			continue
		}
		tags := map[string]string{
			"source":       serviceClient.TR064Client.DeviceUrl.Hostname(),
			"service":      serviceClient.Service.ShortId(),
			"node":         connection.RightDeviceName,
			"node_role":    f.hostRole(connection.RightMeshRole),
			"node_ap":      connection.LeftDeviceName,
			"node_ap_role": f.hostRole(connection.LeftMeshRole),
			"link_type":    connection.InterfaceType,
			"link_name":    connection.InterfaceName,
		}
		fields := map[string]interface{}{
			"max_data_rate_tx": connection.MaxDataRateTx,
			"max_data_rate_rx": connection.MaxDataRateRx,
			"cur_data_rate_tx": connection.CurDataRateTx,
			"cur_data_rate_rx": connection.CurDataRateRx,
		}
		acc.AddGauge("fritzbox_hosts", fields, tags)
	}
	return nil
}

func (f *Fritzbox) hostRole(role string) string {
	if role == "unknown" {
		return "client"
	}
	return role
}

func (f *Fritzbox) fetchHostsConnections(serviceClient *hosts.ServiceClient) ([]*mesh.Connection, error) {
	meshListPath := &hosts.X_AVM_DE_GetMeshListPathResponse{}
	err := serviceClient.X_AVM_DE_GetMeshListPath(meshListPath)
	if err != nil {
		return nil, err
	}
	meshListResponse, err := serviceClient.TR064Client.Get(meshListPath.NewX_AVM_DE_MeshListPath)
	if err != nil {
		return nil, err
	}
	if meshListResponse.StatusCode == http.StatusNotFound {
		return make([]*mesh.Connection, 0), nil
	}
	if meshListResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch mesh list %q (status: %s)", meshListPath.NewX_AVM_DE_MeshListPath, meshListResponse.Status)
	}
	defer meshListResponse.Body.Close()
	meshListBytes, err := io.ReadAll(meshListResponse.Body)
	if err != nil {
		return nil, err
	}
	meshList := &mesh.List{}
	err = json.Unmarshal(meshListBytes, meshList)
	if err != nil {
		return nil, err
	}
	return meshList.Connections(), nil
}

func init() {
	inputs.Add("fritzbox", func() telegraf.Input {
		return &Fritzbox{Timeout: config.Duration(10 * time.Second)}
	})
}
