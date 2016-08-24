// +build linux

package dmcache

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func output2Devices() ([]string, error) {
	return []string{
		"cs-1: 0 4883791872 cache 8 1018/1501122 512 7/464962 139 352643 15 46 0 7 0 1 writeback 2 migration_threshold 2048 mq 10 random_threshold 4 sequential_threshold 512 discard_promote_adjustment 1 read_promote_adjustment 4 write_promote_adjustment 8",
		"cs-2: 0 4294967296 cache 8 72352/1310720 128 26/24327168 2409 286 265 524682 0 0 0 1 writethrough 2 migration_threshold 2048 mq 10 random_threshold 4 sequential_threshold 512 discard_promote_adjustment 1 read_promote_adjustment 4 write_promote_adjustment 8",
	}, nil
}

var dmc1 = &DMCache{
	PerDevice: true,
	rawStatus: output2Devices,
}

func TestDMCacheStats_1(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc1.Gather(&acc)
	require.NoError(t, err)

	tags1 := map[string]string{
		"device": "cs-1",
	}
	fields1 := map[string]interface{}{
		"metadata_used": 4169728,
		"metadata_free": 6144425984,
		"cache_used":    1835008,
		"cache_free":    121885163520,
		"read_hits":     36438016,
		"read_misses":   92443246592,
		"write_hits":    3932160,
		"write_misses":  12058624,
		"demotions":     0,
		"promotions":    1835008,
		"dirty":         0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields1, tags1)

	tags2 := map[string]string{
		"device": "cs-2",
	}
	fields2 := map[string]interface{}{
		"metadata_used": 296353792,
		"metadata_free": 5072355328,
		"cache_used":    1703936,
		"cache_free":    1594303578112,
		"read_hits":     157876224,
		"read_misses":   18743296,
		"write_hits":    17367040,
		"write_misses":  34385559552,
		"demotions":     0,
		"promotions":    0,
		"dirty":         0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields2, tags2)
}

var dmc2 = &DMCache{
	PerDevice: false,
	rawStatus: output2Devices,
}

func TestDMCacheStats_2(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc2.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{}

	fields := map[string]interface{}{
		"metadata_used": 300523520,
		"metadata_free": 11216781312,
		"cache_used":    3538944,
		"cache_free":    1716188741632,
		"read_hits":     194314240,
		"read_misses":   92461989888,
		"write_hits":    21299200,
		"write_misses":  34397618176,
		"demotions":     0,
		"promotions":    1835008,
		"dirty":         0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields, tags)
}
