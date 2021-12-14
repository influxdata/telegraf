//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"testing"

	ia "github.com/intel/iaevents"
	"github.com/stretchr/testify/require"
)

type mockPlacementFactory struct {
	err bool
}

func (m *mockPlacementFactory) NewPlacements(_ string, cpu int, cpus ...int) ([]ia.PlacementProvider, error) {
	if m.err {
		return nil, errors.New("mock error")
	}
	placements := []ia.PlacementProvider{
		&ia.Placement{CPU: cpu, PMUType: 4},
	}
	for _, cpu := range cpus {
		placements = append(placements, &ia.Placement{CPU: cpu, PMUType: 4})
	}
	return placements, nil
}

func TestActivateEntities(t *testing.T) {
	mEntitiesActivator := &iaEntitiesActivator{}

	// more core test cases in TestActivateCoreEvents
	t.Run("failed to activate core events", func(t *testing.T) {
		tag := "TAG"
		mEntities := []*CoreEventEntity{{EventsTag: tag}}
		err := mEntitiesActivator.activateEntities(mEntities, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate core events `%s`", tag))
	})

	// more uncore test cases in TestActivateUncoreEvents
	t.Run("failed to activate uncore events", func(t *testing.T) {
		tag := "TAG"
		mEntities := []*UncoreEventEntity{{EventsTag: tag}}
		err := mEntitiesActivator.activateEntities(nil, mEntities)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate uncore events `%s`", tag))
	})

	t.Run("nothing to do", func(t *testing.T) {
		err := mEntitiesActivator.activateEntities(nil, nil)
		require.NoError(t, err)
	})
}

func TestActivateUncoreEvents(t *testing.T) {
	mActivator := &mockEventsActivator{}
	mMaker := &mockPlacementMaker{}
	errMock := fmt.Errorf("error mock")

	t.Run("entity is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		err := mEntitiesActivator.activateUncoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore events entity is nil")
	})

	t.Run("event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		mEntity := &UncoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore parsed event is nil")
	})

	t.Run("perf event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		name := "event name"
		mEntity := &UncoreEventEntity{parsedEvents: []*eventWithQuals{{name: name, custom: ia.CustomizableEvent{Event: nil}}}}
		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("perf event of `%s` event is nil", name))
	})

	t.Run("placement maker and perf activator is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: nil, perfActivator: nil}
		err := mEntitiesActivator.activateUncoreEvents(&UncoreEventEntity{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "events activator or placement maker is nil")
	})

	t.Run("failed to create placements", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		eventName := "mock event 1"
		parsedEvents := []*eventWithQuals{{name: eventName, custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: eventName}}}}
		mEntity := &UncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}

		mMaker.On("makeUncorePlacements", parsedEvents[0].custom.Event, mEntity.parsedSockets[0]).Return(nil, errMock).Once()
		err := mEntitiesActivator.activateUncoreEvents(mEntity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("ailed to create uncore placements for event `%s`", eventName))
		mMaker.AssertExpectations(t)
	})

	t.Run("failed to activate event", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		eventName := "mock event 1"
		parsedEvents := []*eventWithQuals{{name: eventName, custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: eventName}}}}
		placements := []ia.PlacementProvider{&ia.Placement{CPU: 0}, &ia.Placement{CPU: 1}}
		mEntity := &UncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}

		mMaker.On("makeUncorePlacements", parsedEvents[0].custom.Event, mEntity.parsedSockets[0]).Return(placements, nil).Once()
		mActivator.On("activateMulti", parsedEvents[0].custom.Event, placements, parsedEvents[0].custom.Options).Return(nil, errMock).Once()

		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate multi event `%s`", eventName))
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("successfully activate core events", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 1", Uncore: true}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 2", Uncore: true}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 3", Uncore: true}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 4", Uncore: true}}},
		}
		mEntity := &UncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}
		placements := []ia.PlacementProvider{&ia.Placement{}, &ia.Placement{}, &ia.Placement{}}

		var expectedEvents []multiEvent
		for _, event := range parsedEvents {
			for _, socket := range mEntity.parsedSockets {
				mMaker.On("makeUncorePlacements", event.custom.Event, socket).Return(placements, nil).Once()
				newActiveMultiEvent := &ia.ActiveMultiEvent{}
				expectedEvents = append(expectedEvents, multiEvent{newActiveMultiEvent.Events(), event.custom.Event, socket})
				mActivator.On("activateMulti", event.custom.Event, placements, event.custom.Options).Return(newActiveMultiEvent, nil).Once()
			}
		}
		err := mEntitiesActivator.activateUncoreEvents(mEntity)

		require.NoError(t, err)
		require.Equal(t, expectedEvents, mEntity.activeMultiEvents)
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})
}

