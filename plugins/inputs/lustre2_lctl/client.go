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

func gatherClientMDCAcitve(measurement string, acc telegraf.Accumulator) {
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
}

func gatherClientOSCAcitve(measurement string, acc telegraf.Accumulator) {
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

func gatherClient(collect []string, namespace string, acc telegraf.Accumulator) {
	measurement := namespace + "_client"

	for _, c := range collect {
		switch c {
		case "mdc.*.active":
			gatherClientMDCAcitve(measurement, acc)
		case "osc.*.active":
			gatherClientOSCAcitve(measurement, acc)
		}
	}
}
