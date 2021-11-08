package starlark

import (
	"fmt"

	"github.com/influxdata/telegraf"
	common "github.com/influxdata/telegraf/plugins/common/starlark"
	"github.com/influxdata/telegraf/plugins/processors"
	"go.starlark.net/starlark"
)

const (
	description  = "Process metrics using a Starlark script"
	sampleConfig = `
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.
  ##
  ## Source of the Starlark script.
  source = '''
def apply(metric):
	return metric
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"

  ## The constants of the Starlark script.
  # [processors.starlark.constants]
  #   max_size = 10
  #   threshold = 0.75
  #   default_name = "Julia"
  #   debug_mode = true
`
)

type Starlark struct {
	common.StarlarkCommon

	applyFunc *starlark.Function
	args      starlark.Tuple
	results   []telegraf.Metric
}

func (s *Starlark) Init() error {
	globals, err := s.StarlarkCommon.Init()
	if err != nil {
		return err
	}

	// The source should define an apply function.
	s.applyFunc, err = common.InitFunction(globals, "apply", 1)
	if err != nil {
		return err
	}

	// Reusing the same metric wrapper to skip an allocation.  This will cause
	// any saved references to point to the new metric, but due to freezing the
	// globals none should exist.
	s.args = make(starlark.Tuple, 1)
	s.args[0] = &common.Metric{}

	// Preallocate a slice for return values.
	s.results = make([]telegraf.Metric, 0, 10)

	return nil
}

func (s *Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Description() string {
	return description
}

func (s *Starlark) Start(_ telegraf.Accumulator) error {
	return nil
}

func (s *Starlark) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	s.args[0].(*common.Metric).Wrap(metric)

	rv, err := s.Call(s.applyFunc, s.args)
	if err != nil {
		metric.Reject()
		return err
	}

	switch rv := rv.(type) {
	case *starlark.List:
		iter := rv.Iterate()
		defer iter.Done()
		var v starlark.Value
		for iter.Next(&v) {
			switch v := v.(type) {
			case *common.Metric:
				m := v.Unwrap()
				if containsMetric(s.results, m) {
					s.Log.Errorf("Duplicate metric reference detected")
					continue
				}
				s.results = append(s.results, m)
				acc.AddMetric(m)
			default:
				s.Log.Errorf("Invalid type returned in list: %s", v.Type())
			}
		}

		// If the script didn't return the original metrics, mark it as
		// successfully handled.
		if !containsMetric(s.results, metric) {
			metric.Accept()
		}

		// clear results
		for i := range s.results {
			s.results[i] = nil
		}
		s.results = s.results[:0]
	case *common.Metric:
		m := rv.Unwrap()

		// If the script returned a different metric, mark this metric as
		// successfully handled.
		if m != metric {
			metric.Accept()
		}
		acc.AddMetric(m)
	case starlark.NoneType:
		metric.Drop()
	default:
		return fmt.Errorf("Invalid type returned: %T", rv)
	}
	return nil
}

func (s *Starlark) Stop() error {
	return nil
}

func containsMetric(metrics []telegraf.Metric, metric telegraf.Metric) bool {
	for _, m := range metrics {
		if m == metric {
			return true
		}
	}
	return false
}

func init() {
	processors.AddStreaming("starlark", func() telegraf.StreamingProcessor {
		return &Starlark{
			StarlarkCommon: common.StarlarkCommon{
				StarlarkLoadFunc: common.LoadFunc,
			},
		}
	})
}
