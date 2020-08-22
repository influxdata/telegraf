package modbus_gateway

import (
	"github.com/influxdata/telegraf/metric"
	"math"
	"time"
)

func outputToGroup(grouper *metric.SeriesGrouper, req *Request, f *FieldDef, value interface{}, timestamp time.Time) {
	if !f.Omit {
		writeableValue := scale(f, value)
		grouper.Add(req.MeasurementName, nil, timestamp, f.Name, writeableValue)
	}
}

func scale(f *FieldDef, value interface{}) interface{} {
	switch f.OutputFormat {
	case "FLOAT", "FLOAT64":
		switch v := value.(type) {
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
		return nil
	}
}
