//go:build linux && amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/intel/iaevents"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type moonClock struct{}

func (moonClock) now() time.Time {
	return time.Date(1969, 7, 20, 20, 17, 0, 0, time.UTC)
}

type eventWithValues struct {
	activeEvent *iaevents.ActiveEvent
	values      iaevents.CounterValue
}

func TestReadCoreEvents(t *testing.T) {
	mReader := &mockValuesReader{}
	mTimer := &moonClock{}
	mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}

	t.Run("event reader is nil", func(t *testing.T) {
		metrics, err := (&iaEntitiesValuesReader{timer: moonClock{}}).readCoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
		require.Nil(t, metrics)
	})

	t.Run("timer is nil", func(t *testing.T) {
		metrics, err := (&iaEntitiesValuesReader{eventReader: &iaValuesReader{}}).readCoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
		require.Nil(t, metrics)
	})

	t.Run("entity is nil", func(t *testing.T) {
		metrics, err := (&iaEntitiesValuesReader{eventReader: &iaValuesReader{}, timer: moonClock{}}).readCoreEvents(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, metrics)
	})

	t.Run("nil events", func(t *testing.T) {
		entity := &coreEventEntity{}

		entity.activeEvents = append(entity.activeEvents, nil)
		metrics, err := mEntitiesReader.readCoreEvents(entity)

		require.Error(t, err)
		require.Contains(t, err.Error(), "active event or corresponding perf event is nil")
		require.Nil(t, metrics)
	})

	t.Run("reading failed", func(t *testing.T) {
		errMock := errors.New("mock error")
		event := &iaevents.ActiveEvent{PerfEvent: &iaevents.PerfEvent{Name: "event1"}}

		entity := &coreEventEntity{}

		entity.activeEvents = append(entity.activeEvents, event)
		mReader.On("readValue", event).Return(iaevents.CounterValue{}, errMock).Once()

		metrics, err := mEntitiesReader.readCoreEvents(entity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to read core event %q values: %v", event, errMock))
		require.Nil(t, metrics)
		mReader.AssertExpectations(t)
	})

	t.Run("read active events values", func(t *testing.T) {
		entity := &coreEventEntity{}

		tEvents := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: &iaevents.PerfEvent{Name: "event1"}}, iaevents.CounterValue{Raw: 316, Enabled: 182060524, Running: 182060524}},
			{&iaevents.ActiveEvent{PerfEvent: &iaevents.PerfEvent{Name: "event2"}}, iaevents.CounterValue{Raw: 1238901, Enabled: 18234123, Running: 18234123}},
			{&iaevents.ActiveEvent{PerfEvent: &iaevents.PerfEvent{Name: "event3"}}, iaevents.CounterValue{Raw: 412323, Enabled: 1823132, Running: 1823180}},
		}

		expected := make([]coreMetric, 0, len(tEvents))
		for _, tc := range tEvents {
			entity.activeEvents = append(entity.activeEvents, tc.activeEvent)
			cpu, _ := tc.activeEvent.PMUPlacement()
			newMetric := coreMetric{
				values: tc.values,
				tag:    entity.EventsTag,
				cpu:    cpu,
				name:   tc.activeEvent.PerfEvent.Name,
				time:   mTimer.now(),
			}
			expected = append(expected, newMetric)
			mReader.On("readValue", tc.activeEvent).Return(tc.values, nil).Once()
		}
		metrics, err := mEntitiesReader.readCoreEvents(entity)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)
	})
}

