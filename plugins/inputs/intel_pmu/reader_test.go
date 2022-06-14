//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"fmt"
	"math"
	"testing"
	"time"

	ia "github.com/intel/iaevents"
	"github.com/stretchr/testify/require"
)

type moonClock struct{}

func (moonClock) now() time.Time {
	return time.Date(1969, 7, 20, 20, 17, 0, 0, time.UTC)
}

type eventWithValues struct {
	activeEvent *ia.ActiveEvent
	values      ia.CounterValue
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
		entity := &CoreEventEntity{}

		entity.activeEvents = append(entity.activeEvents, nil)
		metrics, err := mEntitiesReader.readCoreEvents(entity)

		require.Error(t, err)
		require.Contains(t, err.Error(), "active event or corresponding perf event is nil")
		require.Nil(t, metrics)
	})

	t.Run("reading failed", func(t *testing.T) {
		errMock := fmt.Errorf("mock error")
		event := &ia.ActiveEvent{PerfEvent: &ia.PerfEvent{Name: "event1"}}

		entity := &CoreEventEntity{}

		entity.activeEvents = append(entity.activeEvents, event)
		mReader.On("readValue", event).Return(ia.CounterValue{}, errMock).Once()

		metrics, err := mEntitiesReader.readCoreEvents(entity)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to read core event `%s` values: %v", event, errMock))
		require.Nil(t, metrics)
		mReader.AssertExpectations(t)
	})

	t.Run("read active events values", func(t *testing.T) {
		entity := &CoreEventEntity{}
		var expected []coreMetric

		tEvents := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: &ia.PerfEvent{Name: "event1"}}, ia.CounterValue{Raw: 316, Enabled: 182060524, Running: 182060524}},
			{&ia.ActiveEvent{PerfEvent: &ia.PerfEvent{Name: "event2"}}, ia.CounterValue{Raw: 1238901, Enabled: 18234123, Running: 18234123}},
			{&ia.ActiveEvent{PerfEvent: &ia.PerfEvent{Name: "event3"}}, ia.CounterValue{Raw: 412323, Enabled: 1823132, Running: 1823180}},
		}

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
		errMock := fmt.Errorf("mock error")
		perfEvent := &ia.PerfEvent{Name: "event"}

		event := &ia.ActiveEvent{PerfEvent: perfEvent}
		multi := multiEvent{perfEvent: perfEvent, activeEvents: []*ia.ActiveEvent{event}}

		mReader.On("readValue", event).Return(ia.CounterValue{}, errMock).Once()

		metrics, err := mEntitiesReader.readMultiEventSeparately(multi)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("failed to read uncore event `%s` values: %v", event, errMock))
		require.Nil(t, metrics)
		mReader.AssertExpectations(t)
	})

	t.Run("read active events values", func(t *testing.T) {
		perfEvent := &ia.PerfEvent{Name: "event", PMUName: "pmu name"}
		multi := multiEvent{perfEvent: perfEvent}
		var expected []uncoreMetric

		tEvents := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 316, Enabled: 182060524, Running: 182060524}},
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 1238901, Enabled: 18234123, Running: 18234123}},
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 412323, Enabled: 1823132, Running: 1823180}},
		}

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
	errMock := fmt.Errorf("mock error")

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

	perfEvent := &ia.PerfEvent{Name: "event", PMUName: "pmu name"}

	tests := []struct {
		name     string
		multi    multiEvent
		events   []eventWithValues
		result   ia.CounterValue
		readFail bool
		errMsg   string
	}{
		{
			name:   "no events",
			multi:  multiEvent{perfEvent: perfEvent},
			events: nil,
			result: ia.CounterValue{},
			errMsg: "no active events or perf event is nil",
		},
		{
			name:   "no perf event",
			multi:  multiEvent{perfEvent: nil, activeEvents: []*ia.ActiveEvent{{}, {}}},
			events: nil,
			result: ia.CounterValue{},
			errMsg: "no active events or perf event is nil",
		},
		{
			name:  "successful reading and aggregation",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 5123, Enabled: 1231242, Running: 41123}},
				{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 4500, Enabled: 1823423, Running: 182343}},
			},
			result: ia.CounterValue{Raw: 9623, Enabled: 3054665, Running: 223466},
			errMsg: "",
		},
		{
			name:  "to big numbers",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: math.MaxUint64, Enabled: 0, Running: 0}},
				{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 1, Enabled: 0, Running: 0}},
			},
			result: ia.CounterValue{},
			errMsg: fmt.Sprintf("cannot aggregate `%s` values, uint64 exceeding", perfEvent),
		},
		{
			name:  "reading fail",
			multi: multiEvent{perfEvent: perfEvent},
			events: []eventWithValues{
				{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 0, Enabled: 0, Running: 0}},
			},
			readFail: true,
			result:   ia.CounterValue{},
			errMsg:   "failed to read uncore event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, eventWithValue := range test.events {
				test.multi.activeEvents = append(test.multi.activeEvents, eventWithValue.activeEvent)
				if test.readFail {
					mReader.On("readValue", eventWithValue.activeEvent).Return(ia.CounterValue{}, errMock).Once()
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
	errMock := fmt.Errorf("mock error")

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

		perfEvent := &ia.PerfEvent{Name: "mock event", PMUName: "cbox", PMUTypes: []ia.NamedPMUType{{Name: "cbox"}}}
		perfEvent2 := &ia.PerfEvent{Name: "mock event2", PMUName: "rad", PMUTypes: []ia.NamedPMUType{{Name: "rad2"}}}

		multi := multiEvent{perfEvent: perfEvent}
		events := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 2003}},
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 4005}},
		}
		multi2 := multiEvent{perfEvent: perfEvent2}
		events2 := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: perfEvent2}, ia.CounterValue{Raw: 2003}},
			{&ia.ActiveEvent{PerfEvent: perfEvent2}, ia.CounterValue{Raw: 123005}},
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
			values:   ia.CounterValue{Raw: 6008, Enabled: 0, Running: 0},
			socket:   multi.socket,
			unitType: perfEvent.PMUName,
			name:     perfEvent.Name,
			time:     mTimer.now(),
		}
		newMetric2 := uncoreMetric{
			values:   ia.CounterValue{Raw: 125008, Enabled: 0, Running: 0},
			socket:   multi2.socket,
			unitType: perfEvent2.PMUName,
			name:     perfEvent2.Name,
			time:     mTimer.now(),
		}
		expected := []uncoreMetric{newMetric, newMetric2}
		entityAgg := &UncoreEventEntity{Aggregate: true, activeMultiEvents: []multiEvent{multi, multi2}}

		metrics, err := mEntitiesReader.readUncoreEvents(entityAgg)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)

		t.Run("reading error", func(t *testing.T) {
			event := &ia.ActiveEvent{PerfEvent: perfEvent}
			multi := multiEvent{perfEvent: perfEvent, activeEvents: []*ia.ActiveEvent{event}}

			mReader.On("readValue", event).Return(ia.CounterValue{}, errMock).Once()

			entityAgg := &UncoreEventEntity{Aggregate: true, activeMultiEvents: []multiEvent{multi}}
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

		perfEvent := &ia.PerfEvent{Name: "mock event", PMUName: "cbox", PMUTypes: []ia.NamedPMUType{{Name: "cbox"}}}
		perfEvent2 := &ia.PerfEvent{Name: "mock event2", PMUName: "rad", PMUTypes: []ia.NamedPMUType{{Name: "rad2"}}}

		multi := multiEvent{perfEvent: perfEvent, socket: 2}
		events := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 2003}},
			{&ia.ActiveEvent{PerfEvent: perfEvent}, ia.CounterValue{Raw: 4005}},
		}
		multi2 := multiEvent{perfEvent: perfEvent2, socket: 1}
		events2 := []eventWithValues{
			{&ia.ActiveEvent{PerfEvent: perfEvent2}, ia.CounterValue{Raw: 2003}},
			{&ia.ActiveEvent{PerfEvent: perfEvent2}, ia.CounterValue{Raw: 123005}},
		}
		var expected []uncoreMetric
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
		entity := &UncoreEventEntity{activeMultiEvents: []multiEvent{multi, multi2}}

		metrics, err := mEntitiesReader.readUncoreEvents(entity)

		require.NoError(t, err)
		require.Equal(t, expected, metrics)
		mReader.AssertExpectations(t)

		t.Run("reading error", func(t *testing.T) {
			event := &ia.ActiveEvent{PerfEvent: perfEvent}
			multi := multiEvent{perfEvent: perfEvent, activeEvents: []*ia.ActiveEvent{event}}

			mReader.On("readValue", event).Return(ia.CounterValue{}, errMock).Once()

			entityAgg := &UncoreEventEntity{activeMultiEvents: []multiEvent{multi}}
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
		values := ia.CounterValue{}
		socket := 0

		corePerfEvent := &ia.PerfEvent{Name: "core event 1", PMUName: "cpu"}
		activeCoreEvent := []*ia.ActiveEvent{{PerfEvent: corePerfEvent}}
		coreMetric1 := coreMetric{values: values, name: corePerfEvent.Name, time: mTimer.now()}

		corePerfEvent2 := &ia.PerfEvent{Name: "core event 2", PMUName: "cpu"}
		activeCoreEvent2 := []*ia.ActiveEvent{{PerfEvent: corePerfEvent2}}
		coreMetric2 := coreMetric{values: values, name: corePerfEvent2.Name, time: mTimer.now()}

		uncorePerfEvent := &ia.PerfEvent{Name: "uncore event 1", PMUName: "cbox"}
		activeUncoreEvent := []*ia.ActiveEvent{{PerfEvent: uncorePerfEvent}}
		uncoreMetric1 := uncoreMetric{
			values:   values,
			name:     uncorePerfEvent.Name,
			unitType: uncorePerfEvent.PMUName,
			socket:   socket,
			time:     mTimer.now(),
		}

		uncorePerfEvent2 := &ia.PerfEvent{Name: "uncore event 2", PMUName: "rig"}
		activeUncoreEvent2 := []*ia.ActiveEvent{{PerfEvent: uncorePerfEvent2}}
		uncoreMetric2 := uncoreMetric{
			values:   values,
			name:     uncorePerfEvent2.Name,
			unitType: uncorePerfEvent2.PMUName,
			socket:   socket,
			time:     mTimer.now(),
		}

		coreEntities := []*CoreEventEntity{{activeEvents: activeCoreEvent}, {activeEvents: activeCoreEvent2}}

		uncoreEntities := []*UncoreEventEntity{
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
		coreEntities := []*CoreEventEntity{nil}
		coreMetrics, uncoreMetrics, err := mEntitiesReader.readEntities(coreEntities, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, coreMetrics)
		require.Nil(t, uncoreMetrics)
	})

	t.Run("uncore entity reading failed", func(t *testing.T) {
		uncoreEntities := []*UncoreEventEntity{nil}
		coreMetrics, uncoreMetrics, err := mEntitiesReader.readEntities(nil, uncoreEntities)

		require.Error(t, err)
		require.Contains(t, err.Error(), "entity is nil")
		require.Nil(t, coreMetrics)
		require.Nil(t, uncoreMetrics)
	})
}