func TestActivateCoreEvents(t *testing.T) {
	mMaker := &mockPlacementMaker{}
	mActivator := &mockEventsActivator{}
	errMock := fmt.Errorf("error mock")

	t.Run("entity is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		err := mEntitiesActivator.activateCoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core events entity is nil")
	})

	t.Run("placement maker is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: nil, perfActivator: mActivator}
		err := mEntitiesActivator.activateCoreEvents(&CoreEventEntity{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "placement maker is nil")
	})

	t.Run("event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		mEntity := &CoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core parsed event is nil")
	})

	t.Run("failed to create placements", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		parsedEvents := []*eventWithQuals{{name: "mock event 1", custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 1"}}}}
		mEntity := &CoreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		mMaker.On("makeCorePlacements", mEntity.parsedCores, parsedEvents[0].custom.Event).Return(nil, errMock).Once()
		err := mEntitiesActivator.activateCoreEvents(mEntity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to create core placements for event `%s`", parsedEvents[0].name))
		mMaker.AssertExpectations(t)
	})

	t.Run("failed to activate event", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{{name: "mock event 1", custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 1"}}}}
		placements := []ia.PlacementProvider{&ia.Placement{CPU: 0}, &ia.Placement{CPU: 1}}
		mEntity := &CoreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		event := parsedEvents[0]
		plc := placements[0]
		mMaker.On("makeCorePlacements", mEntity.parsedCores, event.custom.Event).Return(placements, nil).Once()
		mActivator.On("activateEvent", event.custom.Event, plc, event.custom.Options).Return(nil, errMock).Once()

		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate core event `%s`", parsedEvents[0].name))
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("failed to activate core events group", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: nil}
		mEntity := &CoreEventEntity{PerfGroup: true, parsedEvents: nil}

		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to activate core events group")
	})

	t.Run("successfully activate core events", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 1"}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 2"}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 3"}}},
			{custom: ia.CustomizableEvent{Event: &ia.PerfEvent{Name: "mock event 4"}}},
		}
		placements := []ia.PlacementProvider{&ia.Placement{CPU: 0}, &ia.Placement{CPU: 1}, &ia.Placement{CPU: 2}}
		mEntity := &CoreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		var activeEvents []*ia.ActiveEvent
		for _, event := range parsedEvents {
			mMaker.On("makeCorePlacements", mEntity.parsedCores, event.custom.Event).Return(placements, nil).Once()
			for _, plc := range placements {
				newActiveEvent := &ia.ActiveEvent{PerfEvent: event.custom.Event}
				activeEvents = append(activeEvents, newActiveEvent)
				mActivator.On("activateEvent", event.custom.Event, plc, event.custom.Options).Return(newActiveEvent, nil).Once()
			}
		}

		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.NoError(t, err)
		require.Equal(t, activeEvents, mEntity.activeEvents)
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})
}

