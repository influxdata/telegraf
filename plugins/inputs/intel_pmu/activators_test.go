//go:build linux && amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"testing"

	"github.com/intel/iaevents"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPlacementFactory struct {
	err bool
}

func (m *mockPlacementFactory) NewPlacements(_ string, cpu int, cpus ...int) ([]iaevents.PlacementProvider, error) {
	if m.err {
		return nil, errors.New("mock error")
	}
	placements := make([]iaevents.PlacementProvider, 0, len(cpus)+1)
	placements = append(placements, &iaevents.Placement{CPU: cpu, PMUType: 4})
	for _, cpu := range cpus {
		placements = append(placements, &iaevents.Placement{CPU: cpu, PMUType: 4})
	}
	return placements, nil
}

func TestActivateEntities(t *testing.T) {
	mEntitiesActivator := &iaEntitiesActivator{}

	// more core test cases in TestActivateCoreEvents
	t.Run("failed to activate core events", func(t *testing.T) {
		tag := "TAG"
		mEntities := []*coreEventEntity{{EventsTag: tag}}
		err := mEntitiesActivator.activateEntities(mEntities, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate core events %q", tag))
	})

	// more uncore test cases in TestActivateUncoreEvents
	t.Run("failed to activate uncore events", func(t *testing.T) {
		tag := "TAG"
		mEntities := []*uncoreEventEntity{{EventsTag: tag}}
		err := mEntitiesActivator.activateEntities(nil, mEntities)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate uncore events %q", tag))
	})

	t.Run("nothing to do", func(t *testing.T) {
		err := mEntitiesActivator.activateEntities(nil, nil)
		require.NoError(t, err)
	})
}

func TestActivateUncoreEvents(t *testing.T) {
	mActivator := &mockEventsActivator{}
	mMaker := &mockPlacementMaker{}
	errMock := errors.New("error mock")

	t.Run("entity is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		err := mEntitiesActivator.activateUncoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore events entity is nil")
	})

	t.Run("event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		mEntity := &uncoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore parsed event is nil")
	})

	t.Run("perf event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		name := "event name"
		mEntity := &uncoreEventEntity{parsedEvents: []*eventWithQuals{{name: name, custom: iaevents.CustomizableEvent{Event: nil}}}}
		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("perf event of %q event is nil", name))
	})

	t.Run("placement maker and perf activator is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: nil, perfActivator: nil}
		err := mEntitiesActivator.activateUncoreEvents(&uncoreEventEntity{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "events activator or placement maker is nil")
	})

	t.Run("failed to create placements", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		eventName := "mock event 1"
		parsedEvents := []*eventWithQuals{{name: eventName, custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: eventName}}}}
		mEntity := &uncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}

		mMaker.On("makeUncorePlacements", parsedEvents[0].custom.Event, mEntity.parsedSockets[0]).Return(nil, errMock).Once()
		err := mEntitiesActivator.activateUncoreEvents(mEntity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("ailed to create uncore placements for event %q", eventName))
		mMaker.AssertExpectations(t)
	})

	t.Run("failed to activate event", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		eventName := "mock event 1"
		parsedEvents := []*eventWithQuals{{name: eventName, custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: eventName}}}}
		placements := []iaevents.PlacementProvider{&iaevents.Placement{CPU: 0}, &iaevents.Placement{CPU: 1}}
		mEntity := &uncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}

		mMaker.On("makeUncorePlacements", parsedEvents[0].custom.Event, mEntity.parsedSockets[0]).Return(placements, nil).Once()
		mActivator.On("activateMulti", parsedEvents[0].custom.Event, placements, parsedEvents[0].custom.Options).Return(nil, errMock).Once()

		err := mEntitiesActivator.activateUncoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate multi event %q", eventName))
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("successfully activate core events", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 1", Uncore: true}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 2", Uncore: true}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 3", Uncore: true}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 4", Uncore: true}}},
		}
		mEntity := &uncoreEventEntity{parsedEvents: parsedEvents, parsedSockets: []int{0, 1, 2}}
		placements := []iaevents.PlacementProvider{&iaevents.Placement{}, &iaevents.Placement{}, &iaevents.Placement{}}

		expectedEvents := make([]multiEvent, 0, len(parsedEvents))
		for _, event := range parsedEvents {
			for _, socket := range mEntity.parsedSockets {
				mMaker.On("makeUncorePlacements", event.custom.Event, socket).Return(placements, nil).Once()
				newActiveMultiEvent := &iaevents.ActiveMultiEvent{}
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
	errMock := errors.New("error mock")

	t.Run("entity is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		err := mEntitiesActivator.activateCoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core events entity is nil")
	})

	t.Run("placement maker is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: nil, perfActivator: mActivator}
		err := mEntitiesActivator.activateCoreEvents(&coreEventEntity{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "placement maker is nil")
	})

	t.Run("event is nil", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		mEntity := &coreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core parsed event is nil")
	})

	t.Run("failed to create placements", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}
		parsedEvents := []*eventWithQuals{{name: "mock event 1", custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 1"}}}}
		mEntity := &coreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		mMaker.On("makeCorePlacements", mEntity.parsedCores, parsedEvents[0].custom.Event).Return(nil, errMock).Once()
		err := mEntitiesActivator.activateCoreEvents(mEntity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to create core placements for event %q", parsedEvents[0].name))
		mMaker.AssertExpectations(t)
	})

	t.Run("failed to activate event", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{{name: "mock event 1", custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 1"}}}}
		placements := []iaevents.PlacementProvider{&iaevents.Placement{CPU: 0}, &iaevents.Placement{CPU: 1}}
		mEntity := &coreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		event := parsedEvents[0]
		plc := placements[0]
		mMaker.On("makeCorePlacements", mEntity.parsedCores, event.custom.Event).Return(placements, nil).Once()
		mActivator.On("activateEvent", event.custom.Event, plc, event.custom.Options).Return(nil, errMock).Once()

		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate core event %q", parsedEvents[0].name))
		mMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("failed to activate core events group", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: nil}
		mEntity := &coreEventEntity{PerfGroup: true, parsedEvents: nil}

		err := mEntitiesActivator.activateCoreEvents(mEntity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to activate core events group")
	})

	t.Run("successfully activate core events", func(t *testing.T) {
		mEntitiesActivator := &iaEntitiesActivator{placementMaker: mMaker, perfActivator: mActivator}

		parsedEvents := []*eventWithQuals{
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 1"}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 2"}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 3"}}},
			{custom: iaevents.CustomizableEvent{Event: &iaevents.PerfEvent{Name: "mock event 4"}}},
		}
		placements := []iaevents.PlacementProvider{&iaevents.Placement{CPU: 0}, &iaevents.Placement{CPU: 1}, &iaevents.Placement{CPU: 2}}
		mEntity := &coreEventEntity{PerfGroup: false, parsedEvents: parsedEvents, parsedCores: []int{0, 1, 2}}

		activeEvents := make([]*iaevents.ActiveEvent, 0, len(parsedEvents))
		for _, event := range parsedEvents {
			mMaker.On("makeCorePlacements", mEntity.parsedCores, event.custom.Event).Return(placements, nil).Once()
			for _, plc := range placements {
				newActiveEvent := &iaevents.ActiveEvent{PerfEvent: event.custom.Event}
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

	leader := &iaevents.PerfEvent{Name: "mock event 1"}
	perfEvent2 := &iaevents.PerfEvent{Name: "mock event 2"}

	parsedEvents := []*eventWithQuals{{custom: iaevents.CustomizableEvent{Event: leader}}, {custom: iaevents.CustomizableEvent{Event: perfEvent2}}}
	placements := []iaevents.PlacementProvider{&iaevents.Placement{}, &iaevents.Placement{}}

	// cannot populate this struct due to unexported events field
	activeGroup := &iaevents.ActiveEventGroup{}

	mEntity := &coreEventEntity{
		EventsTag:    "mock group",
		PerfGroup:    true,
		parsedEvents: parsedEvents,
		parsedCores:  nil,
	}

	events := make([]iaevents.CustomizableEvent, 0, len(parsedEvents))
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
		mEntity := &coreEventEntity{EventsTag: "Nice tag", PerfGroup: true, parsedEvents: []*eventWithQuals{nil, nil}}
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

	var allActive []*iaevents.ActiveEvent
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
		perfEvent iaevents.PlacementFactory
		result    []iaevents.PlacementProvider
		errMsg    string
	}{
		{"no cores", nil, &iaevents.PerfEvent{}, nil, "no cores provided"},
		{"one core placement", []int{1}, &mockPlacementFactory{}, []iaevents.PlacementProvider{&iaevents.Placement{CPU: 1, PMUType: 4}}, ""},
		{"multiple core placement", []int{1, 2, 4}, &mockPlacementFactory{}, []iaevents.PlacementProvider{
			&iaevents.Placement{CPU: 1, PMUType: 4},
			&iaevents.Placement{CPU: 2, PMUType: 4},
			&iaevents.Placement{CPU: 4, PMUType: 4}},
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
	placement1 := &iaevents.Placement{CPU: 0}
	placement2 := &iaevents.Placement{CPU: 1}
	placement3 := &iaevents.Placement{CPU: 2}

	mPlacements := []iaevents.PlacementProvider{placement1, placement2, placement3}

	mPerfEvent := &iaevents.PerfEvent{Name: "mock1"}
	mOptions := &iaevents.PerfEventOptions{}
	mEvent := &eventWithQuals{name: mPerfEvent.Name, custom: iaevents.CustomizableEvent{Event: mPerfEvent, Options: mOptions}}

	mPerfActivator := &mockEventsActivator{}
	mActivator := &iaEntitiesActivator{perfActivator: mPerfActivator}

	t.Run("event is nil", func(t *testing.T) {
		activeEvents, err := mActivator.activateEventForPlacements(nil, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), "core event is nil")
		require.Empty(t, activeEvents)
	})

	t.Run("perf activator is nil", func(t *testing.T) {
		mActivator := &iaEntitiesActivator{}
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing perf activator")
		require.Empty(t, activeEvents)
	})

	t.Run("placements are nil", func(t *testing.T) {
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, nil)
		require.NoError(t, err)
		require.Empty(t, activeEvents)
	})

	t.Run("activation error", func(t *testing.T) {
		mPerfActivator.On("activateEvent", mPerfEvent, placement1, mOptions).Once().Return(nil, errors.New("err"))
		activeEvents, err := mActivator.activateEventForPlacements(mEvent, mPlacements)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to activate event %q", mEvent.name))
		require.Nil(t, activeEvents)
		mPerfActivator.AssertExpectations(t)
	})

	t.Run("successfully activated", func(t *testing.T) {
		mActiveEvent := &iaevents.ActiveEvent{}
		mActiveEvent2 := &iaevents.ActiveEvent{}
		mActiveEvent3 := &iaevents.ActiveEvent{}

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

// Mocking

// mockEventsActivator is an autogenerated mock type for the eventsActivator type
type mockEventsActivator struct {
	mock.Mock
}

// activateEvent provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockEventsActivator) activateEvent(_a0 iaevents.Activator, _a1 iaevents.PlacementProvider, _a2 iaevents.Options) (*iaevents.ActiveEvent, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *iaevents.ActiveEvent
	if rf, ok := ret.Get(0).(func(iaevents.Activator, iaevents.PlacementProvider, iaevents.Options) *iaevents.ActiveEvent); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iaevents.ActiveEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(iaevents.Activator, iaevents.PlacementProvider, iaevents.Options) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// activateGroup provides a mock function with given fields: _a0, _a1
func (_m *mockEventsActivator) activateGroup(_a0 iaevents.PlacementProvider, _a1 []iaevents.CustomizableEvent) (*iaevents.ActiveEventGroup, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *iaevents.ActiveEventGroup
	if rf, ok := ret.Get(0).(func(iaevents.PlacementProvider, []iaevents.CustomizableEvent) *iaevents.ActiveEventGroup); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iaevents.ActiveEventGroup)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(iaevents.PlacementProvider, []iaevents.CustomizableEvent) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// activateMulti provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockEventsActivator) activateMulti(
	_a0 iaevents.MultiActivator,
	_a1 []iaevents.PlacementProvider,
	_a2 iaevents.Options,
) (*iaevents.ActiveMultiEvent, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *iaevents.ActiveMultiEvent
	if rf, ok := ret.Get(0).(func(iaevents.MultiActivator, []iaevents.PlacementProvider, iaevents.Options) *iaevents.ActiveMultiEvent); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iaevents.ActiveMultiEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(iaevents.MultiActivator, []iaevents.PlacementProvider, iaevents.Options) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockPlacementMaker is an autogenerated mock type for the placementMaker type
type mockPlacementMaker struct {
	mock.Mock
}

// makeCorePlacements provides a mock function with given fields: cores, perfEvent
func (_m *mockPlacementMaker) makeCorePlacements(cores []int, factory iaevents.PlacementFactory) ([]iaevents.PlacementProvider, error) {
	ret := _m.Called(cores, factory)

	var r0 []iaevents.PlacementProvider
	if rf, ok := ret.Get(0).(func([]int, iaevents.PlacementFactory) []iaevents.PlacementProvider); ok {
		r0 = rf(cores, factory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iaevents.PlacementProvider)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int, iaevents.PlacementFactory) error); ok {
		r1 = rf(cores, factory)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// makeUncorePlacements provides a mock function with given fields: factory, socket
func (_m *mockPlacementMaker) makeUncorePlacements(socket int, factory iaevents.PlacementFactory) ([]iaevents.PlacementProvider, error) {
	ret := _m.Called(factory, socket)

	var r0 []iaevents.PlacementProvider
	if rf, ok := ret.Get(0).(func(iaevents.PlacementFactory, int) []iaevents.PlacementProvider); ok {
		r0 = rf(factory, socket)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iaevents.PlacementProvider)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(iaevents.PlacementFactory, int) error); ok {
		r1 = rf(factory, socket)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
