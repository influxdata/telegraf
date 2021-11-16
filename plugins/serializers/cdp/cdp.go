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
	Timestamp        int       `json:"ts" validate:"required"`
	EventID          string    `json:"eventId" validate:"required"`
	ServiceID        string    `json:"serviceId" validate:"required"`
	ProjectID        string    `json:"projectId"`
	ProjectGenesisID string    `json:"projectGenesisId" validate:"required"`
	EnvironmentID    string    `json:"environmentId"`
	Region           string    `json:"region"`
	StartTime        int       `json:"startTime" validate:"required"`
	EndTime          int       `json:"endTime" validate:"required"`
	Type             string    `json:"type" validate:"required"`
	Amount           float64   `json:"amount" validate:"required"`
	Tags             eventTags `json:"tags"`
}

/// Additional event tags that can be optionally specified.
type eventTags struct {
	FleedID   string `json:"fleetId"`
	MachineID string `json:"machineId"`
	InfraType string `json:"infraType"`
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

	// extract the event properties
	serviceID, ok := metric.GetTag("service")
	if !ok {
		return nil, errMissingTag("service")
	}

	projectGenesisID, ok := metric.GetTag("project_id")
	if !ok {
		return nil, errMissingTag("project_id")
	}

	environmentID, ok := metric.GetTag("environment_id")
	if !ok {
		return nil, errMissingTag("environment_id")
	}

	region, ok := metric.GetTag("billing_region_id")
	if !ok {
		return nil, errMissingTag("billing_region_id")
	}

	startTime, err := getStartTime(metric)
	if err != nil {
		return nil, err
	}

	amount, err := getAmount(metric)
	if err != nil {
		return nil, err
	}

	fleetID, ok := metric.GetTag("fleet_id")
	if !ok {
		return nil, errMissingTag("fleet_id")
	}

	machineID, ok := metric.GetTag("metering_event_machine")
	if !ok {
		return nil, errMissingTag("metering_event_machine")
	}

	infraType, ok := metric.GetTag("virtual_type")
	if !ok {
		return nil, errMissingTag("virtual_type")
	}

	// convert the properties extracted to the CDP structure
	e := &event{
		Timestamp:        timestamp,
		EventID:          eventID,
		ServiceID:        serviceID,
		ProjectID:        "",
		ProjectGenesisID: projectGenesisID,
		EnvironmentID:    environmentID,
		Region:           region,
		StartTime:        startTime,
		EndTime:          endTime,
		Type:             "egress",
		Amount:           amount,
		Tags: eventTags{
			FleedID:   fleetID,
			MachineID: machineID,
			InfraType: infraType,
		},
	}
	return e, nil
}

func getStartTime(metric telegraf.Metric) (int, error) {

	rawStartTime, ok := metric.GetField("start_time")
	if !ok {
		return 0, fmt.Errorf("missing required field: start_time")
	}

	startTime, err := time.Parse(time.RFC3339, rawStartTime.(string))
	if err != nil {
		return 0, err
	}

	return toCdpTimestamp(startTime), nil
}

func getAmount(metric telegraf.Metric) (float64, error) {

	quantity, ok := metric.GetField("quantity_total")
	if !ok {
		return 0, fmt.Errorf("missing required field: quantity_total")
	}

	switch i := quantity.(type) {
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
		return 0, fmt.Errorf("unrecognized type %T (value: %v)", quantity, quantity)
	}
}

/// Convenience method for converting 'time.Time' to the milliseconds required by CDP.
func toCdpTimestamp(t time.Time) int {
	return int(t.UnixNano() / 1000000)
}