func TestActivateCoreEventsGroup(t *testing.T) {
	mMaker := &mockPlacementMaker{}
	mActivator := &mockEventsActivator{}
	eActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
	errMock := errors.New("mock error")

	leader := &ia.PerfEvent{Name: "mock event 1"}
	perfEvent2 := &ia.PerfEvent{Name: "mock event 2"}

	parsedEvents := []*eventWithQuals{{custom: ia.CustomizableEvent{Event: leader}}, {custom: ia.CustomizableEvent{Event: perfEvent2}}}
	placements := []ia.PlacementProvider{&ia.Placement{}, &ia.Placement{}}

	// cannot populate this struct due to unexported events field
	activeGroup := &ia.ActiveEventGroup{}

	mEntity := &CoreEventEntity{
		EventsTag:    "mock group",
		PerfGroup:    true,
		parsedEvents: parsedEvents,
		parsedCores:  nil,
	}

	var events []ia.CustomizableEvent
	for _, event := range parsedEvents {
		events = append(events, event.custom)
	}

	t.Run("missing perf activator and placement maker", func(t *testing.T) {
		mActivator := &iaEntitiesActivator{}
		err := mActivator.activateCoreEventsGroup(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing perf activator or placement maker")
	})

	t.Run("missing parsed events", func(t *testing.T) {
		mActivator := &iaEntitiesActivator{placementMaker: &mockPlacementMaker{}, perfActivator: &mockEventsActivator{}}
		err := mActivator.activateCoreEventsGroup(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing parsed events")
	})

	t.Run("nil in parsed event", func(t *testing.T) {
		mEntity := &CoreEventEntity{EventsTag: "Nice tag", PerfGroup: true, parsedEvents: []*eventWithQuals{nil, nil}}
		err := eActivator.activateCoreEventsGroup(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core event is nil")
	})

	t.Run("failed to make core placements", func(t *testing.T) {
		mMaker.On("makeCorePlacements", mEntity.parsedCores, leader).Return(nil, errMock).Once()
		err := eActivator.activateCoreEventsGroup(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to make core placements")
		mMaker.AssertExpectations(t)
	})

	t.Run("failed to activate group", func(t *testing.T) {
		mMaker.On("makeCorePlacements", mEntity.parsedCores, leader).Return(placements, nil).Once()
		mActivator.On("activateGroup", placements[0], events).Return(nil, errMock).Once()

		err := eActivator.activateCoreEventsGroup(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), errMock.Error())
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	var allActive []*ia.ActiveEvent
	t.Run("successfully activated group", func(t *testing.T) {
		mMaker.On("makeCorePlacements", mEntity.parsedCores, leader).Return(placements, nil).Once()
		for _, plc := range placements {
			mActivator.On("activateGroup", plc, events).Return(activeGroup, nil).Once()
			allActive = append(allActive, activeGroup.Events()...)
		}

		err := eActivator.activateCoreEventsGroup(mEntity)
		require.NoError(t, err)
		require.Equal(t, allActive, mEntity.activeEvents)
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})
}

func TestMakeCorePlacements(t *testing.T) {
	tests := []struct {
		name      string
		cores     []int
		perfEvent ia.PlacementFactory
		result    []ia.PlacementProvider
		errMsg    string
	}{
		{"no cores", nil, &ia.PerfEvent{}, nil, "no cores provided"},
		{"one core placement", []int{1}, &mockPlacementFactory{}, []ia.PlacementProvider{&ia.Placement{CPU: 1, PMUType: 4}}, ""},
		{"multiple core placement", []int{1, 2, 4}, &mockPlacementFactory{}, []ia.PlacementProvider{
			&ia.Placement{CPU: 1, PMUType: 4},
			&ia.Placement{CPU: 2, PMUType: 4},
			&ia.Placement{CPU: 4, PMUType: 4}},
			""},
		{"placement factory error", []int{1}, &mockPlacementFactory{true}, nil, "mock error"},
		{"placement factory error 2", []int{1, 2, 3}, &mockPlacementFactory{true}, nil, "mock error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			maker := &iaPlacementMaker{}
			providers, err := maker.makeCorePlacements(test.cores, test.perfEvent)
			if len(test.errMsg) > 0 {
				require.Error(t, err)
				require.Nil(t, providers)
				require.Contains(t, err.Error(), test.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.result, providers)
		})
	}
}

func TestActivateEventForPlacement(t *testing.T) {
	placement1 := &ia.Placement{CPU: 0}
	placement2 := &ia.Placement{CPU: 1}
	placement3 := &ia.Placement{CPU: 2}

	mPlacements := []ia.PlacementProvider{placement1, placement2, placement3}

	mPerfEvent := &ia.PerfEvent{Name: "mock1"}
	mOptions := &ia.PerfEventOptions{}
	mEvent := &eventWithQuals{name: mPerfEvent.Name, custom: ia.CustomizableEvent{Event: mPerfEvent, Options: mOptions}}

	mPerfActivator := &mockEventsActivator{}
	mActivator := &iaEntitiesActivator{perfActivator: mPerfActivator}

	t.Run("event is nil", func(t *testing.T) {
		activeEvents, err := mActivator.activateEventForPlacements(nil, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core event is nil")
		require.Nil(t, activeEvents)
	})

	t.Run("perf activator is nil", func(t *testing.T) {
		mActivator := &iaEntitiesActivator{}
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing perf activator")
		require.Nil(t, activeEvents)
	})

	t.Run("placements are nil", func(t *testing.T) {
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, nil)
		require.NoError(t, err)
		require.Nil(t, activeEvents)
	})

	t.Run("activation error", func(t *testing.T) {
		mPerfActivator.On("activateEvent", mPerfEvent, placement1, mOptions).Once().Return(nil, errors.New("err"))
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate event `%s`", mEvent.name))
		require.Nil(t, activeEvents)
		mPerfActivator.AssertExpectations(t)
	})

	t.Run("successfully activated", func(t *testing.T) {
		mActiveEvent := &ia.ActiveEvent{}
		mActiveEvent2 := &ia.ActiveEvent{}
		mActiveEvent3 := &ia.ActiveEvent{}

		mPerfActivator.On("activateEvent", mPerfEvent, placement1, mOptions).Once().Return(mActiveEvent, nil).
			On("activateEvent", mPerfEvent, placement2, mOptions).Once().Return(mActiveEvent2, nil).
			On("activateEvent", mPerfEvent, placement3, mOptions).Once().Return(mActiveEvent3, nil)

		activeEvents, err := mActivator.activateEventForPlacements(mEvent, mPlacements)
		require.NoError(t, err)
		require.Len(t, activeEvents, len(mPlacements))
		require.Contains(t, activeEvents, mActiveEvent)
		require.Contains(t, activeEvents, mActiveEvent2)
		mPerfActivator.AssertExpectations(t)
	})
}
