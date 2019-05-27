package ecs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseTask(t *testing.T) {
	r, err := os.Open("testdata/metadata.golden")
	if err != nil {
		t.Errorf("error opening test files")
	}
	parsed, err := unmarshalTask(r)
	if err != nil {
		t.Errorf("error parsing task %v", err)
	}
	assert.Equal(t, validMeta, *parsed, "Got = %v, want = %v", parsed, validMeta)
}

func Test_parseStats(t *testing.T) {
	r, err := os.Open("testdata/stats.golden")
	if err != nil {
		t.Errorf("error opening test files")
	}

	parsed, err := unmarshalStats(r)
	if err != nil {
		t.Errorf("error parsing stats %v", err)
	}
	assert.Equal(t, validStats, parsed, "Got = %v, want = %v", parsed, validStats)
}

func Test_mergeTaskStats(t *testing.T) {
	metadata, err := os.Open("testdata/metadata.golden")
	if err != nil {
		t.Errorf("error opening test files")
	}

	parsedMetadata, err := unmarshalTask(metadata)
	if err != nil {
		t.Errorf("error parsing task %v", err)
	}

	stats, err := os.Open("testdata/stats.golden")
	if err != nil {
		t.Errorf("error opening test files")
	}

	parsedStats, err := unmarshalStats(stats)
	if err != nil {
		t.Errorf("error parsing stats %v", err)
	}

	mergeTaskStats(parsedMetadata, parsedStats)

	for _, cont := range parsedMetadata.Containers {
		assert.Equal(t, validStats[cont.ID], cont.Stats, "Got = %v, want = %v", cont.Stats, validStats[cont.ID])
	}
}
