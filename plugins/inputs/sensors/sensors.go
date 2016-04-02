// +build linux,sensors

package sensors

import (
	"strings"

	"github.com/md14454/gosensors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Sensors struct {
	Sensors []string
}

func (_ *Sensors) Description() string {
	return "Monitor sensors using lm-sensors package"
}

var sensorsSampleConfig = `
  ## By default, telegraf gathers stats from all sensors detected by the
  ## lm-sensors module.
  ##
  ## Only collect stats from the selected sensors. Sensors are listed as
  ## <chip name>:<feature name>. This information can be found by running the
  ## sensors command, e.g. sensors -u
  ##
  ## A * as the feature name will return all features of the chip
  ##
  # sensors = ["coretemp-isa-0000:Core 0", "coretemp-isa-0001:*"]
`

func (_ *Sensors) SampleConfig() string {
	return sensorsSampleConfig
}

func (s *Sensors) Gather(acc telegraf.Accumulator) error {
	gosensors.Init()
	defer gosensors.Cleanup()

	for _, chip := range gosensors.GetDetectedChips() {
		for _, feature := range chip.GetFeatures() {
			chipName := chip.String()
			featureLabel := feature.GetLabel()

			if len(s.Sensors) != 0 {
				var found bool

				for _, sensor := range s.Sensors {
					parts := strings.SplitN(sensor, ":", 2)

					if parts[0] == chipName {
						if parts[1] == "*" || parts[1] == featureLabel {
							found = true
							break
						}
					}
				}

				if !found {
					continue
				}
			}

			tags := map[string]string{
				"chip":          chipName,
				"adapter":       chip.AdapterName(),
				"feature-name":  feature.Name,
				"feature-label": featureLabel,
			}

			fieldName := chipName + ":" + featureLabel

			fields := map[string]interface{}{
				fieldName: feature.GetValue(),
			}

			acc.AddFields("sensors", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("sensors", func() telegraf.Input {
		return &Sensors{}
	})
}
