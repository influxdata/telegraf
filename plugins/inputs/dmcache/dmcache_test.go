package dmcache

import (
	"errors"
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
	PerDevice:        true,
	getCurrentStatus: output2Devices,
}

func TestDMCacheStats_1(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc1.Gather(&acc)
	require.NoError(t, err)

	tags1 := map[string]string{
		"device": "cs-1",
	}
	fields1 := map[string]interface{}{
		"length":             4883791872,
		"metadata_blocksize": 8,
		"metadata_used":      1018,
		"metadata_total":     1501122,
		"cache_blocksize":    512,
		"cache_used":         7,
		"cache_total":        464962,
		"read_hits":          139,
		"read_misses":        352643,
		"write_hits":         15,
		"write_misses":       46,
		"demotions":          0,
		"promotions":         7,
		"dirty":              0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields1, tags1)

	tags2 := map[string]string{
		"device": "cs-2",
	}
	fields2 := map[string]interface{}{
		"length":             4294967296,
		"metadata_blocksize": 8,
		"metadata_used":      72352,
		"metadata_total":     1310720,
		"cache_blocksize":    128,
		"cache_used":         26,
		"cache_total":        24327168,
		"read_hits":          2409,
		"read_misses":        286,
		"write_hits":         265,
		"write_misses":       524682,
		"demotions":          0,
		"promotions":         0,
		"dirty":              0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields2, tags2)

	tags3 := map[string]string{
		"device": "all",
	}

	fields3 := map[string]interface{}{
		"length":             9178759168,
		"metadata_blocksize": 16,
		"metadata_used":      73370,
		"metadata_total":     2811842,
		"cache_blocksize":    640,
		"cache_used":         33,
		"cache_total":        24792130,
		"read_hits":          2548,
		"read_misses":        352929,
		"write_hits":         280,
		"write_misses":       524728,
		"demotions":          0,
		"promotions":         7,
		"dirty":              0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields3, tags3)
}

var dmc2 = &DMCache{
	PerDevice:        false,
	getCurrentStatus: output2Devices,
}

func TestDMCacheStats_2(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc2.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"device": "all",
	}

	fields := map[string]interface{}{
		"length":             9178759168,
		"metadata_blocksize": 16,
		"metadata_used":      73370,
		"metadata_total":     2811842,
		"cache_blocksize":    640,
		"cache_used":         33,
		"cache_total":        24792130,
		"read_hits":          2548,
		"read_misses":        352929,
		"write_hits":         280,
		"write_misses":       524728,
		"demotions":          0,
		"promotions":         7,
		"dirty":              0,
	}
	acc.AssertContainsTaggedFields(t, "dmcache", fields, tags)
}

func outputNoDevices() ([]string, error) {
	return []string{}, nil
}

var dmc3 = &DMCache{
	PerDevice:        true,
	getCurrentStatus: outputNoDevices,
}

func TestDMCacheStats_3(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc3.Gather(&acc)
	require.NoError(t, err)
}

func noDMSetup() ([]string, error) {
	return []string{}, errors.New("dmsetup doesn't exist")
}

var dmc4 = &DMCache{
	PerDevice:        true,
	getCurrentStatus: noDMSetup,
}

func TestDMCacheStats_4(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc4.Gather(&acc)
	require.Error(t, err)
}

func badFormat() ([]string, error) {
	return []string{
		"cs-1: 0 4883791872 cache 8 1018/1501122 512 7/464962 139 352643 ",
	}, nil
}

var dmc5 = &DMCache{
	PerDevice:        true,
	getCurrentStatus: badFormat,
}

func TestDMCacheStats_5(t *testing.T) {
	var acc testutil.Accumulator

	err := dmc5.Gather(&acc)
	require.Error(t, err)
}
