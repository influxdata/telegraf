package load

import (
	"encoding/json"
)

type LoadAvgStat struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

func (l LoadAvgStat) String() string {
	s, _ := json.Marshal(l)
	return string(s)
}
