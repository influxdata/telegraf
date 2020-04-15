package histogram

import (
	"math"
	"sync"
	"time"

	tdigest "github.com/caio/go-tdigest"
)

// Histogram a quantile approximation data structure
type Histogram interface {
	Update(v float64)
	Distributions() []Distribution
	Snapshot() []Distribution
	Count() uint64
	Quantile(q float64) float64
	Max() float64
	Min() float64
	Sum() float64
	Mean() float64
	Granularity() Granularity
}

// Option allows histogram customization
type Option func(*histogramImpl)

// GranularityOption of the histogram
func GranularityOption(g Granularity) Option {
	return func(args *histogramImpl) {
		args.granularity = g
	}
}

// Compression of the histogram
func Compression(c uint32) Option {
	return func(args *histogramImpl) {
		args.compression = c
	}
}

// MaxBins of the histogram
func MaxBins(c int) Option {
	return func(args *histogramImpl) {
		args.maxBins = c
	}
}

func defaultHistogramImpl() *histogramImpl {
	return &histogramImpl{maxBins: 10, granularity: MINUTE, compression: 5}
}

// Creates a new Wavefront histogram
func New(setters ...Option) Histogram {
	h := defaultHistogramImpl()
	for _, setter := range setters {
		setter(h)
	}
	return h
}

type histogramImpl struct {
	mutex              sync.Mutex
	priorTimedBinsList []*timedBin
	currentTimedBin    *timedBin

	granularity Granularity
	compression uint32
	maxBins     int
}

type timedBin struct {
	tdigest   *tdigest.TDigest
	timestamp time.Time
}

// Distribution holds the samples and its timestamp.
type Distribution struct {
	Centroids []Centroid
	Timestamp time.Time
}

// Update registers a new sample in the histogram.
func (h *histogramImpl) Update(v float64) {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.currentTimedBin.tdigest.Add(v)
}

// Count returns the total number of samples on this histogram.
func (h *histogramImpl) Count() uint64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.currentTimedBin.tdigest.Count()
}

// Quantile returns the desired percentile estimation.
func (h *histogramImpl) Quantile(q float64) float64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.currentTimedBin.tdigest.Quantile(q)
}

// Max returns the maximum value of samples on this histogram.
func (h *histogramImpl) Max() float64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	max := math.SmallestNonzeroFloat64
	h.currentTimedBin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
		max = math.Max(max, mean)
		return true
	})
	return max
}

// Min returns the minimum value of samples on this histogram.
func (h *histogramImpl) Min() float64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	min := math.MaxFloat64
	for _, bin := range append(h.priorTimedBinsList, h.currentTimedBin) {
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			min = math.Min(min, mean)
			return true
		})
	}
	return min
}

// Sum returns the sum of all values on this histogram.
func (h *histogramImpl) Sum() float64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	sum := float64(0)
	for _, bin := range append(h.priorTimedBinsList, h.currentTimedBin) {
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			sum += mean * float64(count)
			return true
		})
	}
	return sum
}

// Mean returns the mean values of samples on this histogram.
func (h *histogramImpl) Mean() float64 {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	t := float64(0)
	c := uint32(0)
	for _, bin := range append(h.priorTimedBinsList, h.currentTimedBin) {
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			t += mean * float64(count)
			c += count
			return true
		})
	}
	return t / float64(c)
}

// Granularity value
func (h *histogramImpl) Granularity() Granularity {
	return h.granularity
}

// Snapshot returns a copy of all samples on comlepted time slices
func (h *histogramImpl) Snapshot() []Distribution {
	return h.distributions(false)
}

// Distributions returns all samples on comlepted time slices, and clear the histogram
func (h *histogramImpl) Distributions() []Distribution {
	return h.distributions(true)
}

func (h *histogramImpl) distributions(clean bool) []Distribution {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	distributions := make([]Distribution, len(h.priorTimedBinsList))
	for idx, bin := range h.priorTimedBinsList {
		var centroids []Centroid
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			centroids = append(centroids, Centroid{Value: mean, Count: int(count)})
			return true
		})
		distributions[idx] = Distribution{Timestamp: bin.timestamp, Centroids: centroids}
	}
	if clean {
		h.priorTimedBinsList = h.priorTimedBinsList[:0]
	}
	return distributions
}

func (h *histogramImpl) rotateCurrentTDigestIfNeedIt() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.currentTimedBin == nil {
		h.currentTimedBin = h.newTimedBin()
	} else if h.currentTimedBin.timestamp != h.now() {
		h.priorTimedBinsList = append(h.priorTimedBinsList, h.currentTimedBin)
		if len(h.priorTimedBinsList) > h.maxBins {
			h.priorTimedBinsList = h.priorTimedBinsList[1:]
		}
		h.currentTimedBin = h.newTimedBin()
	}
}

func (h *histogramImpl) now() time.Time {
	return time.Now().Truncate(h.granularity.Duration())
}

func (h *histogramImpl) newTimedBin() *timedBin {
	td, _ := tdigest.New(tdigest.Compression(h.compression))
	return &timedBin{timestamp: h.now(), tdigest: td}
}
