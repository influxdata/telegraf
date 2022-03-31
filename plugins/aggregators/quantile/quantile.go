package quantile

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Quantile struct {
	Quantiles     []float64 `toml:"quantiles"`
	Compression   float64   `toml:"compression"`
	AlgorithmType string    `toml:"algorithm"`

	newAlgorithm newAlgorithmFunc

	cache    map[uint64]aggregate
	suffixes []string
}

type aggregate struct {
	name   string
	fields map[string]algorithm
	tags   map[string]string
}

type newAlgorithmFunc func(compression float64) (algorithm, error)

func (q *Quantile) Add(in telegraf.Metric) {
	id := in.HashID()
	if cached, ok := q.cache[id]; ok {
		fields := in.Fields()
		for k, algo := range cached.fields {
			if field, ok := fields[k]; ok {
				if v, isconvertible := convert(field); isconvertible {
					algo.Add(v)
				}
			}
		}
		return
	}

	// New metric, setup cache and init algorithm
	a := aggregate{
		name:   in.Name(),
		tags:   in.Tags(),
		fields: make(map[string]algorithm),
	}
	for k, field := range in.Fields() {
		if v, isconvertible := convert(field); isconvertible {
			// This should never error out as we tested it in Init()
			algo, _ := q.newAlgorithm(q.Compression)
			algo.Add(v)
			a.fields[k] = algo
		}
	}
	q.cache[id] = a
}

func (q *Quantile) Push(acc telegraf.Accumulator) {
	for _, aggregate := range q.cache {
		fields := map[string]interface{}{}
		for k, algo := range aggregate.fields {
			for i, qtl := range q.Quantiles {
				fields[k+q.suffixes[i]] = algo.Quantile(qtl)
			}
		}
		acc.AddFields(aggregate.name, fields, aggregate.tags)
	}
}

func (q *Quantile) Reset() {
	q.cache = make(map[uint64]aggregate)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (q *Quantile) Init() error {
	switch q.AlgorithmType {
	case "t-digest", "":
		q.newAlgorithm = newTDigest
	case "exact R7":
		q.newAlgorithm = newExactR7
	case "exact R8":
		q.newAlgorithm = newExactR8
	default:
		return fmt.Errorf("unknown algorithm type %q", q.AlgorithmType)
	}
	if _, err := q.newAlgorithm(q.Compression); err != nil {
		return fmt.Errorf("cannot create %q algorithm: %v", q.AlgorithmType, err)
	}

	if len(q.Quantiles) == 0 {
		q.Quantiles = []float64{0.25, 0.5, 0.75}
	}

	duplicates := make(map[float64]bool)
	q.suffixes = make([]string, len(q.Quantiles))
	for i, qtl := range q.Quantiles {
		if qtl < 0.0 || qtl > 1.0 {
			return fmt.Errorf("quantile %v out of range", qtl)
		}
		if _, found := duplicates[qtl]; found {
			return fmt.Errorf("duplicate quantile %v", qtl)
		}
		duplicates[qtl] = true
		q.suffixes[i] = fmt.Sprintf("_%03d", int(qtl*100.0))
	}

	q.Reset()

	return nil
}

func init() {
	aggregators.Add("quantile", func() telegraf.Aggregator {
		return &Quantile{Compression: 100}
	})
}
