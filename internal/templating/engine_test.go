package templating

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineAlternateSeparator(t *testing.T) {
	defaultTemplate, _ := NewDefaultTemplateWithPattern("topic*")
	engine, err := NewEngine("_", defaultTemplate, []string{
		"/ /*/*/* /measurement/origin/measurement*",
	})
	require.NoError(t, err)
	name, tags, field, err := engine.Apply("/telegraf/host01/cpu")
	require.NoError(t, err)
	require.Equal(t, "telegraf_cpu", name)
	require.Equal(t, map[string]string{
		"origin": "host01",
	}, tags)
	require.Equal(t, "", field)
}

func TestEngineWithWildcardTemplate(t *testing.T) {
	var (
		defaultTmpl, err = NewDefaultTemplateWithPattern("measurement*")
		templates        = []string{
			"taskmanagerTask.alarm-detector.Assign.alarmDefinitionId metricsType.process.nodeId.x.alarmDefinitionId.measurement.field rule=1",
			"taskmanagerTask.*.*.*.*                                 metricsType.process.nodeId.measurement rule=2",
		}

		lineOne = "taskmanagerTask.alarm-detector.Assign.alarmDefinitionId.timeout_errors.duration.p75"
		lineTwo = "taskmanagerTask.alarm-detector.Assign.numRecordsInPerSecond.m5_rate"
	)
	require.NoError(t, err)

	engine, err := NewEngine(".", defaultTmpl, templates)
	require.NoError(t, err)

	measurement, tags, field, err := engine.Apply(lineOne)
	require.NoError(t, err)

	assert.Equal(t, "duration", measurement)
	assert.Equal(t, "p75", field)
	assert.Equal(t, map[string]string{
		"metricsType":       "taskmanagerTask",
		"process":           "alarm-detector",
		"nodeId":            "Assign",
		"x":                 "alarmDefinitionId",
		"alarmDefinitionId": "timeout_errors",
		"rule":              "1",
	}, tags)

	measurement, tags, field, err = engine.Apply(lineTwo)
	require.NoError(t, err)

	assert.Equal(t, "numRecordsInPerSecond", measurement)
	assert.Equal(t, "", field)
	assert.Equal(t, map[string]string{
		"metricsType": "taskmanagerTask",
		"process":     "alarm-detector",
		"nodeId":      "Assign",
		"rule":        "2",
	}, tags)
}
