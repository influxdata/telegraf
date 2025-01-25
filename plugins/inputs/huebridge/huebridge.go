//go:generate ../../../tools/readme_config_includer/generator
package huebridge

import (
	_ "embed"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tdrn-org/go-hue"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type BridgeClientConfig struct {
	RemoteClientId     string          `toml:"remote_client_id"`
	RemoteClientSecret string          `toml:"remote_client_secret"`
	RemoteCallbackUrl  string          `toml:"remote_callback_url"`
	RemoteTokenDir     string          `toml:"remote_token_dir"`
	Timeout            config.Duration `toml:"timeout"`
	tls.ClientConfig
}

type HueBridge struct {
	Bridges []string `toml:"bridges"`
	BridgeClientConfig
	RoomAssignments map[string]string `toml:"room_assignments"`
	Log             telegraf.Logger   `toml:"-"`
	resolvedBridges map[*BridgeURL]hue.BridgeClient
}

func defaultHueBridge() *HueBridge {
	return &HueBridge{
		Bridges:         make([]string, 0),
		RoomAssignments: make(map[string]string),
		BridgeClientConfig: BridgeClientConfig{
			Timeout: config.Duration(10 * time.Second),
		},
		resolvedBridges: make(map[*BridgeURL]hue.BridgeClient),
	}
}

func (*HueBridge) SampleConfig() string {
	return sampleConfig
}

func (plugin *HueBridge) Init() error {
	// Front load URL parsing and warn & skip invalid URLs already during
	// initialization to prevent unnecessary log flooding during Gather calls.
	for _, bridge := range plugin.Bridges {
		bridgeUrl, err := ParseBridgeURL(bridge)
		if err != nil {
			plugin.Log.Warnf("Failed to parse bridge URL %q (reason: %s)", bridgeUrl, err)
			continue
		}
		// Collect the valid URLs with a nil client; the latter will be re-resolved during
		// Gather call.
		plugin.resolvedBridges[bridgeUrl] = nil
	}
	return nil
}

func (plugin *HueBridge) Gather(acc telegraf.Accumulator) error {
	var waitComplete sync.WaitGroup
	reResolvedBridges := make(chan struct {
		*BridgeURL
		hue.BridgeClient
	})
	for bridgeUrl, bridgeClient := range plugin.resolvedBridges {
		waitComplete.Add(1)
		go func() {
			defer waitComplete.Done()
			resolvedBridgeClient := bridgeClient
			if resolvedBridgeClient == nil {
				plugin.Log.Infof("Re-resolving bridge %q...", bridgeUrl)
				reResolvedBridgeClient, err := bridgeUrl.ResolveBridge(&plugin.BridgeClientConfig)
				if err != nil {
					plugin.Log.Warnf("Failed to resolve bridge %q (reason %s)", bridgeUrl, err)
					return
				}
				resolvedBridgeClient = reResolvedBridgeClient
			}
			err := plugin.processBridge(acc, resolvedBridgeClient)
			if err != nil && bridgeClient != nil {
				// Previously resolved client failed; discard it and re-resolve on next run
				reResolvedBridges <- struct {
					*BridgeURL
					hue.BridgeClient
				}{bridgeUrl, nil}
				acc.AddError(err)
			} else if bridgeClient == nil {
				// Bridge client successfully re-resolved; re-use it on following runs
				reResolvedBridges <- struct {
					*BridgeURL
					hue.BridgeClient
				}{bridgeUrl, resolvedBridgeClient}
			}
		}()
	}
	go func() {
		waitComplete.Wait()
		close(reResolvedBridges)
	}()
	for reResolvedBridge := range reResolvedBridges {
		plugin.resolvedBridges[reResolvedBridge.BridgeURL] = reResolvedBridge.BridgeClient
	}
	return nil
}

