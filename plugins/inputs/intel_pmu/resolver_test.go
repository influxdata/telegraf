//go:build linux && amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"testing"

	"github.com/intel/iaevents"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestResolveEntities(t *testing.T) {
	errMock := errors.New("mock error")
	mLog := testutil.Logger{}
	mTransformer := &mockTransformer{}
	mResolver := &iaEntitiesResolver{transformer: mTransformer, log: mLog}

	type test struct {
		perfEvent *iaevents.PerfEvent
		options   iaevents.Options
		event     *eventWithQuals
	}

	t.Run("nil entities", func(t *testing.T) {
		err := mResolver.resolveEntities([]*coreEventEntity{nil}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "core entity is nil")

		err = mResolver.resolveEntities(nil, []*uncoreEventEntity{nil})

		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore entity is nil")
	})

	t.Run("nil parsed events", func(t *testing.T) {
		mCoreEntity := &coreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		mUncoreEntity := &uncoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}

		err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "parsed core event is nil")

		err = mResolver.resolveEntities(nil, []*uncoreEventEntity{mUncoreEntity})

		require.Error(t, err)
		require.Contains(t, err.Error(), "parsed uncore event is nil")
	})

	t.Run("fail to resolve core events", func(t *testing.T) {
		name := "mock event 1"
		mCoreEntity := &coreEventEntity{parsedEvents: []*eventWithQuals{{name: name}}, allEvents: false}
		matcher := iaevents.NewNameMatcher(name)

		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
		err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to resolve core event %q", name))
		mTransformer.AssertExpectations(t)
	})

	t.Run("fail to resolve uncore events", func(t *testing.T) {
		name := "mock event 1"
		mUncoreEntity := &uncoreEventEntity{parsedEvents: []*eventWithQuals{{name: name}}, allEvents: false}
		matcher := iaevents.NewNameMatcher(name)

		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
		err := mResolver.resolveEntities(nil, []*uncoreEventEntity{mUncoreEntity})

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to resolve uncore event %q", name))
		mTransformer.AssertExpectations(t)
	})

	t.Run("resolve all core and uncore events", func(t *testing.T) {
		mCoreEntity := &coreEventEntity{allEvents: true}
		mUncoreEntity := &uncoreEventEntity{allEvents: true}
		corePerfEvents := []*iaevents.PerfEvent{
			{Name: "core event1"},
			{Name: "core event2"},
			{Name: "core event3"},
		}
		uncorePerfEvents := []*iaevents.PerfEvent{
			{Name: "uncore event1", Uncore: true},
			{Name: "uncore event2", Uncore: true},
			{Name: "uncore event3", Uncore: true},
		}
		matcher := iaevents.NewNameMatcher()

		t.Run("fail to resolve all core events", func(t *testing.T) {
			mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
			err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to resolve all events")
			mTransformer.AssertExpectations(t)
		})

		t.Run("fail to resolve all uncore events", func(t *testing.T) {
			mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
			err := mResolver.resolveEntities(nil, []*uncoreEventEntity{mUncoreEntity})
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to resolve all events")
			mTransformer.AssertExpectations(t)
		})

		t.Run("fail to resolve all events with transformationError", func(t *testing.T) {
			transformErr := &iaevents.TransformationError{}

			mTransformer.On("Transform", nil, matcher).Once().Return(corePerfEvents, transformErr).Once()
			mTransformer.On("Transform", nil, matcher).Once().Return(uncorePerfEvents, transformErr).Once()

			err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, []*uncoreEventEntity{mUncoreEntity})
			require.NoError(t, err)
			require.Len(t, mCoreEntity.parsedEvents, len(corePerfEvents))
			require.Len(t, mUncoreEntity.parsedEvents, len(uncorePerfEvents))
			for _, coreEvent := range mCoreEntity.parsedEvents {
				require.Contains(t, corePerfEvents, coreEvent.custom.Event)
			}
			for _, uncoreEvent := range mUncoreEntity.parsedEvents {
				require.Contains(t, uncorePerfEvents, uncoreEvent.custom.Event)
			}
			mTransformer.AssertExpectations(t)
		})

		mTransformer.On("Transform", nil, matcher).Once().Return(corePerfEvents, nil).Once()
		mTransformer.On("Transform", nil, matcher).Once().Return(uncorePerfEvents, nil).Once()

		err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, []*uncoreEventEntity{mUncoreEntity})
		require.NoError(t, err)
		require.Len(t, mCoreEntity.parsedEvents, len(corePerfEvents))
		require.Len(t, mUncoreEntity.parsedEvents, len(uncorePerfEvents))
		for _, coreEvent := range mCoreEntity.parsedEvents {
			require.Contains(t, corePerfEvents, coreEvent.custom.Event)
		}
		for _, uncoreEvent := range mUncoreEntity.parsedEvents {
			require.Contains(t, uncorePerfEvents, uncoreEvent.custom.Event)
		}
		mTransformer.AssertExpectations(t)
	})

	t.Run("uncore event found in core entity", func(t *testing.T) {
		mQuals := []string{"config1=0x23h"}
		eventName := "uncore event 1"

		testCase := test{
			event:     &eventWithQuals{name: eventName, qualifiers: mQuals},
			perfEvent: &iaevents.PerfEvent{Name: eventName, Uncore: true},
		}

		matcher := iaevents.NewNameMatcher(eventName)
		mTransformer.On("Transform", nil, matcher).Return([]*iaevents.PerfEvent{testCase.perfEvent}, nil).Once()

		mCoreEntity := &coreEventEntity{parsedEvents: []*eventWithQuals{testCase.event}, allEvents: false}
		err := mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, nil)
		require.ErrorContains(t, err, fmt.Sprintf("uncore event %q found in core entity", eventName))
		mTransformer.AssertExpectations(t)
	})

	t.Run("core event found in uncore entity", func(t *testing.T) {
		mQuals := []string{"config1=0x23h"}
		eventName := "core event 1"

		testCase := test{
			event:     &eventWithQuals{name: eventName, qualifiers: mQuals},
			perfEvent: &iaevents.PerfEvent{Name: eventName, Uncore: false},
		}

		matcher := iaevents.NewNameMatcher(eventName)
		mTransformer.On("Transform", nil, matcher).Return([]*iaevents.PerfEvent{testCase.perfEvent}, nil).Once()

		mUncoreEntity := &uncoreEventEntity{parsedEvents: []*eventWithQuals{testCase.event}, allEvents: false}
		err := mResolver.resolveEntities(nil, []*uncoreEventEntity{mUncoreEntity})

		require.ErrorContains(t, err, fmt.Sprintf("core event %q found in uncore entity", eventName))
		mTransformer.AssertExpectations(t)
	})

	t.Run("resolve core and uncore events", func(t *testing.T) {
		mQuals := []string{"config1=0x23h"}
		mOptions, err := iaevents.NewOptions().SetAttrModifiers(mQuals).Build()
		require.NoError(t, err)
		emptyOptions, err := iaevents.NewOptions().Build()
		require.NoError(t, err)

		coreTestCases := []test{
			{event: &eventWithQuals{name: "core1", qualifiers: mQuals},
				options:   mOptions,
				perfEvent: &iaevents.PerfEvent{Name: "core1"}},
			{event: &eventWithQuals{name: "core2", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &iaevents.PerfEvent{Name: "core2"}},
			{event: &eventWithQuals{name: "core3", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &iaevents.PerfEvent{Name: "core3"}},
		}
		uncoreTestCases := []test{
			{event: &eventWithQuals{name: "uncore1", qualifiers: mQuals},
				options:   mOptions,
				perfEvent: &iaevents.PerfEvent{Name: "uncore1", Uncore: true}},
			{event: &eventWithQuals{name: "uncore2", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &iaevents.PerfEvent{Name: "uncore2", Uncore: true}},
			{event: &eventWithQuals{name: "uncore3", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &iaevents.PerfEvent{Name: "uncore3", Uncore: true}},
		}

		mCoreEvents := make([]*eventWithQuals, 0, len(coreTestCases))
		for _, test := range coreTestCases {
			matcher := iaevents.NewNameMatcher(test.event.name)
			mTransformer.On("Transform", nil, matcher).Return([]*iaevents.PerfEvent{test.perfEvent}, nil).Once()
			mCoreEvents = append(mCoreEvents, test.event)
		}

		nUncoreEvents := make([]*eventWithQuals, 0, len(uncoreTestCases))
		for _, test := range uncoreTestCases {
			matcher := iaevents.NewNameMatcher(test.event.name)
			mTransformer.On("Transform", nil, matcher).Return([]*iaevents.PerfEvent{test.perfEvent}, nil).Once()
			nUncoreEvents = append(nUncoreEvents, test.event)
		}

		mCoreEntity := &coreEventEntity{parsedEvents: mCoreEvents, allEvents: false}
		mUncoreEntity := &uncoreEventEntity{parsedEvents: nUncoreEvents, allEvents: false}
		err = mResolver.resolveEntities([]*coreEventEntity{mCoreEntity}, []*uncoreEventEntity{mUncoreEntity})

		require.NoError(t, err)
		for _, test := range append(coreTestCases, uncoreTestCases...) {
			require.Equal(t, test.perfEvent, test.event.custom.Event)
			require.Equal(t, test.options, test.event.custom.Options)
		}
		mTransformer.AssertExpectations(t)
	})
}

func TestResolveAllEvents(t *testing.T) {
	mTransformer := &mockTransformer{}

	mResolver := &iaEntitiesResolver{transformer: mTransformer}

	t.Run("transformer is nil", func(t *testing.T) {
		mResolver := &iaEntitiesResolver{transformer: nil}
		_, _, err := mResolver.resolveAllEvents()
		require.Error(t, err)
	})

	t.Run("transformer returns error", func(t *testing.T) {
		matcher := iaevents.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errors.New("mock error"))

		_, _, err := mResolver.resolveAllEvents()
		require.Error(t, err)
		mTransformer.AssertExpectations(t)
	})

	t.Run("no events", func(t *testing.T) {
		matcher := iaevents.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, nil)

		_, _, err := mResolver.resolveAllEvents()
		require.NoError(t, err)
		mTransformer.AssertExpectations(t)
	})

	t.Run("successfully resolved events", func(t *testing.T) {
		perfEvent1 := &iaevents.PerfEvent{Name: "mock1"}
		perfEvent2 := &iaevents.PerfEvent{Name: "mock2"}
		uncorePerfEvent1 := &iaevents.PerfEvent{Name: "mock3", Uncore: true}
		uncorePerfEvent2 := &iaevents.PerfEvent{Name: "mock4", Uncore: true}

		options, err := iaevents.NewOptions().Build()
		require.NoError(t, err)
		perfEvents := []*iaevents.PerfEvent{perfEvent1, perfEvent2, uncorePerfEvent1, uncorePerfEvent2}

		expectedCore := []*eventWithQuals{
			{name: perfEvent1.Name, custom: iaevents.CustomizableEvent{Event: perfEvent1, Options: options}},
			{name: perfEvent2.Name, custom: iaevents.CustomizableEvent{Event: perfEvent2, Options: options}},
		}

		expectedUncore := []*eventWithQuals{
			{name: uncorePerfEvent1.Name, custom: iaevents.CustomizableEvent{Event: uncorePerfEvent1, Options: options}},
			{name: uncorePerfEvent2.Name, custom: iaevents.CustomizableEvent{Event: uncorePerfEvent2, Options: options}},
		}

		matcher := iaevents.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(perfEvents, nil)

		coreEvents, uncoreEvents, err := mResolver.resolveAllEvents()
		require.NoError(t, err)
		require.Equal(t, expectedCore, coreEvents)
		require.Equal(t, expectedUncore, uncoreEvents)

		mTransformer.AssertExpectations(t)
	})
}

func TestResolveEvent(t *testing.T) {
	mTransformer := &mockTransformer{}
	mEvent := "mock event"

	mResolver := &iaEntitiesResolver{transformer: mTransformer}

	t.Run("transformer is nil", func(t *testing.T) {
		mResolver := &iaEntitiesResolver{transformer: nil}
		_, err := mResolver.resolveEvent("event", nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "events transformer is nil")
	})

	t.Run("event is empty", func(t *testing.T) {
		_, err := mResolver.resolveEvent("", nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event name is empty")
	})

	t.Run("transformer returns error", func(t *testing.T) {
		matcher := iaevents.NewNameMatcher(mEvent)
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errors.New("mock error"))

		_, err := mResolver.resolveEvent(mEvent, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to transform perf events")
		mTransformer.AssertExpectations(t)
	})

	t.Run("no events transformed", func(t *testing.T) {
		matcher := iaevents.NewNameMatcher(mEvent)
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, nil)

		_, err := mResolver.resolveEvent(mEvent, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve unknown event")
		mTransformer.AssertExpectations(t)
	})

	t.Run("not valid qualifiers", func(t *testing.T) {
		event := "mock event 1"
		qualifiers := []string{"wrong modifiers"}

		matcher := iaevents.NewNameMatcher(event)
		mPerfEvent := &iaevents.PerfEvent{Name: event}
		mPerfEvents := []*iaevents.PerfEvent{mPerfEvent}
		mTransformer.On("Transform", nil, matcher).Once().Return(mPerfEvents, nil)

		_, err := mResolver.resolveEvent(event, qualifiers)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to build options for event %q", event))
		mTransformer.AssertExpectations(t)
	})

	t.Run("successfully transformed", func(t *testing.T) {
		event := "mock event 1"
		qualifiers := []string{"config1=0x012h", "config2=0x034k"}

		matcher := iaevents.NewNameMatcher(event)

		mPerfEvent := &iaevents.PerfEvent{Name: event}
		mPerfEvents := []*iaevents.PerfEvent{mPerfEvent}

		expectedOptions, err := iaevents.NewOptions().SetAttrModifiers(qualifiers).Build()
		require.NoError(t, err)

		mTransformer.On("Transform", nil, matcher).Once().Return(mPerfEvents, nil)

		customEvent, err := mResolver.resolveEvent(event, qualifiers)
		require.NoError(t, err)
		require.Equal(t, mPerfEvent, customEvent.Event)
		require.Equal(t, expectedOptions, customEvent.Options)
		mTransformer.AssertExpectations(t)
	})
}

// Mocking

// mockTransformer is an autogenerated mock type for the Transformer type
type mockTransformer struct {
	mock.Mock
}

// Transform provides a mock function with given fields: reader, matcher
func (_m *mockTransformer) Transform(reader iaevents.Reader, matcher iaevents.Matcher) ([]*iaevents.PerfEvent, error) {
	ret := _m.Called(reader, matcher)

	var r0 []*iaevents.PerfEvent
	if rf, ok := ret.Get(0).(func(iaevents.Reader, iaevents.Matcher) []*iaevents.PerfEvent); ok {
		r0 = rf(reader, matcher)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*iaevents.PerfEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(iaevents.Reader, iaevents.Matcher) error); ok {
		r1 = rf(reader, matcher)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
