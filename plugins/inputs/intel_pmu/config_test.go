//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	ia "github.com/intel/iaevents"
	"github.com/stretchr/testify/require"
)

func TestConfigParser_parseEntities(t *testing.T) {
	mSysInfo := &mockSysInfoProvider{}
	mConfigParser := &configParser{
		sys: mSysInfo,
		log: testutil.Logger{},
	}
	e := ia.CustomizableEvent{}

	t.Run("no entities", func(t *testing.T) {
		err := mConfigParser.parseEntities(nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "neither core nor uncore entities configured")
	})

	// more specific parsing cases in TestConfigParser_parseIntRanges and TestConfigParser_parseEvents
	coreTests := []struct {
		name string

		coreEntity       *CoreEventEntity
		parsedCoreEvents []*eventWithQuals
		parsedCores      []int
		coreAll          bool

		uncoreEntity       *UncoreEventEntity
		parsedUncoreEvents []*eventWithQuals
		parsedSockets      []int
		uncoreAll          bool

		failMsg string
	}{
		{"no events provided",
			&CoreEventEntity{Events: nil, Cores: []string{"1"}}, nil, []int{1}, true,
			&UncoreEventEntity{Events: nil, Sockets: []string{"0"}}, nil, []int{0}, true,
			""},
		{"uncore entity is nil",
			&CoreEventEntity{Events: []string{"EVENT"}, Cores: []string{"1,2"}}, []*eventWithQuals{{"EVENT", nil, e}}, []int{1, 2}, false,
			nil, nil, nil, false,
			"uncore entity is nil"},
		{"core entity is nil",
			nil, nil, nil, false,
			&UncoreEventEntity{Events: []string{"EVENT"}, Sockets: []string{"1,2"}}, []*eventWithQuals{{"EVENT", nil, e}}, []int{1, 2}, false,
			"core entity is nil"},
		{"error parsing sockets",
			&CoreEventEntity{Events: nil, Cores: []string{"1,2"}}, nil, []int{1, 2}, true,
			&UncoreEventEntity{Events: []string{"E"}, Sockets: []string{"wrong sockets"}}, []*eventWithQuals{{"E", nil, e}}, nil, false,
			"error during sockets parsing"},
		{"error parsing cores",
			&CoreEventEntity{Events: nil, Cores: []string{"wrong cpus"}}, nil, nil, true,
			&UncoreEventEntity{Events: nil, Sockets: []string{"0,1"}}, nil, []int{0, 1}, true,
			"error during cores parsing"},
		{"valid settings",
			&CoreEventEntity{Events: []string{"E1", "E2:config=123"}, Cores: []string{"1-5"}}, []*eventWithQuals{{"E1", nil, e}, {"E2", []string{"config=123"}, e}}, []int{1, 2, 3, 4, 5}, false,
			&UncoreEventEntity{Events: []string{"E1", "E2", "E3"}, Sockets: []string{"0,2-6"}}, []*eventWithQuals{{"E1", nil, e}, {"E2", nil, e}, {"E3", nil, e}}, []int{0, 2, 3, 4, 5, 6}, false,
			""},
	}

	for _, test := range coreTests {
		t.Run(test.name, func(t *testing.T) {
			coreEntities := []*CoreEventEntity{test.coreEntity}
			uncoreEntities := []*UncoreEventEntity{test.uncoreEntity}

			err := mConfigParser.parseEntities(coreEntities, uncoreEntities)

			if len(test.failMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.failMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.coreAll, test.coreEntity.allEvents)
			require.Equal(t, test.parsedCores, test.coreEntity.parsedCores)
			require.Equal(t, test.parsedCoreEvents, test.coreEntity.parsedEvents)

			require.Equal(t, test.uncoreAll, test.uncoreEntity.allEvents)
			require.Equal(t, test.parsedSockets, test.uncoreEntity.parsedSockets)
			require.Equal(t, test.parsedUncoreEvents, test.uncoreEntity.parsedEvents)
		})
	}
}

