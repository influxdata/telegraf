//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"fmt"
	"time"

	ia "github.com/intel/iaevents"
	"golang.org/x/sync/errgroup"
)

type coreMetric struct {
	values ia.CounterValue
	scaled uint64

	name string
	tag  string
	cpu  int

	time time.Time
}

type uncoreMetric struct {
	values ia.CounterValue
	scaled uint64

	name     string
	unitType string
	unit     string
	tag      string
	socket   int

	agg bool

	time time.Time
}

type valuesReader interface {
	readValue(event *ia.ActiveEvent) (ia.CounterValue, error)
}

type iaValuesReader struct{}

func (iaValuesReader) readValue(event *ia.ActiveEvent) (ia.CounterValue, error) {
	return event.ReadValue()
}

type entitiesValuesReader interface {
	readEntities([]*CoreEventEntity, []*UncoreEventEntity) ([]coreMetric, []uncoreMetric, error)
}

type iaEntitiesValuesReader struct {
	eventReader valuesReader
	timer       clock
}

type clock interface {
	now() time.Time
}

type realClock struct{}

func (realClock) now() time.Time {
	return time.Now()
}

func (ie *iaEntitiesValuesReader) readEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) ([]coreMetric, []uncoreMetric, error) {
	var coreMetrics []coreMetric
	var uncoreMetrics []uncoreMetric

	for _, entity := range coreEntities {
		newMetrics, err := ie.readCoreEvents(entity)
		if err != nil {
			return nil, nil, err
		}
		coreMetrics = append(coreMetrics, newMetrics...)
	}
	for _, entity := range uncoreEntities {
		newMetrics, err := ie.readUncoreEvents(entity)
		if err != nil {
			return nil, nil, err
		}
		uncoreMetrics = append(uncoreMetrics, newMetrics...)
	}
	return coreMetrics, uncoreMetrics, nil
}

func (ie *iaEntitiesValuesReader) readCoreEvents(entity *CoreEventEntity) ([]coreMetric, error) {
	if ie.eventReader == nil || ie.timer == nil {
		return nil, fmt.Errorf("event values reader or timer is nil")
	}
	if entity == nil {
		return nil, fmt.Errorf("entity is nil")
	}
	metrics := make([]coreMetric, len(entity.activeEvents))
	errGroup := errgroup.Group{}

	for i, event := range entity.activeEvents {
		id := i
		actualEvent := event

		if event == nil || event.PerfEvent == nil {
			return nil, fmt.Errorf("active event or corresponding perf event is nil")
		}

		errGroup.Go(func() error {
			values, err := ie.eventReader.readValue(actualEvent)
			if err != nil {
				return fmt.Errorf("failed to read core event `%s` values: %v", actualEvent, err)
			}
			cpu, _ := actualEvent.PMUPlacement()
			newMetric := coreMetric{
				values: values,
				tag:    entity.EventsTag,
				cpu:    cpu,
				name:   actualEvent.PerfEvent.Name,
				time:   ie.timer.now(),
			}
			metrics[id] = newMetric
			return nil
		})
	}
	err := errGroup.Wait()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (ie *iaEntitiesValuesReader) readUncoreEvents(entity *UncoreEventEntity) ([]uncoreMetric, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity is nil")
	}
	var uncoreMetrics []uncoreMetric

	for _, event := range entity.activeMultiEvents {
		if entity.Aggregate {
			newMetric, err := ie.readMultiEventAgg(event)
			if err != nil {
				return nil, err
			}
			newMetric.tag = entity.EventsTag
			uncoreMetrics = append(uncoreMetrics, newMetric)
		} else {
			newMetrics, err := ie.readMultiEventSeparately(event)
			if err != nil {
				return nil, err
			}
			for i := range newMetrics {
				newMetrics[i].tag = entity.EventsTag
			}
			uncoreMetrics = append(uncoreMetrics, newMetrics...)
		}
	}
	return uncoreMetrics, nil
}

func (ie *iaEntitiesValuesReader) readMultiEventSeparately(multiEvent multiEvent) ([]uncoreMetric, error) {
	if ie.eventReader == nil || ie.timer == nil {
		return nil, fmt.Errorf("event values reader or timer is nil")
	}
	if len(multiEvent.activeEvents) < 1 || multiEvent.perfEvent == nil {
		return nil, fmt.Errorf("no active events or perf event is nil")
	}
	activeEvents := multiEvent.activeEvents
	perfEvent := multiEvent.perfEvent

	metrics := make([]uncoreMetric, len(activeEvents))
	group := errgroup.Group{}

	for i, event := range activeEvents {
		id := i
		actualEvent := event

		group.Go(func() error {
			values, err := ie.eventReader.readValue(actualEvent)
			if err != nil {
				return fmt.Errorf("failed to read uncore event `%s` values: %v", actualEvent, err)
			}
			newMetric := uncoreMetric{
				values:   values,
				socket:   multiEvent.socket,
				unitType: perfEvent.PMUName,
				name:     perfEvent.Name,
				unit:     actualEvent.PMUName(),
				time:     ie.timer.now(),
			}
			metrics[id] = newMetric
			return nil
		})
		err := group.Wait()
		if err != nil {
			return nil, err
		}
	}
	return metrics, nil
}

func (ie *iaEntitiesValuesReader) readMultiEventAgg(multiEvent multiEvent) (uncoreMetric, error) {
	if ie.eventReader == nil || ie.timer == nil {
		return uncoreMetric{}, fmt.Errorf("event values reader or timer is nil")
	}
	if len(multiEvent.activeEvents) < 1 || multiEvent.perfEvent == nil {
		return uncoreMetric{}, fmt.Errorf("no active events or perf event is nil")
	}
	activeEvents := multiEvent.activeEvents
	perfEvent := multiEvent.perfEvent

	values := make([]ia.CounterValue, len(activeEvents))
	group := errgroup.Group{}

	for i, event := range activeEvents {
		id := i
		actualEvent := event

		group.Go(func() error {
			value, err := ie.eventReader.readValue(actualEvent)
			if err != nil {
				return fmt.Errorf("failed to read uncore event `%s` values: %v", actualEvent, err)
			}
			values[id] = value
			return nil
		})
	}
	err := group.Wait()
	if err != nil {
		return uncoreMetric{}, err
	}

	bRaw, bEnabled, bRunning := ia.AggregateValues(values)
	if !bRaw.IsUint64() || !bEnabled.IsUint64() || !bRunning.IsUint64() {
		return uncoreMetric{}, fmt.Errorf("cannot aggregate `%s` values, uint64 exceeding", perfEvent)
	}
	aggValues := ia.CounterValue{
		Raw:     bRaw.Uint64(),
		Enabled: bEnabled.Uint64(),
		Running: bRunning.Uint64(),
	}
	newMetric := uncoreMetric{
		values:   aggValues,
		socket:   multiEvent.socket,
		unitType: perfEvent.PMUName,
		name:     perfEvent.Name,
		time:     ie.timer.now(),
	}
	return newMetric, nil
}