func TestReadMultiEventSeparately(t *testing.T) {
	mReader := &mockValuesReader{}
	mTimer := &moonClock{}
	mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}

	t.Run("event reader is nil", func(t *testing.T) {
		event := multiEvent{}
		metrics, err := (&iaEntitiesValuesReader{timer: moonClock{}}).readMultiEventSeparately(event)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
		require.Nil(t, metrics)
	})

	t.Run("timer is nil", func(t *testing.T) {
		event := multiEvent{}
		metrics, err := (&iaEntitiesValuesReader{eventReader: &iaValuesReader{}}).readMultiEventSeparately(event)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
		require.Nil(t, metrics)
	})

	t.Run("multi event is nil", func(t *testing.T) {
		event := multiEvent{}
		metrics, err := (&iaEntitiesValuesReader{&iaValuesReader{}, moonClock{}}).readMultiEventSeparately(event)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no active events or perf event is nil")
		require.Nil(t, metrics)
	})

	t.Run("reading failed", func(t *testing.T) {
		errMock := errors.New("mock error")
		perfEvent := &iaevents.PerfEvent{Name: "event"}

		event := &iaevents.ActiveEvent{PerfEvent: perfEvent}
		multi := multiEvent{perfEvent: perfEvent, activeEvents: []*iaevents.ActiveEvent{event}}

		mReader.On("readValue", event).Return(iaevents.CounterValue{}, errMock).Once()

		metrics, err := mEntitiesReader.readMultiEventSeparately(multi)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to read uncore event %q values: %v", event, errMock))
		require.Nil(t, metrics)
		mReader.AssertExpectations(t)
	})

	t.Run("read active events values", func(t *testing.T) {
		perfEvent := &iaevents.PerfEvent{Name: "event", PMUName: "pmu name"}
		multi := multiEvent{perfEvent: perfEvent}

		tEvents := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 316, Enabled: 182060524, Running: 182060524}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 1238901, Enabled: 18234123, Running: 18234123}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 412323, Enabled: 1823132, Running: 1823180}},
		}

		expected := make([]uncoreMetric, 0, len(tEvents))
		for _, tc := range tEvents {
			multi.activeEvents = append(multi.activeEvents, tc.activeEvent)
			newMetric := uncoreMetric{
				values:   tc.values,
				socket:   multi.socket,
				unitType: multi.perfEvent.PMUName,
				name:     multi.perfEvent.Name,
				unit:     tc.activeEvent.PMUName(),
				time:     mTimer.now(),
			}
			expected = append(expected, newMetric)
			mReader.On("readValue", tc.activeEvent).Return(tc.values, nil).Once()
		}
		metrics, err := mEntitiesReader.readMultiEventSeparately(multi)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)
	})
}

func TestReadMultiEventAgg(t *testing.T) {
	mReader := &mockValuesReader{}
	mTimer := &moonClock{}
	mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}
	errMock := errors.New("mock error")

	t.Run("event reader is nil", func(t *testing.T) {
		event := multiEvent{}
		_, err := (&iaEntitiesValuesReader{timer: moonClock{}}).readMultiEventAgg(event)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
	})

	t.Run("timer is nil", func(t *testing.T) {
		event := multiEvent{}
		_, err := (&iaEntitiesValuesReader{eventReader: &iaValuesReader{}}).readMultiEventAgg(event)
		require.Error(t, err)
		require.Contains(t, err.Error(), "event values reader or timer is nil")
	})

	perfEvent := &iaevents.PerfEvent{Name: "event", PMUName: "pmu name"}

	tests := []struct {
		name     string
		multi    multiEvent
		events   []eventWithValues
		result   iaevents.CounterValue
		readFail bool
		errMsg   string
	}{
		{
			name:   "no events",
			multi:  multiEvent{perfEvent: perfEvent},
			events: nil,
			result: iaevents.CounterValue{},
			errMsg: "no active events or perf event is nil",
		},
		{
			name:   "no perf event",
			multi:  multiEvent{perfEvent: nil, activeEvents: []*iaevents.ActiveEvent{{}, {}}},
			events: nil,
			result: iaevents.CounterValue{},
			errMsg: "no active events or perf event is nil",
		},
		{
			name:  "successful reading and aggregation",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 5123, Enabled: 1231242, Running: 41123}},
				{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 4500, Enabled: 1823423, Running: 182343}},
			},
			result: iaevents.CounterValue{Raw: 9623, Enabled: 3054665, Running: 223466},
			errMsg: "",
		},
		{
			name:  "to big numbers",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: math.MaxUint64, Enabled: 0, Running: 0}},
				{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 1, Enabled: 0, Running: 0}},
			},
			result: iaevents.CounterValue{},
			errMsg: fmt.Sprintf("cannot aggregate %q values, uint64 exceeding", perfEvent),
		},
		{
			name:  "reading fail",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 0, Enabled: 0, Running: 0}},
			},
			readFail: true,
			result:   iaevents.CounterValue{},
			errMsg:   "failed to read uncore event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, eventWithValue := range test.events {
				test.multi.activeEvents = append(test.multi.activeEvents, eventWithValue.activeEvent)
				if test.readFail {
					mReader.On("readValue", eventWithValue.activeEvent).Return(iaevents.CounterValue{}, errMock).Once()
					continue
				}
				mReader.On("readValue", eventWithValue.activeEvent).Return(eventWithValue.values, nil).Once()
			}
			metric, err := mEntitiesReader.readMultiEventAgg(test.multi)
			mReader.AssertExpectations(t)

			if len(test.errMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errMsg)
				return
			}
			expected := uncoreMetric{
				values:   test.result,
				socket:   test.multi.socket,
				unitType: test.multi.perfEvent.PMUName,
				name:     test.multi.perfEvent.Name,
				time:     mTimer.now(),
			}
			require.NoError(t, err)
			require.Equal(t, expected, metric)
		})
	}
}

