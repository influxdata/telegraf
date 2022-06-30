package health

import (
	"github.com/influxdata/telegraf"
)

type Compares struct {
	Field string   `toml:"field"`
	GT    *float64 `toml:"gt"`
	GE    *float64 `toml:"ge"`
	LT    *float64 `toml:"lt"`
	LE    *float64 `toml:"le"`
	EQ    *float64 `toml:"eq"`
	NE    *float64 `toml:"ne"`
}

func (c *Compares) runChecks(fv float64) bool {
	if c.GT != nil && !(fv > *c.GT) {
		return false
	}
	if c.GE != nil && !(fv >= *c.GE) {
		return false
	}
	if c.LT != nil && !(fv < *c.LT) {
		return false
	}
	if c.LE != nil && !(fv <= *c.LE) {
		return false
	}
	if c.EQ != nil && !(fv == *c.EQ) {
		return false
	}
	if c.NE != nil && !(fv != *c.NE) {
		return false
	}
	return true
}

func (c *Compares) Check(metrics []telegraf.Metric) bool {
	success := true
	for _, m := range metrics {
		fv, ok := m.GetField(c.Field)
		if !ok {
			continue
		}

		f, ok := asFloat(fv)
		if !ok {
			return false
		}

		result := c.runChecks(f)
		if !result {
			success = false
		}
	}
	return success
}

func asFloat(fv interface{}) (float64, bool) {
	switch v := fv.(type) {
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case uint64:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	default:
		return 0.0, false
	}
}
