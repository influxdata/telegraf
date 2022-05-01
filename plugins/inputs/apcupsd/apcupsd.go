package apcupsd

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	apcupsdClient "github.com/mdlayher/apcupsd"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultAddress = "tcp://127.0.0.1:3551"

var defaultTimeout = config.Duration(5 * time.Second)

type ApcUpsd struct {
	Servers []string
	Timeout config.Duration
}

func (h *ApcUpsd) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	for _, server := range h.Servers {
		err := func(address string) error {
			addrBits, err := url.Parse(address)
			if err != nil {
				return err
			}
			if addrBits.Scheme == "" {
				addrBits.Scheme = "tcp"
			}

			ctx, cancel := context.WithTimeout(ctx, time.Duration(h.Timeout))
			defer cancel()

			status, err := fetchStatus(ctx, addrBits)
			if err != nil {
				return err
			}

			tags := map[string]string{
				"serial":   status.SerialNumber,
				"ups_name": status.UPSName,
				"status":   status.Status,
				"model":    status.Model,
			}

			flags, err := strconv.ParseUint(strings.Fields(status.StatusFlags)[0], 0, 64)
			if err != nil {
				return err
			}

			fields := map[string]interface{}{
				"status_flags":            flags,
				"input_voltage":           status.LineVoltage,
				"load_percent":            status.LoadPercent,
				"battery_charge_percent":  status.BatteryChargePercent,
				"time_left_ns":            status.TimeLeft.Nanoseconds(),
				"output_voltage":          status.OutputVoltage,
				"internal_temp":           status.InternalTemp,
				"battery_voltage":         status.BatteryVoltage,
				"input_frequency":         status.LineFrequency,
				"time_on_battery_ns":      status.TimeOnBattery.Nanoseconds(),
				"nominal_input_voltage":   status.NominalInputVoltage,
				"nominal_battery_voltage": status.NominalBatteryVoltage,
				"nominal_power":           status.NominalPower,
				"firmware":                status.Firmware,
				"battery_date":            status.BatteryDate,
			}

			acc.AddFields("apcupsd", fields, tags)
			return nil
		}(server)

		if err != nil {
			return err
		}
	}
	return nil
}

func fetchStatus(ctx context.Context, addr *url.URL) (*apcupsdClient.Status, error) {
	client, err := apcupsdClient.DialContext(ctx, addr.Scheme, addr.Host)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client.Status()
}

func init() {
	inputs.Add("apcupsd", func() telegraf.Input {
		return &ApcUpsd{
			Servers: []string{defaultAddress},
			Timeout: defaultTimeout,
		}
	})
}
