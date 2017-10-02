package particle

import (
	"time"
)

type DummyData struct {
	Event       string `json:"event"`
	Data        string `json:"data"`
	Ttl         int    `json:"ttl"`
	PublishedAt string `json:"published_at"`
	InfluxDB    string `json:"influx_db"`
}
type ParticleData struct {
	Event  string                 `json:"event"`
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"values"`
}

func (d *DummyData) Time() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", d.PublishedAt)
}
