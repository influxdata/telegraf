//go:generate ../../../tools/readme_config_includer/generator
package upsd

import (
	_ "embed"
	"fmt"
	"strings"

	nut "github.com/robbiet480/go.nut"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

//See: https://networkupstools.org/docs/developer-guide.chunked/index.html

const defaultAddress = "127.0.0.1"
const defaultPort = 3493

type Upsd struct {
	Server     string `toml:"server"`
	Port       int    `toml:"port"`
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	ForceFloat bool   `toml:"force_float"`

	Log telegraf.Logger `toml:"-"`
}

func (*Upsd) SampleConfig() string {
	return sampleConfig
}

func (u *Upsd) Gather(acc telegraf.Accumulator) error {
	upsList, err := u.fetchVariables(u.Server, u.Port)
	if err != nil {
		return err
	}
	for name, variables := range upsList {
		u.gatherUps(acc, name, variables)
	}
	return nil
}

func (u *Upsd) gatherUps(acc telegraf.Accumulator, name string, variables []nut.Variable) {
	metrics := make(map[string]interface{})
	for _, variable := range variables {
		name := variable.Name
		value := variable.Value
		metrics[name] = value
	}

	tags := map[string]string{
		"serial":   fmt.Sprintf("%v", metrics["device.serial"]),
		"ups_name": name,
		//"variables": variables.Status not sure if it's a good idea to provide this
		"model": fmt.Sprintf("%v", metrics["device.model"]),
	}

	// For compatibility with the apcupsd plugin's output we map the status string status into a bit-format
	status := u.mapStatus(metrics, tags)

	timeLeftS, err := internal.ToFloat64(metrics["battery.runtime"])
	if err != nil {
		u.Log.Warnf("Type for 'battery.runtime' is not supported: %v", err)
	}

	timeLeftNS, err := internal.ToInt64(timeLeftS * 1_000_000_000)
	if err != nil {
		u.Log.Warnf("Converting 'battery.runtime' to 'time_left_ns' failed: %v", err)
	}

	fields := map[string]interface{}{
		"battery_date":        metrics["battery.date"],
		"battery_mfr_date":    metrics["battery.mfr.date"],
		"battery_protection":  metrics["battery.protection"],
		"device_mfr":          metrics["device.mfr"],
		"device_model":        metrics["device.model"],
		"driver_version":      metrics["driver.version"],
		"driver_version_data": metrics["driver.version.data"],
		"driver_version_usb":  metrics["driver.version.usb"],
		"device_type":         metrics["device.type"],
		"ups_mfr":             metrics["ups.mfr"],
		"ups_model":           metrics["ups.model"],
		"ups_productid":       metrics["ups.productid"],
		"ups_test_result":     metrics["ups.test.result"],
		"ups_type":            metrics["ups.type"],
		//"ups_beeper_status":   metrics["ups.beeper.status"],
		//"ups_shutdown":        metrics["ups.shutdown"],
		"outlet_switchable": metrics["outlet.switchable"],
		"status_flags":      status,
		"ups_status":        metrics["ups.status"],

		//Compatibility with apcupsd metrics format
		"time_left_ns": timeLeftNS,
	}

	floatValues := map[string]string{
		"battery_charge_percent":   "battery.charge",
		"battery_runtime_low":      "battery.runtime.low",
		"battery_voltage":          "battery.voltage",
		"battery_capacity":         "battery.capacity",
		"input_frequency":          "input.frequency",
		"input_transfer_high":      "input.transfer.high",
		"input_transfer_low":       "input.transfer.low",
		"input_voltage":            "input.voltage",
		"input_bypass_frequency":   "input.bypass.frequency",
		"input_bypass_voltage":     "input.bypass.voltage",
		"input_frequency_nominal":  "input.frequency.nominal",
		"internal_temp":            "ups.temperature",
		"load_percent":             "ups.load",
		"nominal_battery_voltage":  "battery.voltage.nominal",
		"nominal_input_voltage":    "input.voltage.nominal",
		"nominal_power":            "ups.realpower.nominal",
		"output_voltage":           "output.voltage",
		"output_current":           "output.current",
		"output_frequency":         "output.frequency",
		"output_frequency_nominal": "output.frequency.nominal",
		"output_voltage_nominal":   "output.voltage.nominal",
		"ups_power":                "ups.power",
		"ups_power_nominal":        "ups.power.nominal",
		"ups_test_interval":        "ups.test.interval",
		"driver_version_internal":  "driver.version.internal",
		"real_power":               "ups.realpower",
		"ups_delay_shutdown":       "ups.delay.shutdown",
		"ups_delay_start":          "ups.delay.start",
	}

	for key, rawValue := range floatValues {
		if metrics[rawValue] == nil {
			continue
		}

		if !u.ForceFloat {
			fields[key] = metrics[rawValue]
			continue
		}

		// Force expected float values to actually being float (e.g. if delivered as int)
		float, err := internal.ToFloat64(metrics[rawValue])
		if err != nil {
			acc.AddError(fmt.Errorf("converting %s=%v failed: %w", rawValue, metrics[rawValue], err))
			continue
		}
		fields[key] = float
	}

	val, err := internal.ToString(metrics["ups.firmware"])
	if err != nil {
		acc.AddError(fmt.Errorf("converting ups.firmware=%q failed: %w", metrics["ups.firmware"], err))
	} else {
		fields["firmware"] = val
	}

	val_ups_vendorid, err := internal.ToString(metrics["ups.vendorid"])
	if err != nil {
		acc.AddError(fmt.Errorf("converting ups.vendorid=%q failed: %w", metrics["ups.vendorid"], err))
	} else {
		fields["ups_vendorid"] = val_ups_vendorid
	}

	val_ups_beeper_status, err := internal.ToString(metrics["ups.beeper.status"])
	if err != nil {
		acc.AddError(fmt.Errorf("converting ups.beeper.status=%q failed: %w", metrics["ups.beeper.status"], err))
	} else {
		fields["ups_beeper_status"] = val_ups_beeper_status
	}

	val_ups_shutdown, err := internal.ToString(metrics["ups.shutdown"])
	if err != nil {
		acc.AddError(fmt.Errorf("converting ups.shutdown=%q failed: %w", metrics["ups.shutdown"], err))
	} else {
		fields["ups_shutdown"] = val_ups_shutdown
	}

	acc.AddFields("upsd", fields, tags)
}

func (u *Upsd) mapStatus(metrics map[string]interface{}, tags map[string]string) uint64 {
	status := uint64(0)
	statusString := fmt.Sprintf("%v", metrics["ups.status"])
	statuses := strings.Fields(statusString)
	//Source: 1.3.2 at http://rogerprice.org/NUT/ConfigExamples.A5.pdf
	//apcupsd bits:
	//0	Runtime calibration occurring (Not reported by Smart UPS v/s and BackUPS Pro)
	//1	SmartTrim (Not reported by 1st and 2nd generation SmartUPS models)
	//2	SmartBoost
	//3	On line (this is the normal condition)
	//4	On battery
	//5	Overloaded output
	//6	Battery low
	//7	Replace battery
	if choice.Contains("CAL", statuses) {
		status |= 1 << 0
		tags["status_CAL"] = "true"
	}
	if choice.Contains("TRIM", statuses) {
		status |= 1 << 1
		tags["status_TRIM"] = "true"
	}
	if choice.Contains("BOOST", statuses) {
		status |= 1 << 2
		tags["status_BOOST"] = "true"
	}
	if choice.Contains("OL", statuses) {
		status |= 1 << 3
		tags["status_OL"] = "true"
	}
	if choice.Contains("OB", statuses) {
		status |= 1 << 4
		tags["status_OB"] = "true"
	}
	if choice.Contains("OVER", statuses) {
		status |= 1 << 5
		tags["status_OVER"] = "true"
	}
	if choice.Contains("LB", statuses) {
		status |= 1 << 6
		tags["status_LB"] = "true"
	}
	if choice.Contains("RB", statuses) {
		status |= 1 << 7
		tags["status_RB"] = "true"
	}
	return status
}

func (u *Upsd) fetchVariables(server string, port int) (map[string][]nut.Variable, error) {
	client, err := nut.Connect(server, port)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if u.Username != "" && u.Password != "" {
		_, err = client.Authenticate(u.Username, u.Password)
		if err != nil {
			return nil, fmt.Errorf("auth: %w", err)
		}
	}

	upsList, err := client.GetUPSList()
	if err != nil {
		return nil, fmt.Errorf("getupslist: %w", err)
	}

	defer func() {
		_, disconnectErr := client.Disconnect()
		if disconnectErr != nil {
			err = fmt.Errorf("disconnect: %w", disconnectErr)
		}
	}()

	result := make(map[string][]nut.Variable)
	for _, ups := range upsList {
		result[ups.Name] = ups.Variables
	}

	return result, err
}

func init() {
	inputs.Add("upsd", func() telegraf.Input {
		return &Upsd{
			Server: defaultAddress,
			Port:   defaultPort,
		}
	})
}
