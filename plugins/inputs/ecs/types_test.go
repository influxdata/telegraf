package ecs

import (
	"os"
	"reflect"
	"testing"
)

func Test_parseTask(t *testing.T) {
	t.Run("parseTask", func(t *testing.T) {
		r, _ := os.Open("testdata/metadata.golden")
		parsed, err := unmarshalTask(r)
		if err != nil {
			t.Errorf("error parsing task %v", err)
		}
		if !reflect.DeepEqual(*parsed, validMeta) {
			t.Errorf("Got = %v, want = %v", parsed, validMeta)
		}
	})
}

func Test_parseStats(t *testing.T) {
	t.Run("parseStats", func(t *testing.T) {
		r, _ := os.Open("testdata/stats.golden")
		parsed, err := unmarshalStats(r)
		if err != nil {
			t.Errorf("error parsing stats %v", err)
		}
		if !reflect.DeepEqual(parsed, validStats) {
			t.Errorf("Got = %v, want = %v", parsed, validMeta)
		}
	})
}

func Test_mergeTaskStats(t *testing.T) {
	t.Run("mergeStats", func(t *testing.T) {
		metadata, _ := os.Open("testdata/metadata.golden")
		parsedMetadata, err := unmarshalTask(metadata)
		if err != nil {
			t.Errorf("error parsing task %v", err)
		}

		stats, _ := os.Open("testdata/stats.golden")
		parsedStats, err := unmarshalStats(stats)
		if err != nil {
			t.Errorf("error parsing stats %v", err)
		}

		mergeTaskStats(parsedMetadata, parsedStats)

		for _, cont := range parsedMetadata.Containers {
			if !reflect.DeepEqual(cont.Stats, validStats[cont.ID]) {
				t.Errorf("Got = %v, want = %v", cont.Stats, validStats[cont.ID])
			}
		}
	})
}
