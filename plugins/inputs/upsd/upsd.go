package upsd

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
	nut "github.com/robbiet480/go.nut"
	"strings"
)

//See: https://networkupstools.org/docs/developer-guide.chunked/index.html

const defaultAddress = "127.0.0.1"
const defaultPort = 3493

type Upsd struct {
	Server   string
	Port     int
	Username string
	Password string
	Log      telegraf.Logger `toml:"-"`

	batteryRuntimeTypeWarningIssued bool
}

func (*Upsd) Description() string {
	return "Monitor UPSes connected via Network UPS Tools"
}

var sampleConfig = `
 ## A running NUT server to connect to.
 # server = "127.0.0.1"
 # port = 3493
 # username = "user"
 # password = "password"
`

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

	timeLeftS, ok := metrics["battery.runtime"].(int64)
	if !ok && !u.batteryRuntimeTypeWarningIssued {
		u.Log.Warnf("'battery.runtime' type is not int64")
		u.batteryRuntimeTypeWarningIssued = true
	}

	fields := map[string]interface{}{
		"status_flags":            status,
		"ups.status":              metrics["ups.status"],
		"input_voltage":           metrics["input.voltage"],
		"load_percent":            metrics["ups.load"],
		"battery_charge_percent":  metrics["battery.charge"],
		"time_left_ns":            timeLeftS * 1_000_000_000, //Compatibility with apcupsd metrics format
		"output_voltage":          metrics["output.voltage"],
		"internal_temp":           metrics["ups.temperature"],
		"battery_voltage":         metrics["battery.voltage"],
		"input_frequency":         metrics["input.frequency"],
		"nominal_input_voltage":   metrics["input.voltage.nominal"],
		"nominal_battery_voltage": metrics["battery.voltage.nominal"],
		"nominal_power":           metrics["ups.realpower.nominal"],
		"firmware":                metrics["ups.firmware"],
		"battery_date":            metrics["battery.mfr.date"],
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
