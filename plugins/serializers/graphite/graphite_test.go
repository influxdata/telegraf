package graphite

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf"
)

func TestGraphiteTags(t *testing.T) {
	m1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	tags1 := buildTags(m1)
	tags2 := buildTags(m2)
	tags3 := buildTags(m3)

	assert.Equal(t, "192_168_0_1", tags1)
	assert.Equal(t, "192_168_0_1.first.second", tags2)
	assert.Equal(t, "first.second", tags3)
}

func TestSerializeMetricNoHost(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
		"usage_busy": float64(8.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("cpu0.us-west-2.cpu.usage_idle 91.5 %d", now.Unix()),
		fmt.Sprintf("cpu0.us-west-2.cpu.usage_busy 8.5 %d", now.Unix()),
	}
	sort.Strings(mS)
	sort.Strings(expS)
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricHost(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
		"usage_busy": float64(8.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu.usage_idle 91.5 %d", now.Unix()),
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu.usage_busy 8.5 %d", now.Unix()),
	}
	sort.Strings(mS)
	sort.Strings(expS)
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricPrefix(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
		"usage_busy": float64(8.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Prefix: "prefix"}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("prefix.localhost.cpu0.us-west-2.cpu.usage_idle 91.5 %d", now.Unix()),
		fmt.Sprintf("prefix.localhost.cpu0.us-west-2.cpu.usage_busy 8.5 %d", now.Unix()),
	}
	sort.Strings(mS)
	sort.Strings(expS)
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNameNoHost(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS := s.SerializeBucketName(m, "usage_idle")

	expS := fmt.Sprintf("cpu0.us-west-2.cpu.usage_idle")
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNameHost(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS := s.SerializeBucketName(m, "usage_idle")

	expS := fmt.Sprintf("localhost.cpu0.us-west-2.cpu.usage_idle")
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNamePrefix(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Prefix: "prefix"}
	mS := s.SerializeBucketName(m, "usage_idle")

	expS := fmt.Sprintf("prefix.localhost.cpu0.us-west-2.cpu.usage_idle")
	assert.Equal(t, expS, mS)
}
