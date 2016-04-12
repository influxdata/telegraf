package graphite

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf"
)

var defaultTags = map[string]string{
	"host":       "localhost",
	"cpu":        "cpu0",
	"datacenter": "us-west-2",
}

const (
	template1 = "tags.measurement.field"
	template2 = "host.measurement.field"
	template3 = "host.tags.field"
	template4 = "host.tags.measurement"
	// this template explicitly uses all tag keys, so "tags" should be empty
	template5 = "host.datacenter.cpu.tags.measurement.field"
	// this template has non-existent tag keys
	template6 = "foo.host.cpu.bar.tags.measurement.field"
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

	tags1 := buildTags(m1.Tags())
	tags2 := buildTags(m2.Tags())
	tags3 := buildTags(m3.Tags())

	assert.Equal(t, "192_168_0_1", tags1)
	assert.Equal(t, "first.second.192_168_0_1", tags2)
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

// test that a field named "value" gets ignored.
func TestSerializeValueField(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"value": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu 91.5 %d", now.Unix()),
	}
	assert.Equal(t, expS, mS)
}

// test that a field named "value" gets ignored in middle of template.
func TestSerializeValueField2(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"value": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.field.tags.measurement",
	}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu 91.5 %d", now.Unix()),
	}
	assert.Equal(t, expS, mS)
}

// test that a field named "value" gets ignored at beginning of template.
func TestSerializeValueField3(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"value": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "field.host.tags.measurement",
	}
	mS, err := s.Serialize(m)
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu 91.5 %d", now.Unix()),
	}
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
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNameHost(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNamePrefix(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Prefix: "prefix"}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "prefix.localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate1(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template1}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "cpu0.us-west-2.localhost.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate2(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template2}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate3(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template3}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.cpu0.us-west-2.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate4(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template4}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.cpu0.us-west-2.cpu"
	assert.Equal(t, expS, mS)
}

func TestTemplate5(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template5}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.us-west-2.cpu0.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate6(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := telegraf.NewMetric("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Template: template6}
	mS := s.SerializeBucketName(m.Name(), m.Tags())

	expS := "localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}
