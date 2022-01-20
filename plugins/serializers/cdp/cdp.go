package cdp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
)

/// The format of the system usage events defined in CDP.
type event struct {
	Timestamp     int       `json:"ts" validate:"required"`
	EventID       string    `json:"eventId" validate:"required"`
	Fingerprint   string    `json:"fingerprint"`
	ServiceID     string    `json:"serviceId" validate:"required"`
	ProjectID     string    `json:"projectId"`
	EnvironmentID string    `json:"environmentId"`
	PlayerID      string    `json:"playerId"`
	StartTime     int       `json:"startTime" validate:"required"`
	EndTime       int       `json:"endTime" validate:"required"`
	Type          string    `json:"type" validate:"required"`
	Amount        float64   `json:"amount" validate:"required"`
	Tags          eventTags `json:"tags"`
}

/// Additional event tags that can be optionally specified.
type eventTags struct {
	MultiplayFleetID   string `json:"multiplayFleetId"`
	MultiplayMachineID string `json:"multiplayMachineId"`
	MultiplayInfraType string `json:"multiplayInfraType"`
	MultiplayProjectID string `json:"multiplayProjectId"`
	MultiplayRegion    string `json:"multiplayRegion"`
	AnalyticsEventType string `json:"analyticsEventType"`
	AnalyticsEventName string `json:"analyticsEventName"`
}

type errMissingTag string

func (e errMissingTag) Error() string {
	return fmt.Sprintf("missing required tag: %q", string(e))
}

/// Serializer is a serializer to the CDP system usage event format.
type Serializer struct{}

func NewSerializer() (*Serializer, error) {
	return &Serializer{}, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	e, err := s.createEvent(metric)
	if err != nil {
		return []byte{}, err
	}

	serialized, err := json.Marshal(e)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')
	return serialized, nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	events := make([]*event, 0, len(metrics))
	for _, metric := range metrics {
		e, err := s.createEvent(metric)
		if err != nil {
			return []byte{}, err
		}
		events = append(events, e)
	}

	serialized, err := json.Marshal(events)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')
	return serialized, nil
}

func (s *Serializer) createEvent(metric telegraf.Metric) (*event, error) {

	// prepare identifying information for the system usage event
	eventID := uuid.Must(uuid.NewV4()).String()
	timestamp := toCdpTimestamp(time.Now())
	endTime := toCdpTimestamp(metric.Time())

	// mapping of required tag/field names to a function which pulls their values from "metric"
	requiredKeys := map[string]func(telegraf.Metric, string) (interface{}, error){
		"billing_region_id":      getTag,
		"customer_id":            getTag,
		"environment_id":         getTag,
		"fleet_id":               getTag,
		"metering_event_machine": getTag,
		"project_id":             getTag,
		"quantity":               getFieldAsAmount,
		"service":                getTag,
		"start_time":             getFieldAsStartTime,
		"virtual_type":           getTag,
	}

	// populate a map with the values of each key
	values := map[string]interface{}{}
	for key, getter := range requiredKeys {
		value, err := getter(metric, key)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}

	// pull the values from the map into the final event
	return &event{
		Timestamp:     timestamp,
		EventID:       eventID,
		Fingerprint:   eventID, // the UUID is sufficient for distinguishing events
		ServiceID:     values["service"].(string),
		ProjectID:     values["project_id"].(string),
		EnvironmentID: values["environment_id"].(string),
		PlayerID:      "",
		StartTime:     values["start_time"].(int),
		EndTime:       endTime,
		Type:          "network_usage_event",
		Amount:        values["quantity"].(float64),
		Tags: eventTags{
			MultiplayFleetID:   values["fleet_id"].(string),
			MultiplayMachineID: values["metering_event_machine"].(string),
			MultiplayInfraType: values["virtual_type"].(string),
			MultiplayProjectID: values["customer_id"].(string),
			MultiplayRegion:    values["billing_region_id"].(string),
			AnalyticsEventType: "",
			AnalyticsEventName: "",
		},
	}, nil
}

// Given a tag name, extracts it from "metric" and returns its value verbatim.
func getTag(metric telegraf.Metric, tagName string) (interface{}, error) {
	value, ok := metric.GetTag(tagName)
	if !ok {
		return nil, fmt.Errorf("missing required tag: %v", tagName)
	}
	return value, nil
}

// Given a field name, extracts it from "metric" and converts it into a value compatible with a CDP "start_time".
func getFieldAsStartTime(metric telegraf.Metric, propertyName string) (interface{}, error) {

	rawStartTime, ok := metric.GetField(propertyName)
	if !ok {
		return 0, fmt.Errorf("missing required field: %v", propertyName)
	}

	startTime, err := time.Parse(time.RFC3339, rawStartTime.(string))
	if err != nil {
		return 0, err
	}

	return toCdpTimestamp(startTime), nil
}

// Given a field name, extracts it from "metric" and converts it into a value compatible with a CDP "amount".
func getFieldAsAmount(metric telegraf.Metric, fieldName string) (interface{}, error) {
	if amount, ok := metric.GetField(fieldName); ok {
		switch i := amount.(type) {
		case float64:
			return i, nil
		case float32:
			return float64(i), nil
		case int64:
			return float64(i), nil
		case int32:
			return float64(i), nil
		case int:
			return float64(i), nil
		default:
			return 0.0, fmt.Errorf("unrecognized type %T (value: %v)", amount, amount)
		}
	}
	return 0.0, nil
}

// Convenience method for converting 'time.Time' to the milliseconds required by CDP.
func toCdpTimestamp(t time.Time) int {
	return int(t.UnixNano() / 1000000)
}