func (plugin *HueBridge) processBridge(acc telegraf.Accumulator, bridgeClient hue.BridgeClient) error {
	plugin.Log.Debugf("Processing bridge %s", bridgeClient.Bridge().BridgeId)
	metadata, err := FetchMetadata(bridgeClient, plugin.RoomAssignments)
	if err != nil {
		return err
	}
	err = plugin.processLights(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = plugin.processTemperatures(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = plugin.processLightLevels(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = plugin.processMotionSensors(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = plugin.processDevicePowers(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	return nil
}

func (plugin *HueBridge) processLights(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *BridgeMetadata) error {
	getLightsResponse, err := bridgeClient.GetLights()
	if err != nil {
		return err
	}
	if getLightsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch lights (status: %s)", getLightsResponse.HTTPResponse.Status)
	}
	responseData := (*getLightsResponse.JSON200).Data
	if responseData != nil {
		for _, light := range *responseData {
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.ResolveResourceRoom(*light.Id, *light.Metadata.Name)
			tags["device"] = *light.Metadata.Name
			fields := make(map[string]interface{})
			if *light.On.On {
				fields["on"] = 1
			} else {
				fields["on"] = 0
			}
			acc.AddGauge("huebridge_light", fields, tags)
		}
	}
	return nil
}

func (plugin *HueBridge) processTemperatures(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *BridgeMetadata) error {
	getTemperaturesResponse, err := bridgeClient.GetTemperatures()
	if err != nil {
		return err
	}
	if getTemperaturesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge temperatures (status: %s)", getTemperaturesResponse.HTTPResponse.Status)
	}
	responseData := (*getTemperaturesResponse.JSON200).Data
	if responseData != nil {
		for _, temperature := range *responseData {
			temperatureName := metadata.ResolveDeviceName(*temperature.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.ResolveResourceRoom(*temperature.Id, temperatureName)
			tags["device"] = temperatureName
			tags["enabled"] = strconv.FormatBool(*temperature.Enabled)
			fields := make(map[string]interface{})
			fields["temperature"] = *temperature.Temperature.TemperatureReport.Temperature
			acc.AddGauge("huebridge_temperature", fields, tags)
		}
	}
	return nil
}

func (plugin *HueBridge) processLightLevels(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *BridgeMetadata) error {
	getLightLevelsResponse, err := bridgeClient.GetLightLevels()
	if err != nil {
		return err
	}
	if getLightLevelsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch light levels (status: %s)", getLightLevelsResponse.HTTPResponse.Status)
	}
	responseData := (*getLightLevelsResponse.JSON200).Data
	if responseData != nil {
		for _, lightLevel := range *responseData {
			lightLevelName := metadata.ResolveDeviceName(*lightLevel.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.ResolveResourceRoom(*lightLevel.Id, lightLevelName)
			tags["device"] = lightLevelName
			tags["enabled"] = strconv.FormatBool(*lightLevel.Enabled)
			fields := make(map[string]interface{})
			fields["light_level"] = *lightLevel.Light.LightLevelReport.LightLevel
			fields["light_level_lux"] = math.Pow(10.0, (float64(*lightLevel.Light.LightLevelReport.LightLevel)-1.0)/10000.0)
			acc.AddGauge("huebridge_light_level", fields, tags)
		}
	}
	return nil
}

func (plugin *HueBridge) processMotionSensors(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *BridgeMetadata) error {
	getMotionSensorsResponse, err := bridgeClient.GetMotionSensors()
	if err != nil {
		return err
	}
	if getMotionSensorsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch motion sensors (status: %s)", getMotionSensorsResponse.HTTPResponse.Status)
	}
	responseData := (*getMotionSensorsResponse.JSON200).Data
	if responseData != nil {
		for _, motionSensor := range *responseData {
			motionSensorName := metadata.ResolveDeviceName(*motionSensor.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.ResolveResourceRoom(*motionSensor.Id, motionSensorName)
			tags["device"] = motionSensorName
			tags["enabled"] = strconv.FormatBool(*motionSensor.Enabled)
			fields := make(map[string]interface{})
			if *motionSensor.Motion.MotionReport.Motion {
				fields["motion"] = 1
			} else {
				fields["motion"] = 0
			}
			acc.AddGauge("huebridge_motion_sensor", fields, tags)
		}
	}
	return nil
}

func (plugin *HueBridge) processDevicePowers(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *BridgeMetadata) error {
	getDevicePowersResponse, err := bridgeClient.GetDevicePowers()
	if err != nil {
		return err
	}
	if getDevicePowersResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch motion sensors (status: %s)", getDevicePowersResponse.HTTPResponse.Status)
	}
	responseData := (*getDevicePowersResponse.JSON200).Data
	if responseData != nil {
		for _, devicePower := range *responseData {
			if devicePower.PowerState.BatteryLevel == nil && devicePower.PowerState.BatteryState == nil {
				continue
			}
			devicePowerName := metadata.ResolveDeviceName(*devicePower.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.ResolveResourceRoom(*devicePower.Id, devicePowerName)
			tags["device"] = devicePowerName
			fields := make(map[string]interface{})
			fields["battery_level"] = *devicePower.PowerState.BatteryLevel
			fields["battery_state"] = *devicePower.PowerState.BatteryState
			acc.AddGauge("huebridge_device_power", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("huebridge", func() telegraf.Input {
		return defaultHueBridge()
	})
}
