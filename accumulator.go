package telegraf

import "time"

type Accumulator interface {
	// Create a point with a value, decorating it with tags
	// NOTE: tags is expected to be owned by the caller, don't mutate
	// it after passing to Add.
	Add(measurement string,
		value interface{},
		tags map[string]string,
		t ...time.Time)

	AddFields(measurement string,
		fields map[string]interface{},
		tags map[string]string,
		t ...time.Time)

	Debug() bool
	SetDebug(enabled bool)

	SetPrecision(precision, interval time.Duration)

	DisablePrecision()
}