func TestReadUncoreEvents(t *testing.T) {
	errMock := errors.New("mock error")

	t.Run("entity is nil", func(t *testing.T) {
		metrics, err := (&iaEntitiesValuesReader{}).readUncoreEvents(nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, metrics)
	})

	t.Run("read aggregated entities", func(t *testing.T) {
		mReader := &mockValuesReader{}
		mTimer := &moonClock{}
		mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}

		perfEvent := &iaevents.PerfEvent{Name: "mock event", PMUName: "cbox", PMUTypes: []iaevents.NamedPMUType{{Name: "cbox"}}}
		perfEvent2 := &iaevents.PerfEvent{Name: "mock event2", PMUName: "rad", PMUTypes: []iaevents.NamedPMUType{{Name: "rad2"}}}

		multi := multiEvent{perfEvent: perfEvent}
		events := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 2003}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 4005}},
		}
		multi2 := multiEvent{perfEvent: perfEvent2}
		events2 := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: perfEvent2}, iaevents.CounterValue{Raw: 2003}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent2}, iaevents.CounterValue{Raw: 123005}},
		}
		for _, event := range events {
			multi.activeEvents = append(multi.activeEvents, event.activeEvent)
			mReader.On("readValue", event.activeEvent).Return(event.values, nil).Once()
		}
		for _, event := range events2 {
			multi2.activeEvents = append(multi2.activeEvents, event.activeEvent)
			mReader.On("readValue", event.activeEvent).Return(event.values, nil).Once()
		}
		newMetric := uncoreMetric{
			values:   iaevents.CounterValue{Raw: 6008, Enabled: 0, Running: 0},
			socket:   multi.socket,
			unitType: perfEvent.PMUName,
			name:     perfEvent.Name,
			time:     mTimer.now(),
		}
		newMetric2 := uncoreMetric{
			values:   iaevents.CounterValue{Raw: 125008, Enabled: 0, Running: 0},
			socket:   multi2.socket,
			unitType: perfEvent2.PMUName,
			name:     perfEvent2.Name,
			time:     mTimer.now(),
		}
		expected := []uncoreMetric{newMetric, newMetric2}
		entityAgg := &uncoreEventEntity{Aggregate: true, activeMultiEvents: []multiEvent{multi, multi2}}

		metrics, err := mEntitiesReader.readUncoreEvents(entityAgg)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)

		t.Run("reading error", func(t *testing.T) {
			event := &iaevents.ActiveEvent{PerfEvent: perfEvent}
			multi := multiEvent{perfEvent: perfEvent, activeEvents: []*iaevents.ActiveEvent{event}}

			mReader.On("readValue", event).Return(iaevents.CounterValue{}, errMock).Once()

			entityAgg := &uncoreEventEntity{Aggregate: true, activeMultiEvents: []multiEvent{multi}}
			metrics, err = mEntitiesReader.readUncoreEvents(entityAgg)

			require.Error(t, err)
			require.Nil(t, metrics)
			mReader.AssertExpectations(t)
		})
	})

	t.Run("read distributed entities", func(t *testing.T) {
		mReader := &mockValuesReader{}
		mTimer := &moonClock{}
		mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}

		perfEvent := &iaevents.PerfEvent{Name: "mock event", PMUName: "cbox", PMUTypes: []iaevents.NamedPMUType{{Name: "cbox"}}}
		perfEvent2 := &iaevents.PerfEvent{Name: "mock event2", PMUName: "rad", PMUTypes: []iaevents.NamedPMUType{{Name: "rad2"}}}

		multi := multiEvent{perfEvent: perfEvent, socket: 2}
		events := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 2003}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent}, iaevents.CounterValue{Raw: 4005}},
		}
		multi2 := multiEvent{perfEvent: perfEvent2, socket: 1}
		events2 := []eventWithValues{
			{&iaevents.ActiveEvent{PerfEvent: perfEvent2}, iaevents.CounterValue{Raw: 2003}},
			{&iaevents.ActiveEvent{PerfEvent: perfEvent2}, iaevents.CounterValue{Raw: 123005}},
		}

		expected := make([]uncoreMetric, 0, len(events))
		for _, event := range events {
			multi.activeEvents = append(multi.activeEvents, event.activeEvent)
			mReader.On("readValue", event.activeEvent).Return(event.values, nil).Once()

			newMetric := uncoreMetric{
				values:   event.values,
				socket:   multi.socket,
				unitType: perfEvent.PMUName,
				name:     perfEvent.Name,
				unit:     event.activeEvent.PMUName(),
				time:     mTimer.now(),
			}
			expected = append(expected, newMetric)
		}
		for _, event := range events2 {
			multi2.activeEvents = append(multi2.activeEvents, event.activeEvent)
			mReader.On("readValue", event.activeEvent).Return(event.values, nil).Once()

			newMetric := uncoreMetric{
				values:   event.values,
				socket:   multi2.socket,
				unitType: perfEvent2.PMUName,
				name:     perfEvent2.Name,
				unit:     event.activeEvent.PMUName(),
				time:     mTimer.now(),
			}
			expected = append(expected, newMetric)
		}
		entity := &uncoreEventEntity{activeMultiEvents: []multiEvent{multi, multi2}}

		metrics, err := mEntitiesReader.readUncoreEvents(entity)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)

		t.Run("reading error", func(t *testing.T) {
			event := &iaevents.ActiveEvent{PerfEvent: perfEvent}
			multi := multiEvent{perfEvent: perfEvent, activeEvents: []*iaevents.ActiveEvent{event}}

			mReader.On("readValue", event).Return(iaevents.CounterValue{}, errMock).Once()

			entityAgg := &uncoreEventEntity{activeMultiEvents: []multiEvent{multi}}
			metrics, err = mEntitiesReader.readUncoreEvents(entityAgg)

			require.Error(t, err)
			require.Nil(t, metrics)
			mReader.AssertExpectations(t)
		})
	})
}

