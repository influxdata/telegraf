//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"errors"
	"fmt"

	ia "github.com/intel/iaevents"
)

type placementMaker interface {
	makeCorePlacements(cores []int, factory ia.PlacementFactory) ([]ia.PlacementProvider, error)
	makeUncorePlacements(socket int, factory ia.PlacementFactory) ([]ia.PlacementProvider, error)
}

type iaPlacementMaker struct{}

func (iaPlacementMaker) makeCorePlacements(cores []int, factory ia.PlacementFactory) ([]ia.PlacementProvider, error) {
	var err error
	var corePlacements []ia.PlacementProvider

	switch len(cores) {
	case 0:
		return nil, errors.New("no cores provided")
	case 1:
		corePlacements, err = ia.NewCorePlacements(factory, cores[0])
		if err != nil {
			return nil, err
		}
	default:
		corePlacements, err = ia.NewCorePlacements(factory, cores[0], cores[1:]...)
		if err != nil {
			return nil, err
		}
	}
	return corePlacements, nil
}

func (iaPlacementMaker) makeUncorePlacements(socket int, factory ia.PlacementFactory) ([]ia.PlacementProvider, error) {
	return ia.NewUncoreAllPlacements(factory, socket)
}

type eventsActivator interface {
	activateEvent(ia.Activator, ia.PlacementProvider, ia.Options) (*ia.ActiveEvent, error)
	activateGroup(ia.PlacementProvider, []ia.CustomizableEvent) (*ia.ActiveEventGroup, error)
	activateMulti(ia.MultiActivator, []ia.PlacementProvider, ia.Options) (*ia.ActiveMultiEvent, error)
}

type iaEventsActivator struct{}

func (iaEventsActivator) activateEvent(a ia.Activator, p ia.PlacementProvider, o ia.Options) (*ia.ActiveEvent, error) {
	return a.Activate(p, ia.NewEventTargetProcess(-1, 0), o)
}

func (iaEventsActivator) activateGroup(p ia.PlacementProvider, e []ia.CustomizableEvent) (*ia.ActiveEventGroup, error) {
	return ia.ActivateGroup(p, ia.NewEventTargetProcess(-1, 0), e)
}

func (iaEventsActivator) activateMulti(a ia.MultiActivator, p []ia.PlacementProvider, o ia.Options) (*ia.ActiveMultiEvent, error) {
	return a.ActivateMulti(p, ia.NewEventTargetProcess(-1, 0), o)
}

type entitiesActivator interface {
	activateEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) error
}

type iaEntitiesActivator struct {
	placementMaker placementMaker
	perfActivator  eventsActivator
}

func (ea *iaEntitiesActivator) activateEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) error {
	for _, coreEventsEntity := range coreEntities {
		err := ea.activateCoreEvents(coreEventsEntity)
		if err != nil {
			return fmt.Errorf("failed to activate core events `%s`: %v", coreEventsEntity.EventsTag, err)
		}
	}
	for _, uncoreEventsEntity := range uncoreEntities {
		err := ea.activateUncoreEvents(uncoreEventsEntity)
		if err != nil {
			return fmt.Errorf("failed to activate uncore events `%s`: %v", uncoreEventsEntity.EventsTag, err)
		}
	}
	return nil
}

func (ea *iaEntitiesActivator) activateCoreEvents(entity *CoreEventEntity) error {
	if entity == nil {
		return fmt.Errorf("core events entity is nil")
	}
	if ea.placementMaker == nil {
		return fmt.Errorf("placement maker is nil")
	}
	if entity.PerfGroup {
		err := ea.activateCoreEventsGroup(entity)
		if err != nil {
			return fmt.Errorf("failed to activate core events group: %v", err)
		}
	} else {
		for _, event := range entity.parsedEvents {
			if event == nil {
				return fmt.Errorf("core parsed event is nil")
			}
			placements, err := ea.placementMaker.makeCorePlacements(entity.parsedCores, event.custom.Event)
			if err != nil {
				return fmt.Errorf("failed to create core placements for event `%s`: %v", event.name, err)
			}
			activeEvent, err := ea.activateEventForPlacements(event, placements)
			if err != nil {
				return fmt.Errorf("failed to activate core event `%s`: %v", event.name, err)
			}
			entity.activeEvents = append(entity.activeEvents, activeEvent...)
		}
	}
	return nil
}

func (ea *iaEntitiesActivator) activateUncoreEvents(entity *UncoreEventEntity) error {
	if entity == nil {
		return fmt.Errorf("uncore events entity is nil")
	}
	if ea.perfActivator == nil || ea.placementMaker == nil {
		return fmt.Errorf("events activator or placement maker is nil")
	}
	for _, event := range entity.parsedEvents {
		if event == nil {
			return fmt.Errorf("uncore parsed event is nil")
		}
		perfEvent := event.custom.Event
		if perfEvent == nil {
			return fmt.Errorf("perf event of `%s` event is nil", event.name)
		}
		options := event.custom.Options

		for _, socket := range entity.parsedSockets {
			placements, err := ea.placementMaker.makeUncorePlacements(socket, perfEvent)
			if err != nil {
				return fmt.Errorf("failed to create uncore placements for event `%s`: %v", event.name, err)
			}
			activeMultiEvent, err := ea.perfActivator.activateMulti(perfEvent, placements, options)
			if err != nil {
				return fmt.Errorf("failed to activate multi event `%s`: %v", event.name, err)
			}
			events := activeMultiEvent.Events()
			entity.activeMultiEvents = append(entity.activeMultiEvents, multiEvent{events, perfEvent, socket})
		}
	}
	return nil
}

func (ea *iaEntitiesActivator) activateCoreEventsGroup(entity *CoreEventEntity) error {
	if ea.perfActivator == nil || ea.placementMaker == nil {
		return fmt.Errorf("missing perf activator or placement maker")
	}
	if entity == nil || len(entity.parsedEvents) < 1 {
		return fmt.Errorf("missing parsed events")
	}

	var events []ia.CustomizableEvent
	for _, event := range entity.parsedEvents {
		if event == nil {
			return fmt.Errorf("core event is nil")
		}
		events = append(events, event.custom)
	}
	leader := entity.parsedEvents[0].custom

	placements, err := ea.placementMaker.makeCorePlacements(entity.parsedCores, leader.Event)
	if err != nil {
		return fmt.Errorf("failed to make core placements: %v", err)
	}

	for _, plc := range placements {
		activeGroup, err := ea.perfActivator.activateGroup(plc, events)
		if err != nil {
			return err
		}
		entity.activeEvents = append(entity.activeEvents, activeGroup.Events()...)
	}
	return nil
}

func (ea *iaEntitiesActivator) activateEventForPlacements(event *eventWithQuals, placements []ia.PlacementProvider) ([]*ia.ActiveEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("core event is nil")
	}
	if ea.perfActivator == nil {
		return nil, fmt.Errorf("missing perf activator")
	}
	var activeEvents []*ia.ActiveEvent
	for _, placement := range placements {
		perfEvent := event.custom.Event
		options := event.custom.Options

		activeEvent, err := ea.perfActivator.activateEvent(perfEvent, placement, options)
		if err != nil {
			return nil, fmt.Errorf("failed to activate event `%s`: %v", event.name, err)
		}
		activeEvents = append(activeEvents, activeEvent)
	}
	return activeEvents, nil
}
