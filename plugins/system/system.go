package system

import "github.com/influxdb/tivan/plugins"

type SystemStats struct {
	ps PS
}

func (_ *SystemStats) Description() string {
	return "Read metrics about system load"
}

func (_ *SystemStats) SampleConfig() string { return "" }

func (s *SystemStats) add(acc plugins.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

func (s *SystemStats) Gather(acc plugins.Accumulator) error {
	lv, err := s.ps.LoadAvg()
	if err != nil {
		return err
	}

	acc.Add("load1", lv.Load1, nil)
	acc.Add("load5", lv.Load5, nil)
	acc.Add("load15", lv.Load15, nil)

	return nil
}

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{ps: &systemPS{}}
	})
}
