package logql

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type response struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

type stream struct {
	Labels map[string]string `json:"stream"`
	Lines  []logline         `json:"values"`
}

type vector struct {
	Labels map[string]string `json:"metric"`
	Value  value             `json:"value"`
}

type matrix struct {
	Labels map[string]string `json:"metric"`
	Values []value           `json:"values"`
}

type value struct {
	timestamp time.Time
	value     float64
}

// UnmarshalJSON customizes the JSON parsing to decode the raw pair-array into
// timestamp and the numeric value
func (v *value) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("unexpected number of entries (%d) in %v", len(raw), string(data))
	}

	ts, ok := raw[0].(float64)
	if !ok {
		return fmt.Errorf("unexpected type %T for timestamp %v", raw[0], raw[0])
	}
	v.timestamp = time.Unix(0, int64(ts*1e9))

	rawValue, ok := raw[1].(string)
	if !ok {
		return fmt.Errorf("unexpected type %T for value %v", raw[1], raw[1])
	}
	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return fmt.Errorf("parsing value failed: %w", err)
	}
	v.value = value

	return nil
}

type logline struct {
	timestamp time.Time
	message   string
}

// UnmarshalJSON customizes the JSON parsing to decode the raw string pair-array
// into timestamp and message
func (l *logline) UnmarshalJSON(data []byte) error {
	var raw []string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("unexpected number of entries (%d) in %v", len(raw), string(data))
	}

	ts, err := strconv.ParseInt(raw[0], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing timestamp failed: %w", err)
	}
	l.timestamp = time.Unix(0, ts)
	l.message = raw[1]

	return nil
}

func (r *response) parse() (interface{}, error) {
	if r.Status != "success" {
		return nil, fmt.Errorf("invalid status %q", r.Status)
	}

	switch r.Data.ResultType {
	case "vector":
		var v []vector
		if err := json.Unmarshal(r.Data.Result, &v); err != nil {
			return nil, fmt.Errorf("decoding %q result failed: %w", r.Data.ResultType, err)
		}
		return v, nil
	case "matrix":
		var m []matrix
		if err := json.Unmarshal(r.Data.Result, &m); err != nil {
			return nil, fmt.Errorf("decoding %q result failed: %w", r.Data.ResultType, err)
		}
		return m, nil
	case "streams":
		var s []stream
		if err := json.Unmarshal(r.Data.Result, &s); err != nil {
			return nil, fmt.Errorf("decoding %q result failed: %w", r.Data.ResultType, err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unknown result type %q", r.Data.ResultType)
	}
}
