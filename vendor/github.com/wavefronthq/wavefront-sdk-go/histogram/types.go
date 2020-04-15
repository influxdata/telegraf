package histogram

import "time"

// Centroid encapsulates a mean value and the count of points associated with that value.
type Centroid struct {
	Value float64
	Count int
}

// Granularity is the interval (MINUTE, HOUR and/or DAY) by which the histogram data should be aggregated.
type Granularity int8

const (
	MINUTE Granularity = iota
	HOUR
	DAY
)

// Duration of the Granularity
func (hg *Granularity) Duration() time.Duration {
	switch *hg {
	case MINUTE:
		return time.Minute
	case HOUR:
		return time.Hour
	default:
		return time.Hour * 24
	}
}

func (hg *Granularity) String() string {
	switch *hg {
	case MINUTE:
		return "!M"
	case HOUR:
		return "!H"
	default: // DAY
		return "!D"
	}
}
