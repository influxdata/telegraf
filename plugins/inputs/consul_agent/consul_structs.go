package consul_agent

type AgentInfo struct {
	Timestamp string
	Gauges    []GaugeValue
	Points    []PointValue
	Counters  []SampledValue
	Samples   []SampledValue
}

type GaugeValue struct {
	Name   string
	Value  float32
	Labels map[string]string
}

type PointValue struct {
	Name   string
	Points []float32
}

type SampledValue struct {
	Name   string
	Count  int
	Sum    float64
	Min    float64
	Max    float64
	Mean   float64
	Rate   float64
	Stddev float64
	Labels map[string]string
}
