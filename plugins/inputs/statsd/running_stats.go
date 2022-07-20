package statsd

import (
	"math"
	"math/rand"
	"sort"
)

const defaultPercentileLimit = 1000
const defaultMedianLimit = 1000

// RunningStats calculates a running mean, variance, standard deviation,
// lower bound, upper bound, count, and can calculate estimated percentiles.
// It is based on the incremental algorithm described here:
//    https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
type RunningStats struct {
	k   float64
	n   int64
	ex  float64
	ex2 float64

	// Array used to calculate estimated percentiles
	// We will store a maximum of PercLimit values, at which point we will start
	// randomly replacing old values, hence it is an estimated percentile.
	perc      []float64
	PercLimit int

	sum float64

	lower float64
	upper float64

	// cache if we have sorted the list so that we never re-sort a sorted list,
	// which can have very bad performance.
	SortedPerc bool

	// Array used to calculate estimated median values
	// We will store a maximum of MedLimit values, at which point we will start
	// slicing old values
	med            []float64
	MedLimit       int
	MedInsertIndex int
}

func (rs *RunningStats) AddValue(v float64) {
	// Whenever a value is added, the list is no longer sorted.
	rs.SortedPerc = false

	if rs.n == 0 {
		rs.k = v
		rs.upper = v
		rs.lower = v
		if rs.PercLimit == 0 {
			rs.PercLimit = defaultPercentileLimit
		}
		if rs.MedLimit == 0 {
			rs.MedLimit = defaultMedianLimit
			rs.MedInsertIndex = 0
		}
		rs.perc = make([]float64, 0, rs.PercLimit)
		rs.med = make([]float64, 0, rs.MedLimit)
	}

	// These are used for the running mean and variance
	rs.n++
	rs.ex += v - rs.k
	rs.ex2 += (v - rs.k) * (v - rs.k)

	// add to running sum
	rs.sum += v

	// track upper and lower bounds
	if v > rs.upper {
		rs.upper = v
	} else if v < rs.lower {
		rs.lower = v
	}

	if len(rs.perc) < rs.PercLimit {
		rs.perc = append(rs.perc, v)
	} else {
		// Reached limit, choose random index to overwrite in the percentile array
		rs.perc[rand.Intn(len(rs.perc))] = v
	}

	if len(rs.med) < rs.MedLimit {
		rs.med = append(rs.med, v)
	} else {
		// Reached limit, start over
		rs.med[rs.MedInsertIndex] = v
		rs.MedInsertIndex = (rs.MedInsertIndex + 1) % rs.MedLimit
	}
}

func (rs *RunningStats) Mean() float64 {
	return rs.k + rs.ex/float64(rs.n)
}

func (rs *RunningStats) Median() float64 {
	// Need to sort for median, but keep temporal order
	var values []float64
	values = append(values, rs.med...)
	sort.Float64s(values)
	count := len(values)
	if count == 0 {
		return 0
	} else if count%2 == 0 {
		return (values[count/2-1] + values[count/2]) / 2
	} else {
		return values[count/2]
	}
}

func (rs *RunningStats) Variance() float64 {
	return (rs.ex2 - (rs.ex*rs.ex)/float64(rs.n)) / float64(rs.n)
}

func (rs *RunningStats) Stddev() float64 {
	return math.Sqrt(rs.Variance())
}

func (rs *RunningStats) Sum() float64 {
	return rs.sum
}

func (rs *RunningStats) Upper() float64 {
	return rs.upper
}

func (rs *RunningStats) Lower() float64 {
	return rs.lower
}

func (rs *RunningStats) Count() int64 {
	return rs.n
}

func (rs *RunningStats) Percentile(n float64) float64 {
	if n > 100 {
		n = 100
	}

	if !rs.SortedPerc {
		sort.Float64s(rs.perc)
		rs.SortedPerc = true
	}

	i := float64(len(rs.perc)) * n / float64(100)
	return rs.perc[clamp(i, 0, len(rs.perc)-1)]
}

func clamp(i float64, min int, max int) int {
	if i < float64(min) {
		return min
	}
	if i > float64(max) {
		return max
	}
	return int(i)
}
