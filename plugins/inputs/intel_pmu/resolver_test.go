//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	ia "github.com/intel/iaevents"
	"github.com/stretchr/testify/require"
)

func TestResolveEntities(t *testing.T) {
	errMock := errors.New("mock error")
	mLog := testutil.Logger{}
	mTransformer := &MockTransformer{}
	mResolver := &iaEntitiesResolver{transformer: mTransformer, log: mLog}

	type test struct {
		perfEvent *ia.PerfEvent
		options   ia.Options
		event     *eventWithQuals
	}

	t.Run("nil entities", func(t *testing.T) {
		err := mResolver.resolveEntities([]*CoreEventEntity{nil}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "core entity is nil")

		err = mResolver.resolveEntities(nil, []*UncoreEventEntity{nil})

		require.Error(t, err)
		require.Contains(t, err.Error(), "uncore entity is nil")
	})

	t.Run("nil parsed events", func(t *testing.T) {
		mCoreEntity := &CoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}
		mUncoreEntity := &UncoreEventEntity{parsedEvents: []*eventWithQuals{nil, nil}}

		err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "parsed core event is nil")

		err = mResolver.resolveEntities(nil, []*UncoreEventEntity{mUncoreEntity})

		require.Error(t, err)
		require.Contains(t, err.Error(), "parsed uncore event is nil")
	})

	t.Run("fail to resolve core events", func(t *testing.T) {
		name := "mock event 1"
		mCoreEntity := &CoreEventEntity{parsedEvents: []*eventWithQuals{{name: name}}, allEvents: false}
		matcher := ia.NewNameMatcher(name)

		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
		err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to resolve core event `%s`", name))
		mTransformer.AssertExpectations(t)
	})

	t.Run("fail to resolve uncore events", func(t *testing.T) {
		name := "mock event 1"
		mUncoreEntity := &UncoreEventEntity{parsedEvents: []*eventWithQuals{{name: name}}, allEvents: false}
		matcher := ia.NewNameMatcher(name)

		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
		err := mResolver.resolveEntities(nil, []*UncoreEventEntity{mUncoreEntity})

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to resolve uncore event `%s`", name))
		mTransformer.AssertExpectations(t)
	})

	t.Run("resolve all core and uncore events", func(t *testing.T) {
		mCoreEntity := &CoreEventEntity{allEvents: true}
		mUncoreEntity := &UncoreEventEntity{allEvents: true}
		corePerfEvents := []*ia.PerfEvent{
			{Name: "core event1"},
			{Name: "core event2"},
			{Name: "core event3"},
		}
		uncorePerfEvents := []*ia.PerfEvent{
			{Name: "uncore event1", Uncore: true},
			{Name: "uncore event2", Uncore: true},
			{Name: "uncore event3", Uncore: true},
		}
		matcher := ia.NewNameMatcher()

		t.Run("fail to resolve all core events", func(t *testing.T) {
			mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
			err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to resolve all events")
			mTransformer.AssertExpectations(t)
		})

		t.Run("fail to resolve all uncore events", func(t *testing.T) {
			mTransformer.On("Transform", nil, matcher).Once().Return(nil, errMock)
			err := mResolver.resolveEntities(nil, []*UncoreEventEntity{mUncoreEntity})
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to resolve all events")
			mTransformer.AssertExpectations(t)
		})

		t.Run("fail to resolve all events with transformationError", func(t *testing.T) {
			transformErr := &ia.TransformationError{}

			mTransformer.On("Transform", nil, matcher).Once().Return(corePerfEvents, transformErr).Once()
			mTransformer.On("Transform", nil, matcher).Once().Return(uncorePerfEvents, transformErr).Once()

			err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, []*UncoreEventEntity{mUncoreEntity})
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

		err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, []*UncoreEventEntity{mUncoreEntity})
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
		mOptions, _ := ia.NewOptions().SetAttrModifiers(mQuals).Build()
		eventName := "uncore event 1"

		testCase := test{event: &eventWithQuals{name: eventName, qualifiers: mQuals},
			options:   mOptions,
			perfEvent: &ia.PerfEvent{Name: eventName, Uncore: true}}

		matcher := ia.NewNameMatcher(eventName)
		mTransformer.On("Transform", nil, matcher).Return([]*ia.PerfEvent{testCase.perfEvent}, nil).Once()

		mCoreEntity := &CoreEventEntity{parsedEvents: []*eventWithQuals{testCase.event}, allEvents: false}
		err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("uncore event `%s` found in core entity", eventName))
		mTransformer.AssertExpectations(t)
	})

	t.Run("core event found in uncore entity", func(t *testing.T) {
		mQuals := []string{"config1=0x23h"}
		mOptions, _ := ia.NewOptions().SetAttrModifiers(mQuals).Build()
		eventName := "core event 1"

		testCase := test{event: &eventWithQuals{name: eventName, qualifiers: mQuals},
			options:   mOptions,
			perfEvent: &ia.PerfEvent{Name: eventName, Uncore: false}}

		matcher := ia.NewNameMatcher(eventName)
		mTransformer.On("Transform", nil, matcher).Return([]*ia.PerfEvent{testCase.perfEvent}, nil).Once()

		mUncoreEntity := &UncoreEventEntity{parsedEvents: []*eventWithQuals{testCase.event}, allEvents: false}
		err := mResolver.resolveEntities(nil, []*UncoreEventEntity{mUncoreEntity})

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("core event `%s` found in uncore entity", eventName))
		mTransformer.AssertExpectations(t)
	})

	t.Run("resolve core and uncore events", func(t *testing.T) {
		var mCoreEvents []*eventWithQuals
		var nUncoreEvents []*eventWithQuals

		mQuals := []string{"config1=0x23h"}
		mOptions, _ := ia.NewOptions().SetAttrModifiers(mQuals).Build()
		emptyOptions, _ := ia.NewOptions().Build()

		coreTestCases := []test{
			{event: &eventWithQuals{name: "core1", qualifiers: mQuals},
				options:   mOptions,
				perfEvent: &ia.PerfEvent{Name: "core1"}},
			{event: &eventWithQuals{name: "core2", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &ia.PerfEvent{Name: "core2"}},
			{event: &eventWithQuals{name: "core3", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &ia.PerfEvent{Name: "core3"}},
		}
		uncoreTestCases := []test{
			{event: &eventWithQuals{name: "uncore1", qualifiers: mQuals},
				options:   mOptions,
				perfEvent: &ia.PerfEvent{Name: "uncore1", Uncore: true}},
			{event: &eventWithQuals{name: "uncore2", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &ia.PerfEvent{Name: "uncore2", Uncore: true}},
			{event: &eventWithQuals{name: "uncore3", qualifiers: nil},
				options:   emptyOptions,
				perfEvent: &ia.PerfEvent{Name: "uncore3", Uncore: true}},
		}

		for _, test := range coreTestCases {
			matcher := ia.NewNameMatcher(test.event.name)
			mTransformer.On("Transform", nil, matcher).Return([]*ia.PerfEvent{test.perfEvent}, nil).Once()
			mCoreEvents = append(mCoreEvents, test.event)
		}

		for _, test := range uncoreTestCases {
			matcher := ia.NewNameMatcher(test.event.name)
			mTransformer.On("Transform", nil, matcher).Return([]*ia.PerfEvent{test.perfEvent}, nil).Once()
			nUncoreEvents = append(nUncoreEvents, test.event)
		}

		mCoreEntity := &CoreEventEntity{parsedEvents: mCoreEvents, allEvents: false}
		mUncoreEntity := &UncoreEventEntity{parsedEvents: nUncoreEvents, allEvents: false}
		err := mResolver.resolveEntities([]*CoreEventEntity{mCoreEntity}, []*UncoreEventEntity{mUncoreEntity})

		require.NoError(t, err)
		for _, test := range append(coreTestCases, uncoreTestCases...) {
			require.Equal(t, test.perfEvent, test.event.custom.Event)
			require.Equal(t, test.options, test.event.custom.Options)
		}
		mTransformer.AssertExpectations(t)
	})
}

func TestResolveAllEvents(t *testing.T) {
	mTransformer := &MockTransformer{}

	mResolver := &iaEntitiesResolver{transformer: mTransformer}

	t.Run("transformer is nil", func(t *testing.T) {
		mResolver := &iaEntitiesResolver{transformer: nil}
		_, _, err := mResolver.resolveAllEvents()
		require.Error(t, err)
	})

	t.Run("transformer returns error", func(t *testing.T) {
		matcher := ia.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errors.New("mock error"))

		_, _, err := mResolver.resolveAllEvents()
		require.Error(t, err)
		mTransformer.AssertExpectations(t)
	})

	t.Run("no events", func(t *testing.T) {
		matcher := ia.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, nil)

		_, _, err := mResolver.resolveAllEvents()
		require.NoError(t, err)
		mTransformer.AssertExpectations(t)
	})

	t.Run("successfully resolved events", func(t *testing.T) {
		perfEvent1 := &ia.PerfEvent{Name: "mock1"}
		perfEvent2 := &ia.PerfEvent{Name: "mock2"}
		uncorePerfEvent1 := &ia.PerfEvent{Name: "mock3", Uncore: true}
		uncorePerfEvent2 := &ia.PerfEvent{Name: "mock4", Uncore: true}

		options, _ := ia.NewOptions().Build()
		perfEvents := []*ia.PerfEvent{perfEvent1, perfEvent2, uncorePerfEvent1, uncorePerfEvent2}

		expectedCore := []*eventWithQuals{
			{name: perfEvent1.Name, custom: ia.CustomizableEvent{Event: perfEvent1, Options: options}},
			{name: perfEvent2.Name, custom: ia.CustomizableEvent{Event: perfEvent2, Options: options}},
		}

		expectedUncore := []*eventWithQuals{
			{name: uncorePerfEvent1.Name, custom: ia.CustomizableEvent{Event: uncorePerfEvent1, Options: options}},
			{name: uncorePerfEvent2.Name, custom: ia.CustomizableEvent{Event: uncorePerfEvent2, Options: options}},
		}

		matcher := ia.NewNameMatcher()
		mTransformer.On("Transform", nil, matcher).Once().Return(perfEvents, nil)

		coreEvents, uncoreEvents, err := mResolver.resolveAllEvents()
		require.NoError(t, err)
		require.Equal(t, expectedCore, coreEvents)
		require.Equal(t, expectedUncore, uncoreEvents)

		mTransformer.AssertExpectations(t)
	})
}

