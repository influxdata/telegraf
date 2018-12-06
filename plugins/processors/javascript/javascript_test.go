package javascript

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestMetric() telegraf.Metric {
	mtr, _ := metric.New("m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{
			"val":        int64(1),
			"valstring":  "test",
			"valfloat":   float64(1.0),
			"valboolean": false,
		},
		time.Now(),
	)
	return mtr
}

func createJSVM() *otto.Otto {
	return otto.New()
}

func TestSetTags(t *testing.T) {
	jsvm := createJSVM()
	mtr := createTestMetric()

	tagKeys := []string{
		"metric_tag",
	}

	err := setTags(mtr, tagKeys, jsvm)
	require.NoError(t, err)

	_, err = jsvm.Run("")
	require.NoError(t, err)

	val, _ := jsvm.Get("metric_tag")

	tag, _ := val.ToString()
	expected, _ := mtr.GetTag("metric_tag")
	assert.Equal(t, expected, tag)
}

func TestSetFields(t *testing.T) {
	jsvm := createJSVM()
	mtr := createTestMetric()

	tagFields := []string{
		"val",
	}

	err := setFields(mtr, tagFields, jsvm)
	require.NoError(t, err)

	jsvm.Run("")

	val, _ := jsvm.Get("val")

	field, _ := val.ToInteger()
	expected, ok := mtr.GetField("val")
	assert.True(t, ok)
	assert.Equal(t, expected, field)
}

func TestGetTags(t *testing.T) {
	jsvm := createJSVM()
	jsvm.Set("metric_tag", "updated_from_metric")

	tagKeys := []string{
		"metric_tag",
	}

	mtr, err := getTags(createTestMetric(), tagKeys, jsvm)
	require.NoError(t, err)
	require.Implements(t, (*telegraf.Metric)(nil), mtr)

	tag, ok := mtr.GetTag("metric_tag")
	assert.True(t, ok)
	assert.Equal(t, "updated_from_metric", tag)
}

func TestGetFields(t *testing.T) {
	jsvm := createJSVM()
	jsvm.Set("val", int64(10))

	fieldKeys := []*Variable{
		{
			Name:     "val",
			DataType: "integer",
		},
	}

	mtr, err := getFields(createTestMetric(), fieldKeys, jsvm)
	require.NoError(t, err)
	require.Implements(t, (*telegraf.Metric)(nil), mtr)

	field, ok := mtr.GetField("val")
	assert.True(t, ok)
	assert.Equal(t, int64(10), field)
}

func TestApply(t *testing.T) {
	mtr := createTestMetric()

	js := &JavaScript{}
	js.Code = `// javascript
	metric_tag = "new_tag";
	val = 10;
	valstring = "tset";
	valfloat = 10.0;
	valboolean = true;
	`
	js.SetTags = []string{
		"metric_tag",
	}
	js.SetFields = []string{
		"val",
		"valstring",
		"valfloat",
		"valboolean",
	}
	js.GetTags = []string{
		"metric_tag",
	}
	js.GetFields = []*Variable{
		{
			Name:     "val",
			DataType: "integer",
		},
		{
			Name:     "valstring",
			DataType: "string",
		},
		{
			Name:     "valfloat",
			DataType: "float",
		},
		{
			Name:     "valboolean",
			DataType: "boolean",
		},
	}

	mtrs := js.Apply(mtr)
	mtr = mtrs[0]

	tag, ok := mtr.GetTag("metric_tag")
	assert.True(t, ok)
	assert.Equal(t, "new_tag", tag)

	field, ok := mtr.GetField("val")
	assert.True(t, ok)
	assert.Equal(t, int64(10), field)

	field, ok = mtr.GetField("valstring")
	assert.True(t, ok)
	assert.Equal(t, "tset", field)

	field, ok = mtr.GetField("valfloat")
	assert.True(t, ok)
	assert.Equal(t, float64(10.0), field)

	field, ok = mtr.GetField("valboolean")
	assert.True(t, ok)
	assert.Equal(t, true, field)
}

func TestApplyNegative(t *testing.T) {
	mtr := createTestMetric()

	js := &JavaScript{}
	js.Code = `// javascript
	val = "test";
	valboolean = "false";
	`
	js.GetFields = []*Variable{
		{
			Name:     "val",
			DataType: "somethingelse",
		},
		{
			Name:     "undefined",
			DataType: "string",
		},
	}

	mtrs := js.Apply(mtr)
	mtr = mtrs[0]
}

func TestSampleConfig(t *testing.T) {
	js := &JavaScript{}
	assert.Equal(t, sampleConfig, js.SampleConfig())
}

func TestDescription(t *testing.T) {
	js := &JavaScript{}
	assert.Equal(t, "Process values by using JavaScript", js.Description())
}