func TestConfigParser_parseCores(t *testing.T) {
	mSysInfo := &mockSysInfoProvider{}
	mConfigParser := &configParser{
		sys: mSysInfo,
		log: testutil.Logger{},
	}

	t.Run("no cores provided", func(t *testing.T) {
		t.Run("system info provider is nil", func(t *testing.T) {
			result, err := (&configParser{}).parseCores(nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "system info provider is nil")
			require.Nil(t, result)
		})
		t.Run("cannot gather all cpus info", func(t *testing.T) {
			mSysInfo.On("allCPUs").Return(nil, errors.New("all cpus error")).Once()
			result, err := mConfigParser.parseCores(nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot obtain all cpus")
			require.Nil(t, result)
			mSysInfo.AssertExpectations(t)
		})
		t.Run("all cpus gathering succeeded", func(t *testing.T) {
			allCPUs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

			mSysInfo.On("allCPUs").Return(allCPUs, nil).Once()
			result, err := mConfigParser.parseCores(nil)
			require.NoError(t, err)
			require.Equal(t, allCPUs, result)
			mSysInfo.AssertExpectations(t)
		})
	})
}

func TestConfigParser_parseSockets(t *testing.T) {
	mSysInfo := &mockSysInfoProvider{}
	mConfigParser := &configParser{
		sys: mSysInfo,
		log: testutil.Logger{},
	}

	t.Run("no sockets provided", func(t *testing.T) {
		t.Run("system info provider is nil", func(t *testing.T) {
			result, err := (&configParser{}).parseSockets(nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "system info provider is nil")
			require.Nil(t, result)
		})
		t.Run("cannot gather all sockets info", func(t *testing.T) {
			mSysInfo.On("allSockets").Return(nil, errors.New("all sockets error")).Once()
			result, err := mConfigParser.parseSockets(nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot obtain all sockets")
			require.Nil(t, result)
			mSysInfo.AssertExpectations(t)
		})
		t.Run("all cpus gathering succeeded", func(t *testing.T) {
			allSockets := []int{0, 1, 2, 3, 4}

			mSysInfo.On("allSockets").Return(allSockets, nil).Once()
			result, err := mConfigParser.parseSockets(nil)
			require.NoError(t, err)
			require.Equal(t, allSockets, result)
			mSysInfo.AssertExpectations(t)
		})
	})
}

func TestConfigParser_parseEvents(t *testing.T) {
	mConfigParser := &configParser{log: testutil.Logger{}}
	e := ia.CustomizableEvent{}

	tests := []struct {
		name   string
		input  []string
		result []*eventWithQuals
	}{
		{"no events", nil, nil},
		{"single string", []string{"mock string"}, []*eventWithQuals{{"mock string", nil, e}}},
		{"two events", []string{"EVENT.FIRST", "EVENT.SECOND"}, []*eventWithQuals{{"EVENT.FIRST", nil, e}, {"EVENT.SECOND", nil, e}}},
		{"event with configs", []string{"EVENT.SECOND:config1=0x404300k:config2=0x404300k"},
			[]*eventWithQuals{{"EVENT.SECOND", []string{"config1=0x404300k", "config2=0x404300k"}, e}}},
		{"two events with modifiers", []string{"EVENT.FIRST:config1=0x200300:config2=0x231100:u:H", "EVENT.SECOND:K:p"},
			[]*eventWithQuals{{"EVENT.FIRST", []string{"config1=0x200300", "config2=0x231100", "u", "H"}, e}, {"EVENT.SECOND", []string{"K", "p"}, e}}},
		{"duplicates", []string{"EVENT1", "EVENT1", "EVENT2"}, []*eventWithQuals{{"EVENT1", nil, e}, {"EVENT2", nil, e}}},
		{"duplicates with different configs", []string{"EVENT1:config1", "EVENT1:config2"},
			[]*eventWithQuals{{"EVENT1", []string{"config1"}, e}, {"EVENT1", []string{"config2"}, e}}},
		{"duplicates with the same modifiers", []string{"EVENT1:config1", "EVENT1:config1"},
			[]*eventWithQuals{{"EVENT1", []string{"config1"}, e}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mConfigParser.parseEvents(test.input)
			require.Equal(t, test.result, result)
		})
	}
}

func TestConfigParser_parseIntRanges(t *testing.T) {
	mConfigParser := &configParser{log: testutil.Logger{}}
	tests := []struct {
		name    string
		input   []string
		result  []int
		failMsg string
	}{
		{"coma separated", []string{"0,1,2,3,4"}, []int{0, 1, 2, 3, 4}, ""},
		{"range", []string{"0-10"}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, ""},
		{"mixed", []string{"0-3", "4", "12-16"}, []int{0, 1, 2, 3, 4, 12, 13, 14, 15, 16}, ""},
		{"min and max values", []string{"-2147483648", "2147483647"}, []int{math.MinInt32, math.MaxInt32}, ""},
		{"should remove duplicates", []string{"1-5", "2-6"}, []int{1, 2, 3, 4, 5, 6}, ""},
		{"wrong format", []string{"1,2,3%$S,-100"}, nil, "wrong format for id"},
		{"start is greater than end", []string{"10-3"}, nil, "`10` is equal or greater than `3"},
		{"too big value", []string{"18446744073709551615"}, nil, "wrong format for id"},
		{"too much numbers", []string{fmt.Sprintf("0-%d", maxIDsSize)}, nil,
			fmt.Sprintf("requested number of IDs exceeds max size `%d`", maxIDsSize)},
		{"too much numbers mixed", []string{fmt.Sprintf("1-%d", maxIDsSize), "0"}, nil,
			fmt.Sprintf("requested number of IDs exceeds max size `%d`", maxIDsSize)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := mConfigParser.parseIntRanges(test.input)
			require.Equal(t, test.result, result)
			if len(test.failMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.failMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
