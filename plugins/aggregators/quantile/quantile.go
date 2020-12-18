package quantile

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Quantile struct {
	Quantiles        []float64 `toml:"quantiles"`
	Compression      float64   `toml:"compression"`
	ApproximatorType string    `toml:"approximator"`

	newApproximator newApproximatorFunc

	cache    map[uint64]aggregate
	suffixes []string
}

type aggregate struct {
	name   string
	fields map[string]approximator
	tags   map[string]string
}

type newApproximatorFunc func(compression float64) (approximator, error)

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

	## Quantiles to output in the range [0,1]
	# quantiles = [0.25, 0.5, 0.75]

	## Type of approximation
	## Supported are:
	##	"t-digest" -- approximation using centroids, can cope with large number of samples
	##  "exact R7" -- exact approximation also used by Excel or NumPy (Hyndman & Fan 1996 R7)
	##  "exact R8" -- exact approximation (Hyndman & Fan 1996 R8)
	## NOTE: Do not use "exact" approximations with large number of samples
	##       to not impair performance or memory consumption!
	# approximator = "t-digest"

	## Compression for approximation (t-digest). The value needs to be
	## greater or equal to 1.0. Smaller values will result in more
	## performance but less accuracy.
	# compression = 100.0
`

func (q *Quantile) SampleConfig() string {
	return sampleConfig
}

func (q *Quantile) Description() string {
	return "Keep the aggregate quantiles of each metric passing through."
}

func (q *Quantile) Add(in telegraf.Metric) {
	id := in.HashID()
	if cached, ok := q.cache[id]; ok {
		fields := in.Fields()
		for k, approx := range cached.fields {
			if field, ok := fields[k]; ok {
				if v, isconvertible := convert(field); isconvertible {
					approx.Add(v)
				}
			}
		}
		return
	}

	// New metric, setup cache and init approximator
	a := aggregate{
		name:   in.Name(),
		tags:   in.Tags(),
		fields: make(map[string]approximator),
	}
	for k, field := range in.Fields() {
		if v, isconvertible := convert(field); isconvertible {
			// This should never error out as we tested it in Init()
			approx, _ := q.newApproximator(q.Compression)
			approx.Add(v)
			a.fields[k] = approx
		}
	}
	q.cache[id] = a
}

func (q *Quantile) Push(acc telegraf.Accumulator) {
	for _, aggregate := range q.cache {
		fields := map[string]interface{}{}
		for k, approx := range aggregate.fields {
			for i, qtl := range q.Quantiles {
				fields[k+q.suffixes[i]] = approx.Quantile(qtl)
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
	switch q.ApproximatorType {
	case "t-digest", "":
		q.newApproximator = newTDigest
	case "exact R7":
		q.newApproximator = newExactR7
	case "exact R8":
		q.newApproximator = newExactR8
	default:
		return fmt.Errorf("unknown approximator type %q", q.ApproximatorType)
	}
	if _, err := q.newApproximator(q.Compression); err != nil {
		return fmt.Errorf("cannot create %q approximator: %v", q.ApproximatorType, err)
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
