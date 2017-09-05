package graphite

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/metric"
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
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3, _ := metric.New(
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.field.tags.measurement",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu 91.5 %d", now.Unix()),
	}
	assert.Equal(t, expS, mS)
}

func TestSerializeValueString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"value": "asdasd",
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.field.tags.measurement",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)
	assert.Equal(t, "", mS[0])
}

func TestSerializeValueBoolean(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"enabled":  true,
		"disabled": false,
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.field.tags.measurement",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.enabled.cpu0.us-west-2.cpu 1 %d", now.Unix()),
		fmt.Sprintf("localhost.disabled.cpu0.us-west-2.cpu 0 %d", now.Unix()),
	}
	sort.Strings(mS)
	sort.Strings(expS)
	assert.Equal(t, expS, mS)
}

// test that fields with spaces get fixed.
func TestSerializeFieldWithSpaces(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		`field\ with\ spaces`: float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.tags.measurement.field",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu.field_with_spaces 91.5 %d", now.Unix()),
	}
	assert.Equal(t, expS, mS)
}

// test that tags with spaces get fixed.
func TestSerializeTagWithSpaces(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        `cpu\ 0`,
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		`field_with_spaces`: float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "host.tags.measurement.field",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu_0.us-west-2.cpu.field_with_spaces 91.5 %d", now.Unix()),
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: "field.host.tags.measurement",
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.cpu0.us-west-2.cpu 91.5 %d", now.Unix()),
	}
	assert.Equal(t, expS, mS)
}

// test that a field named "value" gets ignored at beginning of template.
func TestSerializeValueField5(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}
	fields := map[string]interface{}{
		"value": float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{
		Template: template5,
	}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{
		fmt.Sprintf("localhost.us-west-2.cpu0.cpu 91.5 %d", now.Unix()),
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := GraphiteSerializer{Prefix: "prefix"}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), "", "")

	expS := "cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNameHost(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), "", "")

	expS := "localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestSerializeBucketNamePrefix(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), "", "prefix")

	expS := "prefix.localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate1(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), template1, "")

	expS := "cpu0.us-west-2.localhost.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate2(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), template2, "")

	expS := "localhost.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate3(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), template3, "")

	expS := "localhost.cpu0.us-west-2.FIELDNAME"
	assert.Equal(t, expS, mS)
}

func TestTemplate4(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), template4, "")

	expS := "localhost.cpu0.us-west-2.cpu"
	assert.Equal(t, expS, mS)
}

func TestTemplate6(t *testing.T) {
	now := time.Now()
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", defaultTags, fields, now)
	assert.NoError(t, err)

	mS := SerializeBucketName(m.Name(), m.Tags(), template6, "")

	expS := "localhost.cpu0.us-west-2.cpu.FIELDNAME"
	assert.Equal(t, expS, mS)
}
