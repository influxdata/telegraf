package addons

import (
	"bytes"
	"strconv"
	"time"
)

// Timestamp wraps time.Time and adapts json.Marshaler.
//
// Timestamp refers to the JSON representation of timestamps, for
// time-data interchange, as a single integer representing a modified version of
// whole milliseconds since the UNIX epoch (00:00:00 UTC on January 1, 1970).
// Individual languages (Go, C, Java) will typically implement this structure
// using a more-complex construction in their respective contexts, containing even
// finer granularity for local data storage, typically at the nanosecond level.
//
// The "modified version" comment reflects the following simplification.
// Despite the already fine-grained representation as milliseconds, this data
// value takes no account of leap seconds; for all of our calculations, we
// simply pretend they don't exist.  Individual feeders will typically map a
// 00:00:60 value for a leap second, obtained as a string so the presence of the
// leap second is obvious, as 00:01:00, and the fact that 00:01:00 will occur
// again in the following second will be silently ignored.  This means that any
// monitoring which really wants to accurately reflect International Atomic Time
// (TAI), UT1, or similar time coordinates will be subject to some disruption.
// It also means that even in ordinary circumstances, any calculations of
// sub-second time differences might run into surprises, since the following
// timestamps could appear in temporal order:
//
//         actual time   relative reported time in milliseconds
//     A:  00:00:59.000  59000
//     B:  00:00:60.000  60000
//     C:  00:00:60.700  60700
//     D:  00:01:00.000  60000
//     E:  00:01:00.300  60300
//     F:  00:01:01.000  61000
//
// In such a situation, (D - C) and (E - C) would be negative numbers.
//
// In other situations, a feeder might obtain a timestamp from a system hardware
// clock which, say, counts local nanoseconds and has no notion of any leap
// seconds having been inserted into human-readable string-time representations.
// So there could be some amount of offset if such values are compared across
// such a boundary.
//
// Beyond that, there is always the issue of computer clocks not being directly
// tied to atomic clocks, using inexpensive non-temperature-compensated crystals
// for timekeeping.  Such hardware can easily drift dramatically off course, and
// the local timekeeping may or may not be subject to course correction using
// HTP, chrony, or similar software that periodically adjusts the system time
// to keep it synchronized with the Internet.  Also, there may be large jumps
// in either a positive or negative direction when a drifted clock is suddenly
// brought back into synchronization with the rest of the world.
//
// In addition, we ignore here all temporal effects of Special Relativity, not
// to mention further adjustments needed to account for General Relativity.
// This is not a theoretical joke; those who monitor GPS satellites should take
// note of the limitations of this data type, and use some other data type for
// time-critical data exchange and calculations.
//
// The point of all this being, fine resolution of clock values should never be
// taken too seriously unless one is sure that the clocks being compared are
// directly hitched together, and even then one must allow for quantum leaps
// into the future and time travel into the past.
//
// Finally, note that the Go zero-value of the internal implementation object
// we use in that language does not have a reasonable value when interpreted
// as milliseconds since the UNIX epoch.  For that reason, the general rule is
// that the JSON representation of a zero-value for any field of this type, no
// matter what the originating language, will be to simply omit it from the
// JSON string.  That fact must be taken into account when marshalling and
// unmarshalling data structures that contain such fields.
//
type Timestamp struct {
	time.Time
}

// NewTimestamp returns reference to new timestamp setted to UTC now.
func NewTimestamp() *Timestamp {
	return &Timestamp{time.Now().UTC()}
}

// Add returns the timestamp t+d.
// Overrides nested time.Time Add.
func (t Timestamp) Add(d time.Duration) Timestamp {
	return Timestamp{t.Time.Add(d)}
}

// AddDate returns the timestamp corresponding to adding the given number of years, months, and days.
// Overrides nested time.Time AddDate.
func (t Timestamp) AddDate(years int, months int, days int) Timestamp {
	return Timestamp{t.Time.AddDate(years, months, days)}
}

// In returns a copy of t with location set to loc.
// Overrides nested time.Time In.
func (t Timestamp) In(loc *time.Location) Timestamp {
	return Timestamp{t.Time.In(loc)}
}

// Local returns a copy of t with the location set to local time.
// Overrides nested time.Time Local.
func (t Timestamp) Local() Timestamp {
	return Timestamp{t.Time.Local()}
}

// Round returns a copy of t rounded to the nearest multiple of d.
// Overrides nested time.Time Round.
func (t Timestamp) Round(d time.Duration) Timestamp {
	return Timestamp{t.Time.Round(d)}
}

// Truncate returns a copy t rounded down to a multiple of d.
// Overrides nested time.Time Truncate.
func (t Timestamp) Truncate(d time.Duration) Timestamp {
	return Timestamp{t.Time.Truncate(d)}
}

// UTC returns a copy of t with the location set to UTC.
// Overrides nested time.Time UTC.
func (t Timestamp) UTC() Timestamp {
	return Timestamp{t.Time.UTC()}
}

// MarshalJSON implements json.Marshaler.
// Overrides nested time.Time MarshalJSON.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	i := t.UnixMilli()
	buf := make([]byte, 0, 16)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, i, 10)
	buf = append(buf, '"')
	return buf, nil
}

// String implements fmt.Stringer.
// Overrides nested time.Time String.
func (t Timestamp) String() string {
	i := t.UnixMilli()
	buf := make([]byte, 0, 16)
	buf = strconv.AppendInt(buf, i, 10)
	return string(buf)
}

// UnmarshalJSON implements json.Unmarshaler.
// Overrides nested time.Time UnmarshalJSON.
func (t *Timestamp) UnmarshalJSON(input []byte) error {
	strInput := string(bytes.Trim(input, `"`))

	i, err := strconv.ParseInt(strInput, 10, 64)
	if err != nil {
		return err
	}

	i *= int64(time.Millisecond)
	*t = Timestamp{time.Unix(0, i).UTC()}
	return nil
}
