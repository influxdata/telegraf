package vault

type SysMetrics struct {
	Timestamp string    `json:"timestamp"`
	Gauges    []gauge   `json:"Gauges"`
	Counters  []counter `json:"Counters"`
	Summaries []summary `json:"Samples"`
}

type baseInfo struct {
	Name   string                 `json:"Name"`
	Labels map[string]interface{} `json:"Labels"`
}

type gauge struct {
	baseInfo
	Value int `json:"Value"`
}

type counter struct {
	baseInfo
	Count  int     `json:"Count"`
	Rate   float64 `json:"Rate"`
	Sum    int     `json:"Sum"`
	Min    int     `json:"Min"`
	Max    int     `json:"Max"`
	Mean   float64 `json:"Mean"`
	Stddev float64 `json:"Stddev"`
}

type summary struct {
	baseInfo
	Count  int     `json:"Count"`
	Rate   float64 `json:"Rate"`
	Sum    float64 `json:"Sum"`
	Min    float64 `json:"Min"`
	Max    float64 `json:"Max"`
	Mean   float64 `json:"Mean"`
	Stddev float64 `json:"Stddev"`
}
