//go:generate ../../../tools/readme_config_includer/generator
package starlark

import (
	_ "embed"

	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	common "github.com/influxdata/telegraf/plugins/common/starlark"
)

//go:embed sample.conf
var sampleConfig string

type Starlark struct {
	common.Common
}

func (*Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Init() error {
	// Execute source
	err := s.Common.Init()
	if err != nil {
		return err
	}

	// The source should define an add function.
	err = s.AddFunction("add", &common.Metric{})
	if err != nil {
		return err
	}

	// The source should define a push function.
	err = s.AddFunction("push")
	if err != nil {
		return err
	}

	// The source should define a reset function.
	err = s.AddFunction("reset")
	if err != nil {
		return err
	}

	return nil
}

func (s *Starlark) Add(metric telegraf.Metric) {
	parameters, found := s.GetParameters("add")
	if !found {
		s.Log.Errorf("The parameters of the add function could not be found")
		return
	}
	parameters[0].(*common.Metric).Wrap(metric)

	_, err := s.Call("add")
	if err != nil {
		s.LogError(err)
	}
}

func (s *Starlark) Push(acc telegraf.Accumulator) {
	rv, err := s.Call("push")
	if err != nil {
		s.LogError(err)
		acc.AddError(err)
		return
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
				acc.AddMetric(m)
			default:
				s.Log.Errorf("Invalid type returned in list: %s", v.Type())
			}
		}
	case *common.Metric:
		m := rv.Unwrap()
		acc.AddMetric(m)
	case starlark.NoneType:
	default:
		s.Log.Errorf("Invalid type returned: %T", rv)
	}
}

func (s *Starlark) Reset() {
	_, err := s.Call("reset")
	if err != nil {
		s.LogError(err)
	}
}

// init initializes starlark aggregator plugin
func init() {
	aggregators.Add("starlark", func() telegraf.Aggregator {
		return &Starlark{
			Common: common.Common{
				StarlarkLoadFunc: common.LoadFunc,
			},
		}
	})
}