func TestResolveEvent(t *testing.T) {
	mTransformer := &MockTransformer{}
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
		matcher := ia.NewNameMatcher(mEvent)
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, errors.New("mock error"))

		_, err := mResolver.resolveEvent(mEvent, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to transform perf events")
		mTransformer.AssertExpectations(t)
	})

	t.Run("no events transformed", func(t *testing.T) {
		matcher := ia.NewNameMatcher(mEvent)
		mTransformer.On("Transform", nil, matcher).Once().Return(nil, nil)

		_, err := mResolver.resolveEvent(mEvent, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve unknown event")
		mTransformer.AssertExpectations(t)
	})

	t.Run("not valid qualifiers", func(t *testing.T) {
		event := "mock event 1"
		qualifiers := []string{"wrong modifiers"}

		matcher := ia.NewNameMatcher(event)
		mPerfEvent := &ia.PerfEvent{Name: event}
		mPerfEvents := []*ia.PerfEvent{mPerfEvent}
		mTransformer.On("Transform", nil, matcher).Once().Return(mPerfEvents, nil)

		_, err := mResolver.resolveEvent(event, qualifiers)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to build options for event `%s`", event))
		mTransformer.AssertExpectations(t)
	})

	t.Run("successfully transformed", func(t *testing.T) {
		event := "mock event 1"
		qualifiers := []string{"config1=0x012h", "config2=0x034k"}

		matcher := ia.NewNameMatcher(event)

		mPerfEvent := &ia.PerfEvent{Name: event}
		mPerfEvents := []*ia.PerfEvent{mPerfEvent}

		expectedOptions, _ := ia.NewOptions().SetAttrModifiers(qualifiers).Build()

		mTransformer.On("Transform", nil, matcher).Once().Return(mPerfEvents, nil)

		customEvent, err := mResolver.resolveEvent(event, qualifiers)
		require.NoError(t, err)
		require.Equal(t, mPerfEvent, customEvent.Event)
		require.Equal(t, expectedOptions, customEvent.Options)
		mTransformer.AssertExpectations(t)
	})
}
