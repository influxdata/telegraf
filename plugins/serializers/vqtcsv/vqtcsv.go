package vqtcsv

import (
	"fmt"

	"github.com/influxdata/telegraf"
	opc "github.com/influxdata/telegraf/lib"
)

type NameStruct struct {
	measurement string
	field       string
	tags        map[string]string
}

type VqtCsvSerializer struct {
}

func (s *VqtCsvSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	out := []byte{}

	// Convert UnixNano to Unix timestamps
	timestamp := metric.Time().Format("02-01-2006 15:04:05.000")
	quality := 192

	for fieldName, value := range metric.Fields() {
		switch v := value.(type) {
		case bool:
			if v {
				value = 1
			} else {
				value = 0
			}
		}

		if val, ok := metric.Tags()["Quality"]; ok {
			quality = opc.ParseQualityString(val)
		}
		if val, ok := metric.Tags()["quality"]; ok {
			quality = opc.ParseQualityString(val)
		}
		name := metric.Name() + "." + fieldName
		for tagName, tagValue := range metric.Tags() {
			name += "(" + tagName + "=" + tagValue + ")"
		}

		metricString := fmt.Sprintf("%s,%#v,%d,%s\n",
			name,
			value,
			quality,
			timestamp)
		point := []byte(metricString)
		out = append(out, point...)
	}
	return out, nil
}

func (s *VqtCsvSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	return nil, nil
}
