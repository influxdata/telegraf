//go:build linux

package lustre2_lctl

import (
	"regexp"
	"strconv"

	"github.com/influxdata/telegraf"
)

var (
	clientOSCActivePattern = regexp.MustCompile(`osc.(.*)-osc-\w*.active=(\d*)`)
	clientMDCActivePattern = regexp.MustCompile(`mdc.(.*)-mdc-\w*.active=(\d*)`)
)

func gatherClient(client bool, namespace string, acc telegraf.Accumulator) {
	if !client {
		return
	}

	measurement := namespace + "_client"

	// mdc
	if result, err := executeCommand("lctl", "get_param", "mdc.*.active"); err != nil { // to get volume.
		acc.AddError(err)
	} else {
		vs := clientMDCActivePattern.FindAllStringSubmatch(result, -1)
		for _, v := range vs {
			if value, err := strconv.ParseInt(v[2], 10, 64); err != nil {
				acc.AddError(err)
			} else {
				volume := v[1]
				acc.AddGauge(measurement, map[string]interface{}{
					"mdc_volume_active": value,
				}, map[string]string{
					"volume": volume,
				})
			}
		}
	}

	// osc.
	if result, err := executeCommand("lctl", "get_param", "osc.*.active"); err != nil { // to get volume.
		acc.AddError(err)
	} else {
		vs := clientOSCActivePattern.FindAllStringSubmatch(result, -1)
		for _, v := range vs {
			if value, err := strconv.ParseInt(v[2], 10, 64); err != nil {
				acc.AddError(err)
			} else {
				volume := v[1]
				acc.AddGauge(measurement, map[string]interface{}{
					"osc_volume_active": value,
				}, map[string]string{
					"volume": volume,
				})
			}
		}
	}
}
