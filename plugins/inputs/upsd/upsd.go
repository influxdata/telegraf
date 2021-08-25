package upsd

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	nut "github.com/robbiet480/go.nut"
	"strings"
)

//See: https://networkupstools.org/docs/developer-guide.chunked/index.html

const defaultAddress = "127.0.0.1"

type Upsd struct {
	Servers  []string
	Username string
	Password string
}

func (*Upsd) Description() string {
	return "Monitor UPSes connected via Network UPS Tools"
}

var sampleConfig = `
  # A list of running NUT servers to connect to.
  # If not provided will default to 127.0.0.1
  servers = ["127.0.0.1"]
  # username = "user"
  # password = "password"
`

func (*Upsd) SampleConfig() string {
	return sampleConfig
}

func (h *Upsd) Gather(accumulator telegraf.Accumulator) error {
	for _, server := range h.Servers {
		err := func(Server string) error {
			upsList, err := fetchVariables(server, h.Username, h.Password)
			if err != nil {
				return err
			}

			for name, variables := range upsList {
				h.GatherUps(accumulator, name, variables)
			}

			return nil
		}(server)

		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Upsd) GatherUps(accumulator telegraf.Accumulator, name string, variables []nut.Variable) {
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

	status := h.mapStatus(metrics)

	timeLeftS := metrics["battery.runtime"].(int64)

	fields := map[string]interface{}{
		"status_flags":           status,
		"input_voltage":          metrics["input.voltage"],
		"load_percent":           metrics["ups.load"],
		"battery_charge_percent": metrics["battery.charge"],
		"time_left_ns":           timeLeftS * 1_000_000_000,
		"output_voltage":         metrics["output.voltage"],
		"internal_temp":          metrics["ups.temperature"],
		"battery_voltage":        metrics["battery.voltage"],
		"input_frequency":        metrics["input.frequency"],
		//"time_on_battery_ns": no clue how to get this one,
		"nominal_input_voltage":   metrics["input.voltage.nominal"],
		"nominal_battery_voltage": metrics["battery.voltage.nominal"],
		"nominal_power":           metrics["ups.realpower.nominal"],
		"firmware":                metrics["ups.firmware"],
		"battery_date":            metrics["battery.mfr.date"],
	}

	accumulator.AddFields("upsd", fields, tags)
}

func (h *Upsd) mapStatus(metrics map[string]interface{}) uint64 {
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
	if contains(statuses, "CAL") {
		status |= 1 << 0
	}
	if contains(statuses, "TRIM") {
		status |= 1 << 1
	}
	if contains(statuses, "BOOST") {
		status |= 1 << 2
	}
	if contains(statuses, "OL") {
		status |= 1 << 3
	}
	if contains(statuses, "OB") {
		status |= 1 << 4
	}
	if contains(statuses, "OVER") {
		status |= 1 << 5
	}
	if contains(statuses, "LB") {
		status |= 1 << 6
	}
	if contains(statuses, "RB") {
		status |= 1 << 7
	}
	return status
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func fetchVariables(server string, username string, password string) (map[string][]nut.Variable, error) {
	client, connectErr := nut.Connect(server)

	if connectErr != nil {
		fmt.Print(connectErr)
	}

	if username != "" && password != "" {
		_, authErr := client.Authenticate(username, password)
		if authErr != nil {
			fmt.Print(authErr)
		}
	}

	upsList, listErr := client.GetUPSList()
	if listErr != nil {
		fmt.Print(listErr)
	}

	defer client.Disconnect()

	result := make(map[string][]nut.Variable)
	for _, ups := range upsList {
		result[ups.Name] = ups.Variables
	}

	return result, nil
}

func init() {
	inputs.Add("upsd", func() telegraf.Input {
		return &Upsd{
			Servers:  []string{defaultAddress},
			Username: "",
			Password: "",
		}
	})
}
