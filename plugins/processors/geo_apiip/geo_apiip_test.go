package geo_apiip

import (
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func createTestMetric() telegraf.Metric {
	m := metric.New("m1",
		map[string]string{"source_ip": "193.200.168.1"},
		map[string]interface{}{"value": int64(1)},
		time.Now(),
	)
	return m
}

func calculateProcessedTags(processor GeoAPI, m telegraf.Metric) map[string]string {
	processed := processor.Apply(m)
	return processed[0].Tags()
}

func TestGetLocation(t *testing.T) {
	p := New()
	p.APIKey = os.Getenv("APIIP_API_KEY")
	if p.APIKey == "" {
		t.Skip("APIIP_API_KEY not set")
	}
	p.Log = &testutil.Logger{}
	go p.cache.Start()

	l, err := p.getLocation("69.6.30.203")

	require.NoError(t, err)
	require.Equal(t, "Europe", l.Region)
	require.Equal(t, "CY", l.Country)
	require.Equal(t, "Nicosia", l.City)
}

func TestAddTagsIP(t *testing.T) {
	p := New()
	p.APIKey = os.Getenv("APIIP_API_KEY")
	if p.APIKey == "" {
		t.Skip("APIIP_API_KEY not set")
	}
	p.IP = "69.6.30.203"
	p.Log = &testutil.Logger{}
	go p.cache.Start()

	tags := calculateProcessedTags(*p, createTestMetric())

	r, rp := tags["region"]
	c, cp := tags["country"]
	ci, cip := tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, "Europe", r, "Value of Tag was changed")
	require.Equal(t, "CY", c, "Value of Tag was changed")
	require.Equal(t, "Nicosia", ci, "Value of Tag was changed")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")
}

func TestAddTagsOrigin(t *testing.T) {
	p := New()
	p.APIKey = os.Getenv("APIIP_API_KEY")
	if p.APIKey == "" {
		t.Skip("APIIP_API_KEY not set")
	}
	p.Log = &testutil.Logger{}
	go p.cache.Start()

	tags := calculateProcessedTags(*p, createTestMetric())

	_, rp := tags["region"]
	_, cp := tags["country"]
	_, cip := tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")
}

func TestAddTagsIPTag(t *testing.T) {
	p := New()
	p.APIKey = os.Getenv("APIIP_API_KEY")
	if p.APIKey == "" {
		t.Skip("APIIP_API_KEY not set")
	}
	p.IPTag = "source_ip"
	p.Log = &testutil.Logger{}
	go p.cache.Start()

	tags := calculateProcessedTags(*p, createTestMetric())

	r, rp := tags["region"]
	c, cp := tags["country"]
	ci, cip := tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, "Europe", r, "Value of Tag was changed")
	require.Equal(t, "RU", c, "Value of Tag was changed")
	require.Equal(t, "Moscow", ci, "Value of Tag was changed")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")
}

func TestAddTagsIPCache(t *testing.T) {
	p := New()
	p.APIKey = os.Getenv("APIIP_API_KEY")
	if p.APIKey == "" {
		t.Skip("APIIP_API_KEY not set")
	}
	p.IP = "69.6.30.203"
	p.Log = &testutil.Logger{}
	p.UpdateInterval = config.Duration(time.Second * time.Duration(5))
	go p.cache.Start()

	tags := calculateProcessedTags(*p, createTestMetric())

	r, rp := tags["region"]
	c, cp := tags["country"]
	ci, cip := tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, "Europe", r, "Value of Tag was changed")
	require.Equal(t, "CY", c, "Value of Tag was changed")
	require.Equal(t, "Nicosia", ci, "Value of Tag was changed")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")

	tags = calculateProcessedTags(*p, createTestMetric())

	r, rp = tags["region"]
	c, cp = tags["country"]
	ci, cip = tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, "Europe", r, "Value of Tag was changed")
	require.Equal(t, "CY", c, "Value of Tag was changed")
	require.Equal(t, "Nicosia", ci, "Value of Tag was changed")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")

	time.Sleep(time.Second * time.Duration(6))

	tags = calculateProcessedTags(*p, createTestMetric())

	r, rp = tags["region"]
	c, cp = tags["country"]
	ci, cip = tags["city"]
	require.True(t, rp, "Region Tag of metric was not present")
	require.True(t, cp, "Country Tag of metric was not present")
	require.True(t, cip, "City Tag of metric was not present")
	require.Equal(t, "Europe", r, "Value of Tag was changed")
	require.Equal(t, "CY", c, "Value of Tag was changed")
	require.Equal(t, "Nicosia", ci, "Value of Tag was changed")
	require.Equal(t, 4, len(tags), "Should have one previous and three added tags.")

}
