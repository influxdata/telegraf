//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	ia "github.com/intel/iaevents"
)

type entitiesResolver interface {
	resolveEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) error
}

type iaEntitiesResolver struct {
	reader      ia.Reader
	transformer ia.Transformer
	log         telegraf.Logger
}

func (e *iaEntitiesResolver) resolveEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) error {
	for _, entity := range coreEntities {
		if entity == nil {
			return fmt.Errorf("core entity is nil")
		}
		if entity.allEvents {
			newEvents, _, err := e.resolveAllEvents()
			if err != nil {
				return fmt.Errorf("failed to resolve all events: %v", err)
			}
			entity.parsedEvents = newEvents
			continue
		}
		for _, event := range entity.parsedEvents {
			if event == nil {
				return fmt.Errorf("parsed core event is nil")
			}
			customEvent, err := e.resolveEvent(event.name, event.qualifiers)
			if err != nil {
				return fmt.Errorf("failed to resolve core event `%s`: %v", event.name, err)
			}
			if customEvent.Event.Uncore {
				return fmt.Errorf("uncore event `%s` found in core entity", event.name)
			}
			event.custom = customEvent
		}
	}
	for _, entity := range uncoreEntities {
		if entity == nil {
			return fmt.Errorf("uncore entity is nil")
		}
		if entity.allEvents {
			_, newEvents, err := e.resolveAllEvents()
			if err != nil {
				return fmt.Errorf("failed to resolve all events: %v", err)
			}
			entity.parsedEvents = newEvents
			continue
		}
		for _, event := range entity.parsedEvents {
			if event == nil {
				return fmt.Errorf("parsed uncore event is nil")
			}
			customEvent, err := e.resolveEvent(event.name, event.qualifiers)
			if err != nil {
				return fmt.Errorf("failed to resolve uncore event `%s`: %v", event.name, err)
			}
			if !customEvent.Event.Uncore {
				return fmt.Errorf("core event `%s` found in uncore entity", event.name)
			}
			event.custom = customEvent
		}
	}
	return nil
}

func (e *iaEntitiesResolver) resolveAllEvents() (coreEvents []*eventWithQuals, uncoreEvents []*eventWithQuals, err error) {
	if e.transformer == nil {
		return nil, nil, errors.New("transformer is nil")
	}

	perfEvents, err := e.transformer.Transform(e.reader, ia.NewNameMatcher())
	if err != nil {
		re, ok := err.(*ia.TransformationError)
		if !ok {
			return nil, nil, err
		}
		if e.log != nil && re != nil {
			var eventErrs []string
			for _, eventErr := range re.Errors() {
				if eventErr == nil {
					continue
				}
				eventErrs = append(eventErrs, eventErr.Error())
			}
			errorsStr := strings.Join(eventErrs, ",\n")
			e.log.Warnf("Cannot resolve all of the events from provided files:\n%s.\nSome events may be omitted.", errorsStr)
		}
	}

	for _, perfEvent := range perfEvents {
		newEvent := &eventWithQuals{
			name:   perfEvent.Name,
			custom: ia.CustomizableEvent{Event: perfEvent},
		}
		// build options for event
		newEvent.custom.Options, err = ia.NewOptions().Build()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build options for event `%s`: %v", perfEvent.Name, err)
		}
		if perfEvent.Uncore {
			uncoreEvents = append(uncoreEvents, newEvent)
			continue
		}
		coreEvents = append(coreEvents, newEvent)
	}
	return coreEvents, uncoreEvents, nil
}

func (e *iaEntitiesResolver) resolveEvent(name string, qualifiers []string) (ia.CustomizableEvent, error) {
	var custom ia.CustomizableEvent
	if e.transformer == nil {
		return custom, errors.New("events transformer is nil")
	}
	if name == "" {
		return custom, errors.New("event name is empty")
	}
	matcher := ia.NewNameMatcher(name)
	perfEvents, err := e.transformer.Transform(e.reader, matcher)
	if err != nil {
		return custom, fmt.Errorf("failed to transform perf events: %v", err)
	}
	if len(perfEvents) < 1 {
		return custom, fmt.Errorf("failed to resolve unknown event `%s`", name)
	}
	// build options for event
	options, err := ia.NewOptions().SetAttrModifiers(qualifiers).Build()
	if err != nil {
		return custom, fmt.Errorf("failed to build options for event `%s`: %v", name, err)
	}
	custom = ia.CustomizableEvent{
		Event:   perfEvents[0],
		Options: options,
	}
	return custom, nil
}
