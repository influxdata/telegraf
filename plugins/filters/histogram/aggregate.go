package histogram

import (
	"github.com/VividCortex/gohistogram"
	"math"
)

type Aggregate struct {
	hist *gohistogram.NumericHistogram
	sum  float64
	max  float64
	min  float64
}

func (a *Aggregate) Add(n float64) {
	a.sum += n
	if a.max < n {
		a.max = n
	}
	if a.min > n {
		a.min = n
	}
	a.hist.Add(n)
}

func (a *Aggregate) Quantile(q float64) float64 {
	return a.hist.Quantile(q)
}

func (a *Aggregate) Sum() float64 {
	return a.sum
}

func (a *Aggregate) CDF(x float64) float64 {
	return a.hist.CDF(x)
}

func (a *Aggregate) Mean() float64 {
	return a.hist.Mean()
}

func (a *Aggregate) Variance() float64 {
	return a.hist.Variance()
}

func (a *Aggregate) Count() float64 {
	return a.hist.Count()
}

func (a *Aggregate) Max() float64 {
	return a.max
}

func (a *Aggregate) Min() float64 {
	return a.min
}

func NewAggregate(n int) *Aggregate {
	return &Aggregate{
		hist: gohistogram.NewHistogram(n),
		max:  math.SmallestNonzeroFloat64,
		min:  math.MaxFloat64,
		sum:  0,
	}
}
