package ubazaar

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
)

type errMissingTag string

func (e errMissingTag) Error() string {
	return fmt.Sprintf("missing required tag: %q", string(e))
}

type serializer struct{}

type event struct {
	EventID           string            `json:"eventID"`
	ServiceCustomerID string            `json:"serviceCustomerID"`
	Service           string            `json:"service"`
	UnitOfMeasure     string            `json:"unitOfMeasure"`
	Quantity          float64           `json:"quantity"`
	StartTime         string            `json:"startTime"`
	EndTime           string            `json:"endTime"`
	MetaData          map[string]string `json:"metadata"`
}

func NewSerializer() (*serializer, error) {
	return &serializer{}, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	e, err := s.createObject(metric)
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

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		e, err := s.createObject(metric)
		if err != nil {
			return []byte{}, err
		}
		objects = append(objects, e)
	}

	obj := map[string]interface{}{
		"metrics": objects,
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) (*event, error) {
	eventID := uuid.Must(uuid.NewV4())
	service, ok := metric.GetTag("service")
	if !ok {
		return nil, errMissingTag("service")
	}

	customerID, ok := metric.GetTag("customer_id")
	if !ok {
		return nil, errMissingTag("customer_id")
	}

	unitOfMeasure, ok := metric.GetTag("unit_of_measure")
	if !ok {
		return nil, errMissingTag("unit_of_measure")
	}

	startTime, ok := metric.GetField("start_time")
	if !ok {
		return nil, errors.New("missing required field: start_time")
	}

	filteredTags := make(map[string]string)
	for k, v := range metric.Tags() {
		switch k {
		case "service", "customer_id", "unit_of_measure":
			continue
		}
		filteredTags[k] = v
	}

	e := &event{
		EventID:           eventID.String(),
		ServiceCustomerID: customerID,
		Service:           service,
		UnitOfMeasure:     unitOfMeasure,
		Quantity:          getQuantity(metric),
		StartTime:         startTime.(string),
		EndTime:           metric.Time().Format(time.RFC3339),
		MetaData:          filteredTags,
	}

	return e, nil
}

func getQuantity(metric telegraf.Metric) float64 {
	if field, ok := metric.GetField("quantity"); ok {
		switch i := field.(type) {
		case float64:
			return i
		case float32:
			return float64(i)
		case int64:
			return float64(i)
		case int32:
			return float64(i)
		case int:
			return float64(i)
		}
	}
	return 0
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
