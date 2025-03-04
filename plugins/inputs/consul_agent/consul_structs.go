package consul_agent

type agentInfo struct {
	Timestamp string
	Gauges    []gaugeValue
	Points    []pointValue
	Counters  []sampledValue
	Samples   []sampledValue
}

type gaugeValue struct {
	Name   string
	Value  float32
	Labels map[string]string
}

type pointValue struct {
	Name   string
	Points []float32
}

type sampledValue struct {
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
