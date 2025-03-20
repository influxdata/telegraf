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
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type serviceHandlerFunc func(telegraf.Accumulator, *tr064.Client, tr064.ServiceDescriptor) error

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
	// No need to run without any devices configured
	if len(f.URLs) == 0 {
		return errors.New("no device URLs configured")
	}

	// Use default collect options if nothing is configured
	if len(f.Collect) == 0 {
		f.Collect = []string{"device", "wan", "ppp", "dsl", "wlan"}
	}

	// Setup TLS
	tlsConfig, err := f.TLSConfig()
	if err != nil {
		return fmt.Errorf("initializing TLS configuration failed: %w", err)
	}

	// Initialize the device clients
	debug := f.Log.Level().Includes(telegraf.Trace)
	f.deviceClients = make([]*tr064.Client, 0, len(f.URLs))
	for _, rawURL := range f.URLs {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("parsing device URL %q failed: %w", rawURL, err)
		}
		client := tr064.NewClient(parsedURL)
		client.Debug = debug
		client.Timeout = time.Duration(f.Timeout)
		client.TlsConfig = tlsConfig
		f.deviceClients = append(f.deviceClients, client)
	}

	// Initialize the service handlers
	f.serviceHandlers = make(map[string]serviceHandlerFunc, len(f.Collect))
	for _, c := range f.Collect {
		switch c {
		case "device":
			f.serviceHandlers[deviceinfo.ServiceShortType] = gatherDeviceInfo
		case "wan":
			f.serviceHandlers[wancommonifconfig.ServiceShortType] = gatherWanInfo
		case "ppp":
			f.serviceHandlers[wanpppconn.ServiceShortType] = gatherPppInfo
		case "dsl":
			f.serviceHandlers[wandslifconfig.ServiceShortType] = gatherDslInfo
		case "wlan":
			f.serviceHandlers[wlanconfig.ServiceShortType] = gatherWlanInfo
		case "hosts":
			f.serviceHandlers[hosts.ServiceShortType] = gatherHostsInfo
		default:
			return fmt.Errorf("invalid service %q in collect parameter", c)
		}
	}

	return nil
}

func (f *Fritzbox) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, deviceClient := range f.deviceClients {
		wg.Add(1)
		// Pass deviceClient as parameter to avoid any race conditions
		go func(client *tr064.Client) {
			defer wg.Done()
			f.gatherDevice(acc, client)
		}(deviceClient)
	}
	wg.Wait()
	return nil
}

func (f *Fritzbox) gatherDevice(acc telegraf.Accumulator, deviceClient *tr064.Client) {
	services, err := deviceClient.Services(tr064.DefaultServiceSpec)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, service := range services {
		serviceHandler, exists := f.serviceHandlers[service.ShortType()]
		// If no serviceHandler has been setup during Init(), we ignore this service.
		if !exists {
			continue
		}
		acc.AddError(serviceHandler(acc, deviceClient, service))
	}
}

func gatherDeviceInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := deviceinfo.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &deviceinfo.GetInfoResponse{}
	if err := serviceClient.GetInfo(info); err != nil {
		return fmt.Errorf("failed to query device info: %w", err)
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

func gatherWanInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wancommonifconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	commonLinkProperties := &wancommonifconfig.GetCommonLinkPropertiesResponse{}
	if err := serviceClient.GetCommonLinkProperties(commonLinkProperties); err != nil {
		return fmt.Errorf("failed to query link properties: %w", err)
	}
	// Prefer igdicfg service over wancommonifconfig service for total bytes stats, because igdicfg supports uint64 counters
	igdServices, err := deviceClient.ServicesByType(tr064.IgdServiceSpec, igdicfg.ServiceShortType)
	if err != nil {
		return fmt.Errorf("failed to lookup IGD service: %w", err)
	}
	var totalBytesSent uint64
	var totalBytesReceived uint64
	if len(igdServices) > 0 {
		igdServiceClient := &igdicfg.ServiceClient{
			TR064Client: deviceClient,
			Service:     igdServices[0],
		}
		addonInfos := &igdicfg.GetAddonInfosResponse{}
		if err = igdServiceClient.GetAddonInfos(addonInfos); err != nil {
			return fmt.Errorf("failed to query addon info: %w", err)
		}
		totalBytesSent, err = strconv.ParseUint(addonInfos.NewX_AVM_DE_TotalBytesSent64, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse total bytes sent: %w", err)
		}
		totalBytesReceived, err = strconv.ParseUint(addonInfos.NewX_AVM_DE_TotalBytesReceived64, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse total bytes received: %w", err)
		}
	} else {
		// Fall back to wancommonifconfig service in case igdicfg is not available (only uint32 based)
		totalBytesSentResponse := &wancommonifconfig.GetTotalBytesSentResponse{}
		if err = serviceClient.GetTotalBytesSent(totalBytesSentResponse); err != nil {
			return fmt.Errorf("failed to query bytes sent: %w", err)
		}
		totalBytesSent = uint64(totalBytesSentResponse.NewTotalBytesSent)
		totalBytesReceivedResponse := &wancommonifconfig.GetTotalBytesReceivedResponse{}
		if err = serviceClient.GetTotalBytesReceived(totalBytesReceivedResponse); err != nil {
			return fmt.Errorf("failed to query bytes received: %w", err)
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

func gatherPppInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wanpppconn.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wanpppconn.GetInfoResponse{}
	if err := serviceClient.GetInfo(info); err != nil {
		return fmt.Errorf("failed to query PPP info: %w", err)
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

func gatherDslInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wandslifconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wandslifconfig.GetInfoResponse{}
	if err := serviceClient.GetInfo(info); err != nil {
		return fmt.Errorf("failed to query DSL info: %w", err)
	}
	statisticsTotal := &wandslifconfig.GetStatisticsTotalResponse{}
	if info.NewStatus == "Up" {
		if err := serviceClient.GetStatisticsTotal(statisticsTotal); err != nil {
			return fmt.Errorf("failed to query DSL statistics: %w", err)
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

func gatherWlanInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := wlanconfig.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	info := &wlanconfig.GetInfoResponse{}
	if err := serviceClient.GetInfo(info); err != nil {
		return fmt.Errorf("failed to query WLAN info: %w", err)
	}
	totalAssociations := &wlanconfig.GetTotalAssociationsResponse{}
	if info.NewStatus == "Up" {
		if err := serviceClient.GetTotalAssociations(totalAssociations); err != nil {
			return fmt.Errorf("failed to query WLAN associations: %w", err)
		}
	}
	tags := map[string]string{
		"source":  serviceClient.TR064Client.DeviceUrl.Hostname(),
		"service": serviceClient.Service.ShortId(),
		"status":  info.NewStatus,
		"ssid":    info.NewSSID,
		"channel": strconv.Itoa(int(info.NewChannel)),
		"band":    wlanBandFromInfo(info),
	}
	fields := map[string]interface{}{
		"total_associations": totalAssociations.NewTotalAssociations,
	}
	acc.AddGauge("fritzbox_wlan", fields, tags)
	return nil
}

func wlanBandFromInfo(info *wlanconfig.GetInfoResponse) string {
	band := info.NewX_AVM_DE_FrequencyBand
	if band != "" {
		return band
	}
	if 1 <= info.NewChannel && info.NewChannel <= 14 {
		return "2400"
	}
	return "5000"
}

func gatherHostsInfo(acc telegraf.Accumulator, deviceClient *tr064.Client, service tr064.ServiceDescriptor) error {
	serviceClient := hosts.ServiceClient{
		TR064Client: deviceClient,
		Service:     service,
	}
	connections, err := fetchHostsConnections(&serviceClient)
	if err != nil {
		return fmt.Errorf("failed to fetch hosts connections: %w", err)
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
			"node_role":    hostRole(connection.RightMeshRole),
			"node_ap":      connection.LeftDeviceName,
			"node_ap_role": hostRole(connection.LeftMeshRole),
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

func hostRole(role string) string {
	if role == "unknown" {
		return "client"
	}
	return role
}

func fetchHostsConnections(serviceClient *hosts.ServiceClient) ([]*mesh.Connection, error) {
	meshListPath := &hosts.X_AVM_DE_GetMeshListPathResponse{}
	if err := serviceClient.X_AVM_DE_GetMeshListPath(meshListPath); err != nil {
		return nil, fmt.Errorf("failed to query mesh list path: %w", err)
	}
	meshListResponse, err := serviceClient.TR064Client.Get(meshListPath.NewX_AVM_DE_MeshListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access mesh list %q: %w", meshListPath.NewX_AVM_DE_MeshListPath, err)
	}
	if meshListResponse.StatusCode == http.StatusNotFound {
		return make([]*mesh.Connection, 0), nil
	}
	if meshListResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch mesh list %q: %s", meshListPath.NewX_AVM_DE_MeshListPath, meshListResponse.Status)
	}
	defer meshListResponse.Body.Close()
	meshListBytes, err := io.ReadAll(meshListResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read mesh list: %w", err)
	}
	meshList := &mesh.List{}
	if err := json.Unmarshal(meshListBytes, meshList); err != nil {
		return nil, fmt.Errorf("failed to parse mesh list: %w", err)
	}
	return meshList.Connections(), nil
}

func init() {
	inputs.Add("fritzbox", func() telegraf.Input {
		return &Fritzbox{Timeout: config.Duration(10 * time.Second)}
	})
}
