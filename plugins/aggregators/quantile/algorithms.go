package quantile

import (
	"math"
	"sort"

	"github.com/caio/go-tdigest"
)

type algorithm interface {
	Add(value float64) error
	Quantile(q float64) float64
}

func newTDigest(compression float64) (algorithm, error) {
	return tdigest.New(tdigest.Compression(compression))
}

type exactAlgorithmR7 struct {
	xs     []float64
	sorted bool
}

func newExactR7(_ float64) (algorithm, error) {
	return &exactAlgorithmR7{xs: make([]float64, 0, 100), sorted: false}, nil
}

func (e *exactAlgorithmR7) Add(value float64) error {
	e.xs = append(e.xs, value)
	e.sorted = false

	return nil
}

func (e *exactAlgorithmR7) Quantile(q float64) float64 {
	size := len(e.xs)

	// No information
	if len(e.xs) == 0 {
		return math.NaN()
	}

	// Sort the array if necessary
	if !e.sorted {
		sort.Float64s(e.xs)
		e.sorted = true
	}

	// Get the quantile index and the fraction to the neighbor
	// Hyndman & Fan; Sample Quantiles in Statistical Packages; The American Statistician vol 50; pp 361-365; 1996 -- R7
	// Same as Excel and Numpy.
	n := q * (float64(size) - 1)
	i, gamma := math.Modf(n)
	j := int(i)
	if j < 0 {
		return e.xs[0]
	}
	if j >= size {
		return e.xs[size-1]
	}
	// Linear interpolation
	return e.xs[j] + gamma*(e.xs[j+1]-e.xs[j])
}

type exactAlgorithmR8 struct {
	xs     []float64
	sorted bool
}

func newExactR8(_ float64) (algorithm, error) {
	return &exactAlgorithmR8{xs: make([]float64, 0, 100), sorted: false}, nil
}

func (e *exactAlgorithmR8) Add(value float64) error {
	e.xs = append(e.xs, value)
	e.sorted = false

	return nil
}

func (e *exactAlgorithmR8) Quantile(q float64) float64 {
	size := len(e.xs)

	// No information
	if size == 0 {
		return math.NaN()
	}

	// Sort the array if necessary
	if !e.sorted {
		sort.Float64s(e.xs)
		e.sorted = true
	}

	// Get the quantile index and the fraction to the neighbor
	// Hyndman & Fan; Sample Quantiles in Statistical Packages; The American Statistician vol 50; pp 361-365; 1996 -- R8
	n := q*(float64(size)+1.0/3.0) - (2.0 / 3.0) // Indices are zero-base here but one-based in the paper
	i, gamma := math.Modf(n)
	j := int(i)
	if j < 0 {
		return e.xs[0]
	}
	if j >= size {
		return e.xs[size-1]
	}
	// Linear interpolation
	return e.xs[j] + gamma*(e.xs[j+1]-e.xs[j])
}
