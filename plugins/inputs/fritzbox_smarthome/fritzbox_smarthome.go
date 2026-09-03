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

	clients []*fritzsmarthome.Client
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
	f.clients = make([]*fritzsmarthome.Client, 0, len(f.URLs))
	for _, rawURL := range f.URLs {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("parsing device URL %q failed: %w", rawURL, err)
		}
		client, err := fritzsmarthome.NewClient(parsedURL, fritzsmarthome.WithHttpClient(httpClient))
		if err != nil {
			return fmt.Errorf("creating smarthome client for URL %q failed: %w", rawURL, err)
		}
		f.clients = append(f.clients, client)
	}

	return nil
}

func (f *FritzboxSmarthome) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, client := range f.clients {
		wg.Add(1)
		// Pass client as parameter to avoid any race conditions
		go func(client *fritzsmarthome.Client) {
			defer wg.Done()
			acc.AddError(gatherClient(context.Background(), acc, client))
		}(client)
	}
	wg.Wait()
	return nil
}

func gatherClient(ctx context.Context, acc telegraf.Accumulator, client *fritzsmarthome.Client) error {
	response, err := client.GetOverview(ctx)
	if err != nil {
		return err
	}
	// err == nil always implies http.StatusOK
	if response.JSON200 == nil {
		return fmt.Errorf("empty or non-JSON response received from URL %q", client.BaseURL().String())
	}
	for _, device := range response.JSON200.Devices {
		gatherDeviceInfo(acc, client, &device)
	}
	for _, unit := range response.JSON200.Units {
		gatherUnitInfo(acc, client, &unit, (*response.JSON200).Groups)
	}
	return nil
}

func gatherDeviceInfo(acc telegraf.Accumulator, client *fritzsmarthome.Client, device *api.EndpointOverviewMultipleDevices) {
	tags := map[string]string{
		"source":           client.BaseURL().Hostname(),
		"manufacturer":     device.Manufacturer,
		"product_category": string(device.ProductCategory),
		"power_source":     getDevicePowerSource(device),
	}
	fields := map[string]interface{}{
		"name":             device.Name,
		"product_name":     device.ProductName,
		"connected":        device.IsConnected,
		"battery_value":    getOptionalInt(device.BatteryValue),
		"battery_low":      getOptionalBool(device.IsBatteryLow),
		"update_available": getOptionalBool(device.IsUpdateAvailable),
	}
	acc.AddFields("fritzbox_smarthome_device", fields, tags)
}

func gatherUnitInfo(acc telegraf.Accumulator, client *fritzsmarthome.Client, unit *api.EndpointOverviewMultipleUnits, groups []api.EndpointOverviewGroup) {
	tags := map[string]string{
		"source": client.BaseURL().Hostname(),
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
	fields := map[string]interface{}{
		"name":  unit.Name,
		"level": getOptionalInt(unit.Interfaces.LevelControlInterface.Level),
	}
	acc.AddFields("fritzbox_smarthome_level_control", fields, tags)
}

func gatherUnitInfoMultimeter(acc telegraf.Accumulator, tags map[string]string, unit *api.EndpointOverviewMultipleUnits) {
	if unit.Interfaces.MultimeterInterface == nil {
		return
	}
	fields := map[string]interface{}{
		"name":    unit.Name,
		"current": getOptionalInt(unit.Interfaces.MultimeterInterface.Current),
		"energy":  getOptionalInt(unit.Interfaces.MultimeterInterface.Energy),
		"power":   getOptionalInt(unit.Interfaces.MultimeterInterface.Power),
		"voltage": getOptionalInt(unit.Interfaces.MultimeterInterface.Voltage),
	}
	acc.AddFields("fritzbox_smarthome_multimeter", fields, tags)
}

func gatherUnitInfoOnOff(acc telegraf.Accumulator, tags map[string]string, unit *api.EndpointOverviewMultipleUnits) {
	if unit.Interfaces.OnOffInterface == nil {
		return
	}
	fields := map[string]interface{}{
		"name":   unit.Name,
		"active": getOptionalBool(unit.Interfaces.OnOffInterface.Active),
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

func getUnitGroupName(unit *api.EndpointOverviewMultipleUnits, groups []api.EndpointOverviewGroup) string {
	groupName := "<none>"
	if unit.GroupUid != nil {
		for _, group := range groups {
			if unit.GroupUid != nil && group.Name != nil && group.UID == *unit.GroupUid {
				groupName = *group.Name
				break
			}
		}
	}
	return groupName
}

func getOptionalBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func getOptionalInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func init() {
	inputs.Add("fritzbox_smarthome", func() telegraf.Input {
		return &FritzboxSmarthome{Timeout: config.Duration(10 * time.Second)}
	})
}
