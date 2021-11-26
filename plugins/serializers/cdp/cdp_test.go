package cdp

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

const (
	serviceID     = "service-id"
	projectID     = "project-genesis-id"
	environmentID = "environment-id"
	region        = "region"
	bytesSent     = 1234567.0
	fleetID       = "fleet-id"
	machineID     = "metering-event-machine"
	infraType     = "infra-type"
	mpProjectID   = "customer-id"
	metricType    = "egress"
)

func getValidTime() time.Time {
	return time.Now()
}

func getValidTags() map[string]string {
	return map[string]string{
		"service":                serviceID,
		"project_id":             projectID,
		"environment_id":         environmentID,
		"billing_region_id":      region,
		"fleet_id":               fleetID,
		"metering_event_machine": machineID,
		"virtual_type":           infraType,
		"customer_id":            mpProjectID,
		"host":                   "host-name",
		"interface":              "interface-name",
	}
}

func getValidFields(t time.Time) map[string]interface{} {
	return map[string]interface{}{
		"bytes_sent_sum": bytesSent,
		"start_time":     t.Format(time.RFC3339Nano),
	}
}

func Test_Serialize_HappyPath(t *testing.T) {

	metricTime := getValidTime()
	eventTime := metricTime.Add(time.Second * -30)
	tags := getValidTags()
	fields := getValidFields(eventTime)

	s, err := NewSerializer()
	assert.NoError(t, err)
	m, err := metric.New("net", tags, fields, metricTime)
	assert.NoError(t, err)

	buf, err := s.Serialize(m)
	assert.NoError(t, err)
	cdpEventPayload := &eventPayload{}
	json.Unmarshal(buf, cdpEventPayload)

	expectedPayload := &eventPayload{
		Type: "unity.services.systemUsage.v1",
		Message: event{
			Timestamp:     cdpEventPayload.Message.Timestamp,
			EventID:       cdpEventPayload.Message.EventID,
			Fingerprint:   cdpEventPayload.Message.Fingerprint,
			ServiceID:     serviceID,
			ProjectID:     projectID,
			EnvironmentID: environmentID,
			PlayerID:      "",
			Region:        region,
			StartTime:     toCdpTimestamp(eventTime),
			EndTime:       toCdpTimestamp(metricTime),
			Type:          metricType,
			Amount:        bytesSent,
			Tags: eventTags{
				MultiplayFleetID:   fleetID,
				MultiplayMachineID: machineID,
				MultiplayInfraType: infraType,
				MultiplayProjectID: mpProjectID,
				AnalyticsEventType: "",
				AnalyticsEventName: "",
			},
		},
	}
	assert.Equal(t, expectedPayload, cdpEventPayload)
}

// Tests a variety of error scenarios where specific tags or fields are missing from the Telegraf metric.
func Test_Serialize_MissingTags(t *testing.T) {

	tests := []struct {
		tagOrFieldName string // the metric tag/field to delete (to simulate it not being present)
		isTag          bool   // true = tag, false = field
	}{
		{tagOrFieldName: "service", isTag: true},
		{tagOrFieldName: "project_id", isTag: true},
		{tagOrFieldName: "start_time", isTag: false},
		{tagOrFieldName: "bytes_sent_sum", isTag: false},
		{tagOrFieldName: "billing_region_id", isTag: true},
		{tagOrFieldName: "environment_id", isTag: true},
		{tagOrFieldName: "fleet_id", isTag: true},
		{tagOrFieldName: "metering_event_machine", isTag: true},
		{tagOrFieldName: "virtual_type", isTag: true},
		{tagOrFieldName: "customer_id", isTag: true},
	}

	for _, test := range tests {

		testName := fmt.Sprintf("missing '%v' property should fail serialization", test.tagOrFieldName)
		t.Run(testName, func(t *testing.T) {

			// prepare valid tags and fields for the event
			metricTime := getValidTime()
			eventTime := metricTime.Add(time.Second * -30)
			tags := getValidTags()
			fields := getValidFields(eventTime)

			// remove the tag/field
			if test.isTag {
				delete(tags, test.tagOrFieldName)
			} else {
				delete(fields, test.tagOrFieldName)
			}

			s, err := NewSerializer()
			assert.NoError(t, err)
			m, err := metric.New("net", tags, fields, metricTime)
			assert.NoError(t, err)

			// ensure an error is returned
			_, err = s.Serialize(m)
			assert.True(t, err != nil)
		})
	}
}

func Test_Serialize_InvalidAmount(t *testing.T) {

	metricTime := getValidTime()
	eventTime := metricTime.Add(time.Second * -30)
	tags := getValidTags()
	fields := getValidFields(eventTime)
	fields["bytes_sent_sum"] = "hello" // invalid

	s, err := NewSerializer()
	assert.NoError(t, err)
	m, err := metric.New("net", tags, fields, metricTime)
	assert.NoError(t, err)

	_, err = s.Serialize(m)
	assert.Error(t, err)
}

func Test_SerializeBatch_HappyPath(t *testing.T) {

	metricTime1 := getValidTime()
	metricTime2 := getValidTime()
	eventTime1 := metricTime1.Add(time.Second * -30)
	eventTime2 := metricTime2.Add(time.Second * -30)
	tags := getValidTags()
	fields1 := getValidFields(eventTime1)
	fields2 := getValidFields(eventTime2)

	s, err := NewSerializer()
	assert.NoError(t, err)
	m1, err := metric.New("net", tags, fields1, metricTime1)
	assert.NoError(t, err)
	m2, err := metric.New("net", tags, fields2, metricTime2)
	assert.NoError(t, err)

	buf, err := s.SerializeBatch([]telegraf.Metric{m1, m2})
	assert.NoError(t, err)
	cdpEvents := []event{}
	json.Unmarshal(buf, &cdpEvents)

	expectedEvent1 := &event{
		Timestamp:     cdpEvents[0].Timestamp,
		EventID:       cdpEvents[0].EventID,
		Fingerprint:   cdpEvents[0].Fingerprint,
		ServiceID:     serviceID,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		PlayerID:      "",
		Region:        region,
		StartTime:     toCdpTimestamp(eventTime1),
		EndTime:       toCdpTimestamp(metricTime1),
		Type:          metricType,
		Amount:        bytesSent,
		Tags: eventTags{
			MultiplayFleetID:   fleetID,
			MultiplayMachineID: machineID,
			MultiplayInfraType: infraType,
			MultiplayProjectID: mpProjectID,
			AnalyticsEventType: "",
			AnalyticsEventName: "",
		},
	}
	expectedEvent2 := &event{
		Timestamp:     cdpEvents[1].Timestamp,
		EventID:       cdpEvents[1].EventID,
		Fingerprint:   cdpEvents[1].Fingerprint,
		ServiceID:     serviceID,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		PlayerID:      "",
		Region:        region,
		StartTime:     toCdpTimestamp(eventTime2),
		EndTime:       toCdpTimestamp(metricTime2),
		Type:          metricType,
		Amount:        bytesSent,
		Tags: eventTags{
			MultiplayFleetID:   fleetID,
			MultiplayMachineID: machineID,
			MultiplayInfraType: infraType,
			MultiplayProjectID: mpProjectID,
			AnalyticsEventType: "",
			AnalyticsEventName: "",
		},
	}
	expectedEvents := []event{*expectedEvent1, *expectedEvent2}
	assert.Equal(t, expectedEvents, cdpEvents)
}
