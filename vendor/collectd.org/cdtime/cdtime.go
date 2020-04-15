/*
Package cdtime implements methods to convert from and to collectd's internal time
representation, cdtime_t.
*/
package cdtime // import "collectd.org/cdtime"

import (
	"strconv"
	"time"
)

// Time represens a time in collectd's internal representation.
type Time uint64

// New returns a new Time representing time t.
func New(t time.Time) Time {
	return newNano(uint64(t.UnixNano()))
}

// NewDuration returns a new Time representing duration d.
func NewDuration(d time.Duration) Time {
	return newNano(uint64(d.Nanoseconds()))
}

// Time converts and returns the time as time.Time.
func (t Time) Time() time.Time {
	s, ns := t.decompose()
	return time.Unix(s, ns)
}

// Duration converts and returns the duration as time.Duration.
func (t Time) Duration() time.Duration {
	s, ns := t.decompose()
	return time.Duration(1000000000*s+ns) * time.Nanosecond
}

// String returns the string representation of Time. The format used is seconds
// since the epoch with millisecond precision, e.g. "1426588900.328".
func (t Time) String() string {
	f := t.Float()
	return strconv.FormatFloat(f /* format */, 'f' /* precision */, 3 /* bits */, 64)
}

// Float returns the time as seocnds since epoch. This is a lossy conversion,
// which will lose up to 11 bits. This means that the returned value should be
// considered to have roughly microsecond precision.
func (t Time) Float() float64 {
	s, ns := t.decompose()
	return float64(s) + float64(ns)/1000000000.0
}

// MarshalJSON implements the "encoding/json".Marshaler interface for Time.
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON implements the "encoding/json".Unmarshaler interface for Time.
func (t *Time) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data) /* bits */, 64)
	if err != nil {
		return err
	}

	s := uint64(f)
	ns := uint64((f - float64(s)) * 1000000000.0)

	*t = newNano(1000000000*s + ns)
	return nil
}

func (t Time) decompose() (s, ns int64) {
	s = int64(t >> 30)

	ns = (int64(t&0x3fffffff) * 1000000000)
	// add 2^29 to correct rounding behavior.
	ns = (ns + (1 << 29)) >> 30

	return
}

func newNano(ns uint64) Time {
	// break into seconds and nano-seconds so the left-shift doesn't overflow.
	s := (ns / 1000000000) << 30

	ns = (ns % 1000000000) << 30
	// add 5e8 to correct rounding behavior.
	ns = (ns + 500000000) / 1000000000

	return Time(s | ns)
}