func TestReadEntities(t *testing.T) {
	mReader := &mockValuesReader{}
	mTimer := &moonClock{}
	mEntitiesReader := &iaEntitiesValuesReader{mReader, mTimer}

	t.Run("read entities", func(t *testing.T) {
		values := iaevents.CounterValue{}
		socket := 0

		corePerfEvent := &iaevents.PerfEvent{Name: "core event 1", PMUName: "cpu"}
		activeCoreEvent := []*iaevents.ActiveEvent{{PerfEvent: corePerfEvent}}
		coreMetric1 := coreMetric{values: values, name: corePerfEvent.Name, time: mTimer.now()}

		corePerfEvent2 := &iaevents.PerfEvent{Name: "core event 2", PMUName: "cpu"}
		activeCoreEvent2 := []*iaevents.ActiveEvent{{PerfEvent: corePerfEvent2}}
		coreMetric2 := coreMetric{values: values, name: corePerfEvent2.Name, time: mTimer.now()}

		uncorePerfEvent := &iaevents.PerfEvent{Name: "uncore event 1", PMUName: "cbox"}
		activeUncoreEvent := []*iaevents.ActiveEvent{{PerfEvent: uncorePerfEvent}}
		uncoreMetric1 := uncoreMetric{
			values:   values,
			name:     uncorePerfEvent.Name,
			unitType: uncorePerfEvent.PMUName,
			socket:   socket,
			time:     mTimer.now(),
		}

		uncorePerfEvent2 := &iaevents.PerfEvent{Name: "uncore event 2", PMUName: "rig"}
		activeUncoreEvent2 := []*iaevents.ActiveEvent{{PerfEvent: uncorePerfEvent2}}
		uncoreMetric2 := uncoreMetric{
			values:   values,
			name:     uncorePerfEvent2.Name,
			unitType: uncorePerfEvent2.PMUName,
			socket:   socket,
			time:     mTimer.now(),
		}

		coreEntities := []*coreEventEntity{{activeEvents: activeCoreEvent}, {activeEvents: activeCoreEvent2}}

		uncoreEntities := []*uncoreEventEntity{
			{activeMultiEvents: []multiEvent{{activeEvents: activeUncoreEvent, perfEvent: uncorePerfEvent, socket: socket}}},
			{activeMultiEvents: []multiEvent{{activeEvents: activeUncoreEvent2, perfEvent: uncorePerfEvent2, socket: socket}}},
		}

		expectedCoreMetrics := []coreMetric{coreMetric1, coreMetric2}
		expectedUncoreMetrics := []uncoreMetric{uncoreMetric1, uncoreMetric2}

		mReader.On("readValue", activeCoreEvent[0]).Return(values, nil).Once()
		mReader.On("readValue", activeCoreEvent2[0]).Return(values, nil).Once()
		mReader.On("readValue", activeUncoreEvent[0]).Return(values, nil).Once()
		mReader.On("readValue", activeUncoreEvent2[0]).Return(values, nil).Once()

		coreMetrics, uncoreMetrics, err := mEntitiesReader.readEntities(coreEntities, uncoreEntities)

		require.NoError(t, err)
		require.Equal(t, expectedCoreMetrics, coreMetrics)
		require.NotNil(t, expectedUncoreMetrics, uncoreMetrics)
		mReader.AssertExpectations(t)
	})

	t.Run("core entity reading failed", func(t *testing.T) {
		coreEntities := []*coreEventEntity{nil}
		coreMetrics, uncoreMetrics, err := mEntitiesReader.readEntities(coreEntities, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, coreMetrics)
		require.Nil(t, uncoreMetrics)
	})

	t.Run("uncore entity reading failed", func(t *testing.T) {
		uncoreEntities := []*uncoreEventEntity{nil}
		coreMetrics, uncoreMetrics, err := mEntitiesReader.readEntities(nil, uncoreEntities)

		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, coreMetrics)
		require.Nil(t, uncoreMetrics)
	})
}

// Mocking

// mockValuesReader is an autogenerated mock type for the valuesReader type
type mockValuesReader struct {
	mock.Mock
}

// readValue provides a mock function with given fields: event
func (_m *mockValuesReader) readValue(event *iaevents.ActiveEvent) (iaevents.CounterValue, error) {
	ret := _m.Called(event)

	var r0 iaevents.CounterValue
	if rf, ok := ret.Get(0).(func(*iaevents.ActiveEvent) iaevents.CounterValue); ok {
		r0 = rf(event)
	} else {
		r0 = ret.Get(0).(iaevents.CounterValue)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*iaevents.ActiveEvent) error); ok {
		r1 = rf(event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
