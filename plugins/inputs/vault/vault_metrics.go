package vault

type sysMetrics struct {
	Timestamp string         `json:"timestamp"`
	Gauges    []gauge        `json:"Gauges"`
	Counters  []sampledValue `json:"Counters"`
	Summaries []sampledValue `json:"Samples"`
}

type baseInfo struct {
	Name   string         `json:"Name"`
	Labels map[string]any `json:"Labels"`
}

type gauge struct {
	baseInfo
	Value float64 `json:"Value"`
}

type sampledValue struct {
	baseInfo
	Count  int     `json:"Count"`
	Rate   float64 `json:"Rate"`
	Sum    float64 `json:"Sum"`
	Min    float64 `json:"Min"`
	Max    float64 `json:"Max"`
	Mean   float64 `json:"Mean"`
	Stddev float64 `json:"Stddev"`
}
