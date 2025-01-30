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

type RemoteClientConfig struct {
	RemoteClientId     string `toml:"remote_client_id"`
	RemoteClientSecret string `toml:"remote_client_secret"`
	RemoteCallbackUrl  string `toml:"remote_callback_url"`
	RemoteTokenDir     string `toml:"remote_token_dir"`
}

type HueBridge struct {
	Bridges         []string          `toml:"bridges"`
	RoomAssignments map[string]string `toml:"room_assignments"`
	Timeout         config.Duration   `toml:"timeout"`
	RemoteClientConfig
	tls.ClientConfig
	Log             telegraf.Logger `toml:"-"`
	bridgeResolvers []*bridgeResolver
}

func (*HueBridge) SampleConfig() string {
	return sampleConfig
}

func (h *HueBridge) Init() error {
	h.bridgeResolvers = make([]*bridgeResolver, 0, len(h.Bridges))
	// Front load URL parsing and warn & skip invalid URLs already during
	// initialization to prevent unnecessary log flooding during Gather calls.
	for _, bridge := range h.Bridges {
		bridgeResolver, err := newBridgeResolver(bridge)
		if err != nil {
			h.Log.Warnf("Failed to parse bridge URL %q (reason: %s)", bridge, err)
			continue
		}
		h.bridgeResolvers = append(h.bridgeResolvers, bridgeResolver)
	}
	return nil
}

func (h *HueBridge) Gather(acc telegraf.Accumulator) error {
	var waitComplete sync.WaitGroup
	for _, bridgeResolver := range h.bridgeResolvers {
		waitComplete.Add(1)
		go func() {
			defer waitComplete.Done()
			bridgeClient, err := bridgeResolver.resolveBridge(&h.RemoteClientConfig, &h.ClientConfig, h.Timeout)
			if err != nil {
				h.Log.Warnf("Failed to resolve bridge %q (reason %s)", bridgeResolver, err)
				return
			}
			err = h.processBridge(acc, bridgeClient)
			if err != nil {
				bridgeResolver.reset()
				acc.AddError(err)
			}
		}()
	}
	waitComplete.Wait()
	return nil
}

func (h *HueBridge) processBridge(acc telegraf.Accumulator, bridgeClient hue.BridgeClient) error {
	h.Log.Debugf("Processing bridge %s", bridgeClient.Bridge().BridgeId)
	metadata, err := fetchMetadata(bridgeClient, h.RoomAssignments)
	if err != nil {
		return err
	}
	err = h.processLights(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = h.processTemperatures(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = h.processLightLevels(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = h.processMotionSensors(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	err = h.processDevicePowers(acc, bridgeClient, metadata)
	if err != nil {
		acc.AddError(err)
	}
	return nil
}

func (h *HueBridge) processLights(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *bridgeMetadata) error {
	getLightsResponse, err := bridgeClient.GetLights()
	if err != nil {
		return fmt.Errorf("failed to access bridge lights on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getLightsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge lights from %q (status: %s)", bridgeClient.Url().Redacted(), getLightsResponse.HTTPResponse.Status)
	}
	responseData := (*getLightsResponse.JSON200).Data
	if responseData != nil {
		for _, light := range *responseData {
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.resolveResourceRoom(*light.Id, *light.Metadata.Name)
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

func (h *HueBridge) processTemperatures(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *bridgeMetadata) error {
	getTemperaturesResponse, err := bridgeClient.GetTemperatures()
	if err != nil {
		return fmt.Errorf("failed to access bridge temperatures on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getTemperaturesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge temperatures from %q (status: %s)", bridgeClient.Url().Redacted(), getTemperaturesResponse.HTTPResponse.Status)
	}
	responseData := (*getTemperaturesResponse.JSON200).Data
	if responseData != nil {
		for _, temperature := range *responseData {
			temperatureName := metadata.resolveDeviceName(*temperature.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.resolveResourceRoom(*temperature.Id, temperatureName)
			tags["device"] = temperatureName
			tags["enabled"] = strconv.FormatBool(*temperature.Enabled)
			fields := make(map[string]interface{})
			fields["temperature"] = *temperature.Temperature.TemperatureReport.Temperature
			acc.AddGauge("huebridge_temperature", fields, tags)
		}
	}
	return nil
}

func (h *HueBridge) processLightLevels(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *bridgeMetadata) error {
	getLightLevelsResponse, err := bridgeClient.GetLightLevels()
	if err != nil {
		return fmt.Errorf("failed to access bridge lights levels on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getLightLevelsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge light levels from %q (status: %s)", bridgeClient.Url().Redacted(), getLightLevelsResponse.HTTPResponse.Status)
	}
	responseData := (*getLightLevelsResponse.JSON200).Data
	if responseData != nil {
		for _, lightLevel := range *responseData {
			lightLevelName := metadata.resolveDeviceName(*lightLevel.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.resolveResourceRoom(*lightLevel.Id, lightLevelName)
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

func (h *HueBridge) processMotionSensors(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *bridgeMetadata) error {
	getMotionSensorsResponse, err := bridgeClient.GetMotionSensors()
	if err != nil {
		return fmt.Errorf("failed to access bridge motion sensors on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getMotionSensorsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge motion sensors from %q (status: %s)", bridgeClient.Url().Redacted(), getMotionSensorsResponse.HTTPResponse.Status)
	}
	responseData := (*getMotionSensorsResponse.JSON200).Data
	if responseData != nil {
		for _, motionSensor := range *responseData {
			motionSensorName := metadata.resolveDeviceName(*motionSensor.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.resolveResourceRoom(*motionSensor.Id, motionSensorName)
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

func (h *HueBridge) processDevicePowers(acc telegraf.Accumulator, bridgeClient hue.BridgeClient, metadata *bridgeMetadata) error {
	getDevicePowersResponse, err := bridgeClient.GetDevicePowers()
	if err != nil {
		return fmt.Errorf("failed to access bridge device powers on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getDevicePowersResponse.HTTPResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch bridge device powers from %q (status: %s)", bridgeClient.Url().Redacted(), getDevicePowersResponse.HTTPResponse.Status)
	}
	responseData := (*getDevicePowersResponse.JSON200).Data
	if responseData != nil {
		for _, devicePower := range *responseData {
			if devicePower.PowerState.BatteryLevel == nil && devicePower.PowerState.BatteryState == nil {
				continue
			}
			devicePowerName := metadata.resolveDeviceName(*devicePower.Id)
			tags := make(map[string]string)
			tags["bridge_id"] = bridgeClient.Bridge().BridgeId
			tags["room"] = metadata.resolveResourceRoom(*devicePower.Id, devicePowerName)
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
		return &HueBridge{Timeout: config.Duration(10 * time.Second)}
	})
}
