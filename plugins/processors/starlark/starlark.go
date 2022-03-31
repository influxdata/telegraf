package starlark

import (
	"fmt"

	"github.com/influxdata/telegraf"
	common "github.com/influxdata/telegraf/plugins/common/starlark"
	"github.com/influxdata/telegraf/plugins/processors"
	"go.starlark.net/starlark"
)

type Starlark struct {
	common.StarlarkCommon

	results []telegraf.Metric
}

func (s *Starlark) Init() error {
	err := s.StarlarkCommon.Init()
	if err != nil {
		return err
	}

	// The source should define an apply function.
	err = s.AddFunction("apply", &common.Metric{})
	if err != nil {
		return err
	}

	// Preallocate a slice for return values.
	s.results = make([]telegraf.Metric, 0, 10)

	return nil
}

func (s *Starlark) Start(_ telegraf.Accumulator) error {
	return nil
}

func (s *Starlark) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	parameters, found := s.GetParameters("apply")
	if !found {
		return fmt.Errorf("The parameters of the apply function could not be found")
	}
	parameters[0].(*common.Metric).Wrap(metric)

	rv, err := s.Call("apply")
	if err != nil {
		s.LogError(err)
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
