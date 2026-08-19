package cisco_telemetry_mdt

import (
	"strconv"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"

	"github.com/influxdata/telegraf/internal"
)

func decode(field *telemetry.TelemetryField) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		return val.BytesValue
	case *telemetry.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	case *telemetry.TelemetryField_BoolValue:
		return val.BoolValue
	case *telemetry.TelemetryField_Uint32Value:
		return val.Uint32Value
	case *telemetry.TelemetryField_Uint64Value:
		return val.Uint64Value
	case *telemetry.TelemetryField_Sint32Value:
		return val.Sint32Value
	case *telemetry.TelemetryField_Sint64Value:
		return val.Sint64Value
	case *telemetry.TelemetryField_DoubleValue:
		return val.DoubleValue
	case *telemetry.TelemetryField_FloatValue:
		return val.FloatValue
	}
	return nil
}

func decodeTag(field *telemetry.TelemetryField) string {
	if value := decode(field); value != nil {
		if tag, err := internal.ToString(value); err == nil {
			return tag
		}
	}
	return ""
}

// xform Field to string
func xformValueString(field *telemetry.TelemetryField) string {
	return decodeTag(field)
}

// xform Uint64 to int64
func nxosValueXformUint64Toint64(field *telemetry.TelemetryField) interface{} {
	if value := field.GetUint64Value(); value != 0 {
		return int64(value)
	}
	return nil
}

// xform string to float
func nxosValueXformStringTofloat(field *telemetry.TelemetryField) interface{} {
	if value := field.GetStringValue(); value != "" {
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return v
		}
	}
	return nil
}

// xform string to uint64
func nxosValueXformStringToUint64(field *telemetry.TelemetryField) interface{} {
	if value := field.GetStringValue(); value != "" {
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			return v
		}
	}
	return nil
}

// xform string to int64
func nxosValueXformStringToInt64(field *telemetry.TelemetryField) interface{} {
	if value := field.GetStringValue(); value != "" {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return v
		}
	}
	return nil
}

// auto-xform float properties
func nxosValueAutoXformFloatProp(field *telemetry.TelemetryField) interface{} {
	if value := field.GetStringValue(); value != "" {
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return v
		}
	}
	return nil
}

// xform uint64 to string
func nxosValueXformUint64ToString(field *telemetry.TelemetryField) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	case *telemetry.TelemetryField_Uint64Value:
		return strconv.FormatUint(val.Uint64Value, 10)
	}
	return nil
}

// Xform value field.
func nxosValueXform(field *telemetry.TelemetryField, propMap map[string]func(*telemetry.TelemetryField) interface{}, prop map[string]string) interface{} {
	value := decode(field)
	if value == nil {
		return nil
	}

	if _, ok := propMap[field.Name]; ok {
		return propMap[field.Name](field)
	}
	// check if we want auto xformation
	if _, ok := propMap["auto-prop-xfromi"]; ok {
		return propMap["auto-prop-xfrom"](field)
	}
	// Now check path based conversion; if a mapping exists apply the transformation
	if prop == nil {
		return nil
	}

	// Xformation supported is only from String, Uint32 and Uint64
	switch prop[field.Name] {
	case "integer":
		switch v := value.(type) {
		case string:
			if x, err := strconv.ParseInt(v, 10, 32); err == nil {
				return x
			}
		case uint32:
			return v
		case uint64:
			return v
		}
		return nil
	// Xformation supported is only from String
	case "float":
		if v, ok := value.(string); ok {
			if x, err := strconv.ParseFloat(v, 64); err == nil {
				return x
			}
		}
		return nil
	case "string":
		return xformValueString(field)
	case "int64":
		switch v := value.(type) {
		case string:
			if x, err := strconv.ParseInt(v, 10, 64); err == nil {
				return x
			}
		case uint64:
			return int64(v)
		}
	}
	return nil
}
