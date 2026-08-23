//go:generate ../../../tools/readme_config_includer/generator
package fritzbox_smarthome

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tdrn-org/go-fritzsmarthome"
	"github.com/tdrn-org/go-fritzsmarthome/api"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type FritzboxSmarthome struct {
	URLs    []string        `toml:"urls"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`
	tls.ClientConfig

	smarthomeClients []*fritzsmarthome.Client
}

func (*FritzboxSmarthome) SampleConfig() string {
	return sampleConfig
}

func (f *FritzboxSmarthome) Init() error {
	// No need to run without any devices configured
	if len(f.URLs) == 0 {
		return errors.New("no device URLs configured")
	}

	// Setup http.Client
	tlsConfig, err := f.TLSConfig()
	if err != nil {
		return fmt.Errorf("initializing TLS configuration failed: %w", err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(f.Timeout),
	}

	// Initialize the smarthome clients
	f.smarthomeClients = make([]*fritzsmarthome.Client, 0, len(f.URLs))
	for _, rawURL := range f.URLs {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("parsing device URL %q failed: %w", rawURL, err)
		}
		client, err := fritzsmarthome.NewClient(parsedURL, fritzsmarthome.WithHttpClient(httpClient))
		if err != nil {
			return fmt.Errorf("creating smarthome for URL %q failed: %w", rawURL, err)
		}
		f.smarthomeClients = append(f.smarthomeClients, client)
	}

	return nil
}

func (f *FritzboxSmarthome) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, smarthomeClient := range f.smarthomeClients {
		wg.Add(1)
		// Pass smarthomeClient as parameter to avoid any race conditions
		go func(client *fritzsmarthome.Client) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(f.Timeout))
			defer cancel()
			f.gatherClient(ctx, acc, client)
		}(smarthomeClient)
	}
	wg.Wait()
	return nil
}

func (f *FritzboxSmarthome) gatherClient(ctx context.Context, acc telegraf.Accumulator, smarthomeClient *fritzsmarthome.Client) {
	response, err := smarthomeClient.GetOverview(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	if response.HTTPResponse.StatusCode != http.StatusOK {
		acc.AddError(fmt.Errorf("requesting Smarthome status failed: %d %s", response.HTTPResponse.StatusCode, response.HTTPResponse.Status))
		return
	}
	for _, device := range (*response.JSON200).Devices {
		gatherDeviceInfo(acc, smarthomeClient, &device)
	}
	for _, unit := range (*response.JSON200).Units {
		gatherUnitInfo(acc, smarthomeClient, &unit, (*response.JSON200).Groups)
	}
}

func gatherDeviceInfo(acc telegraf.Accumulator, smarthomeClient *fritzsmarthome.Client, device *api.EndpointOverviewMultipleDevices) {
	tags := map[string]string{
		"source":           smarthomeClient.BaseURL().Hostname(),
		"manufacturer":     device.Manufacturer,
		"product_category": string(device.ProductCategory),
		"power_source":     getDevicePowerSource(device),
	}
	connected := 0
	if device.IsConnected {
		connected = 1
	}
	batteryValue, batteryLow := getDeviceBatteryValue(device)
	updateAvailable := 0
	if device.IsUpdateAvailable != nil && *device.IsUpdateAvailable {
		updateAvailable = 1
	}
	fields := map[string]interface{}{
		"name":             device.Name,
		"product_name":     device.ProductName,
		"connected":        connected,
		"battery_value":    batteryValue,
		"battery_low":      batteryLow,
		"update_available": updateAvailable,
	}
	acc.AddFields("fritzbox_smarthome_device", fields, tags)
}

func gatherUnitInfo(acc telegraf.Accumulator, smarthomeClient *fritzsmarthome.Client, unit *api.EndpointOverviewMultipleUnits, groups []api.EndpointOverviewGroup) {
	tags := map[string]string{
		"source": smarthomeClient.BaseURL().Hostname(),
		"type":   string(unit.UnitType),
		"group":  getUnitGroupName(unit, groups),
	}
	gatherUnitInfoLevelControl(acc, tags, unit)
	gatherUnitInfoMultimeter(acc, tags, unit)
	gatherUnitInfoOnOff(acc, tags, unit)
}

func gatherUnitInfoLevelControl(acc telegraf.Accumulator, tags map[string]string, unit *api.EndpointOverviewMultipleUnits) {
	if unit.Interfaces.LevelControlInterface == nil {
		return
	}
	level := 0
	if unit.Interfaces.LevelControlInterface.Level != nil {
		level = *unit.Interfaces.LevelControlInterface.Level
	}
	fields := map[string]interface{}{
		"name":  unit.Name,
		"level": level,
	}
	acc.AddFields("fritzbox_smarthome_level_control", fields, tags)
}

func gatherUnitInfoMultimeter(acc telegraf.Accumulator, tags map[string]string, unit *api.EndpointOverviewMultipleUnits) {
	if unit.Interfaces.MultimeterInterface == nil {
		return
	}
	current := 0
	if unit.Interfaces.MultimeterInterface.Current != nil {
		current = *unit.Interfaces.MultimeterInterface.Current
	}
	energy := 0
	if unit.Interfaces.MultimeterInterface.Energy != nil {
		energy = *unit.Interfaces.MultimeterInterface.Energy
	}
	power := 0
	if unit.Interfaces.MultimeterInterface.Power != nil {
		power = *unit.Interfaces.MultimeterInterface.Power
	}
	voltage := 0
	if unit.Interfaces.MultimeterInterface.Voltage != nil {
		voltage = *unit.Interfaces.MultimeterInterface.Voltage
	}
	fields := map[string]interface{}{
		"name":    unit.Name,
		"current": current,
		"energy":  energy,
		"power":   power,
		"voltage": voltage,
	}
	acc.AddFields("fritzbox_smarthome_multimeter", fields, tags)
}

func gatherUnitInfoOnOff(acc telegraf.Accumulator, tags map[string]string, unit *api.EndpointOverviewMultipleUnits) {
	if unit.Interfaces.OnOffInterface == nil {
		return
	}
	active := false
	if unit.Interfaces.OnOffInterface.Active != nil {
		active = *unit.Interfaces.OnOffInterface.Active
	}
	fields := map[string]interface{}{
		"name":   unit.Name,
		"active": active,
	}
	acc.AddFields("fritzbox_smarthome_on_off", fields, tags)
}

func getDevicePowerSource(device *api.EndpointOverviewMultipleDevices) string {
	powerSource := "internal"
	if device.IsBatteryPowered != nil && *device.IsBatteryPowered {
		powerSource = "battery"
	} else if device.IsExternallyPowered != nil && *device.IsExternallyPowered {
		powerSource = "external"
	}
	return powerSource
}

func getDeviceBatteryValue(device *api.EndpointOverviewMultipleDevices) (int, int) {
	batteryValue, batteryLow := -1, 0
	if device.BatteryValue != nil {
		batteryValue = *device.BatteryValue
	}
	if device.IsBatteryLow != nil && *device.IsBatteryLow {
		batteryLow = 1
	}
	return batteryValue, batteryLow
}

func getUnitGroupName(unit *api.EndpointOverviewMultipleUnits, groups []api.EndpointOverviewGroup) string {
	groupName := "<none>"
	if unit.GroupUid != nil {
		for _, group := range groups {
			if group.UID == *unit.GroupUid {
				groupName = *group.Name
				break
			}
		}
	}
	return groupName
}

func init() {
	inputs.Add("fritzbox_smarthome", func() telegraf.Input {
		return &FritzboxSmarthome{Timeout: config.Duration(10 * time.Second)}
	})
}
