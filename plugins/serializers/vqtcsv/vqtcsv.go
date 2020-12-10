package vqtcsv

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type NameStruct struct {
	measurement string
	field       string
	tags        map[string]string
}

type VqtCsvSerializer struct {
}

func parseOpcQualityString(qualityIn string) int {
	quality := 192
	switch qualityIn {
	case "Bad [Non-Specific]":
		quality = 0
		break
	case "Bad [Configuration Error]":
		quality = 4
		break
	case "Bad [Not Connected]":
		quality = 8
		break
	case "Bad [Device Failure]":
		quality = 12
		break
	case "Bad [Sensor Failure]":
		quality = 16
		break
	case "Bad [Last Known Value]":
		quality = 20
		break
	case "Bad [Communication Failure]":
		quality = 24
		break
	case "Bad [Out of Service]":
		quality = 28
		break
	case "Uncertain [Non-Specific]":
		quality = 64
		break
	case "Uncertain [Non-Specific] (Low Limited)":
		quality = 65
		break
	case "Uncertain [Non-Specific] (High Limited)":
		quality = 66
		break
	case "Uncertain [Non-Specific] (Constant)":
		quality = 67
		break
	case "Uncertain [Last Usable]":
		quality = 68
		break
	case "Uncertain [Last Usable] (Low Limited)":
		quality = 69
		break
	case "Uncertain [Last Usable] (High Limited)":
		quality = 70
		break
	case "Uncertain [Last Usable] (Constant)":
		quality = 71
		break
	case "Uncertain [Sensor Not Accurate]":
		quality = 80
		break
	case "Uncertain [Sensor Not Accurate] (Low Limited)":
		quality = 81
		break
	case "Uncertain [Sensor Not Accurate] (High Limited)":
		quality = 82
		break
	case "Uncertain [Sensor Not Accurate] (Constant)":
		quality = 83
		break
	case "Uncertain [EU Exceeded]":
		quality = 84
		break
	case "Uncertain [EU Exceeded] (Low Limited)":
		quality = 85
		break
	case "Uncertain [EU Exceeded] (High Limited)":
		quality = 86
		break
	case "Uncertain [EU Exceeded] (Constant)":
		quality = 87
		break
	case "Uncertain [Sub-Normal]":
		quality = 88
		break
	case "Uncertain [Sub-Normal] (Low Limited)":
		quality = 89
		break
	case "Uncertain [Sub-Normal] (High Limited)":
		quality = 90
		break
	case "Uncertain [Sub-Normal] (Constant)":
		quality = 91
		break
	case "Good [Non-Specific]":
		quality = 192
		break
	case "Good [Non-Specific] (Low Limited)":
		quality = 193
		break
	case "Good [Non-Specific] (High Limited)":
		quality = 194
		break
	case "Good [Non-Specific] (Constant)":
		quality = 195
		break
	case "Good [Local Override]":
		quality = 216
		break
	case "Good [Local Override] (Low Limited)":
		quality = 217
		break
	case "Good [Local Override] (High Limited)":
		quality = 218
		break
	case "Good [Local Override] (Constant)":
		quality = 219
		break
	}

	return quality
}

func parseOpcQuality(qualityIn int) string {
	qualityStr := "Good [Non-Specific]"
	switch qualityIn {
	case 0:
		qualityStr = "Bad [Non-Specific]"
		break
	case 4:
		qualityStr = "Bad [Configuration Error]"
		break
	case 8:
		qualityStr = "Bad [Not Connected]"
		break
	case 12:
		qualityStr = "Bad [Device Failure]"
		break
	case 16:
		qualityStr = "Bad [Sensor Failure]"
		break
	case 20:
		qualityStr = "Bad [Last Known Value]"
		break
	case 24:
		qualityStr = "Bad [Communication Failure]"
		break
	case 28:
		qualityStr = "Bad [Out of Service]"
		break
	case 64:
		qualityStr = "Uncertain [Non-Specific]"
		break
	case 65:
		qualityStr = "Uncertain [Non-Specific] (Low Limited)"
		break
	case 66:
		qualityStr = "Uncertain [Non-Specific] (High Limited)"
		break
	case 67:
		qualityStr = "Uncertain [Non-Specific] (Constant)"
		break
	case 68:
		qualityStr = "Uncertain [Last Usable]"
		break
	case 69:
		qualityStr = "Uncertain [Last Usable] (Low Limited)"
		break
	case 70:
		qualityStr = "Uncertain [Last Usable] (High Limited)"
		break
	case 71:
		qualityStr = "Uncertain [Last Usable] (Constant)"
		break
	case 80:
		qualityStr = "Uncertain [Sensor Not Accurate]"
		break
	case 81:
		qualityStr = "Uncertain [Sensor Not Accurate] (Low Limited)"
		break
	case 82:
		qualityStr = "Uncertain [Sensor Not Accurate] (High Limited)"
		break
	case 83:
		qualityStr = "Uncertain [Sensor Not Accurate] (Constant)"
		break
	case 84:
		qualityStr = "Uncertain [EU Exceeded]"
		break
	case 85:
		qualityStr = "Uncertain [EU Exceeded] (Low Limited)"
		break
	case 86:
		qualityStr = "Uncertain [EU Exceeded] (High Limited)"
		break
	case 87:
		qualityStr = "Uncertain [EU Exceeded] (Constant)"
		break
	case 88:
		qualityStr = "Uncertain [Sub-Normal]"
		break
	case 89:
		qualityStr = "Uncertain [Sub-Normal] (Low Limited)"
		break
	case 90:
		qualityStr = "Uncertain [Sub-Normal] (High Limited)"
		break
	case 91:
		qualityStr = "Uncertain [Sub-Normal] (Constant)"
		break
	case 192:
		qualityStr = "Good [Non-Specific]"
		break
	case 193:
		qualityStr = "Good [Non-Specific] (Low Limited)"
		break
	case 194:
		qualityStr = "Good [Non-Specific] (High Limited)"
		break
	case 195:
		qualityStr = "Good [Non-Specific] (Constant)"
		break
	case 216:
		qualityStr = "Good [Local Override]"
		break
	case 217:
		qualityStr = "Good [Local Override] (Low Limited)"
		break
	case 218:
		qualityStr = "Good [Local Override] (High Limited)"
		break
	case 219:
		qualityStr = "Good [Local Override] (Constant)"
		break
	}

	return qualityStr
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
			quality = parseOpcQualityString(val)
		}
		if val, ok := metric.Tags()["quality"]; ok {
			quality = parseOpcQualityString(val)
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
