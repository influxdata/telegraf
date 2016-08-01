package particle

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
)

const meas = "particle"

type Event interface {
	NewMetric() telegraf.Metric
}

type ParticleEvent struct {
	Event       string    `schema:"event"`
	Data        int       `schema:"data"`
	PublishedAt time.Time `schema:"published_at"`
	CoreID      string    `schema:"coreid"`
}

func (pe ParticleEvent) String() string {
	return fmt.Sprintf(`
  Event == {
    event: %v,
    data: %v,
    published: %v,
    coreid: %v
  }`,
		pe.Event,
		pe.Data,
		pe.PublishedAt,
		pe.CoreID)
}

func (pe ParticleEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"event":  pe.Event,
		"coreid": pe.CoreID,
	}
	f := map[string]interface{}{
		"data": pe.Data,
	}
	m, err := telegraf.NewMetric(pe.Event, t, f, pe.PublishedAt)
	if err != nil {
		log.Fatalf("Failed to create %v event", meas)
	}
	return m
}
