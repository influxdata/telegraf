package modbus_gateway

import (
	"github.com/influxdata/telegraf/metric"
	"github.com/prometheus/common/log"
	"math"
	"time"
)

func outputToGroup(grouper *metric.SeriesGrouper, req *Request, field *FieldDef, value interface{}, timestamp time.Time) {
	if field.Omit == false {
		writeableValue := scale(field, value)
		grouper.Add(req.MeasurementName, nil, timestamp, field.Name, writeableValue)
	}
}

func scale(f *FieldDef, value interface{}) interface{} {
	switch f.OutputFormat {
	case "FLOAT", "FLOAT64":
		switch v := value.(type) {
		case int:
			return float64((float64(v) * f.Scale) + f.Offset)
		case int16:
			return float64((float64(v) * f.Scale) + f.Offset)
		case uint16:
			return float64((float64(v) * f.Scale) + f.Offset)
		case int32:
			return float64((float64(v) * f.Scale) + f.Offset)
		case uint32:
			return float64((float64(v) * f.Scale) + f.Offset)
		default:
			return nil
		}

	case "FLOAT32":
		switch v := value.(type) {
		case int:
			return float32((float64(v) * f.Scale) + f.Offset)
		case int16:
			return float32((float64(v) * f.Scale) + f.Offset)
		case uint16:
			return float32((float64(v) * f.Scale) + f.Offset)
		case int32:
			return float32((float64(v) * f.Scale) + f.Offset)
		case uint32:
			return float32((float64(v) * f.Scale) + f.Offset)
		default:
			return nil
		}

	case "INT", "INT64":
		switch v := value.(type) {
		case int:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int16:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint16:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int32:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint32:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		default:
			return nil
		}

	case "UINT", "UINT64":
		switch v := value.(type) {
		case int:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int16:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint16:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int32:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint32:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		default:
			return nil
		}

	default:
		log.Warn("Invalid output format")
		return nil
	}
}
