// +build linux

package journalparser

import (
	"os/exec"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs/logparser/grok"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setNoExecCommand() func() {
	ec := execCommand
	restoreF := func() { execCommand = ec }
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	return restoreF
}

func TestJournalParser(t *testing.T) {
	jp := &JournalParser{
		Matches: []string{"FOO=fooval"},
		GrokParser: JournalGrokParser{
			Parser: grok.Parser{
				Measurement:    "mymeasurement",
				CustomPatterns: "FOOPAT %{WORD:myval}",
			},
			Patterns: map[string][]string{
				"FOO": {"%{FOOPAT}"},
			},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, jp.Start(&acc))
	defer func() { assert.NoError(t, jp.Stop()) }()

	assert.Equal(t, jp.Matches, jp.journalClient.matches)

	acc.Wait(1)

	fields := map[string]interface{}{"myval": "fooval"}
	acc.AssertContainsFields(t, "mymeasurement", fields)
}

func TestJournalParser_customTime(t *testing.T) {
	defer setNoExecCommand()()

	jp := &JournalParser{
		Matches: []string{"FOO=bar"},
		GrokParser: JournalGrokParser{
			Parser: grok.Parser{
				CustomPatterns: "BAZPAT test %{WORD:myval}",
			},
			Patterns: map[string][]string{
				"MESSAGE": {"%{TIMESTAMP_ISO8601:timestamp:ts-rfc3339} %{BAZPAT}"},
			},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, jp.Start(&acc))
	defer func() { assert.NoError(t, jp.Stop()) }()

	jeChan := jp.journalClient.jeChan
	jeChan <- &journalEntry{
		time: time.Now(),
		fields: map[string][]byte{
			"MESSAGE": []byte("2011-02-03T04:05:06.007Z test baz"),
			"FOO":     []byte("bar"),
		},
	}
	acc.Wait(1)

	assert.Empty(t, acc.Errors)

	tags := map[string]string{}
	fields := map[string]interface{}{
		"myval": "baz",
	}
	acc.AssertContainsTaggedFields(t, "journalparser", fields, tags)

	m := acc.Metrics[0]
	exTime, _ := time.Parse("2006-01-02T03:04:05.000Z0700", "2011-02-03T04:05:06.007Z")
	assert.True(t, m.Time.Equal(exTime), "Expected: %s, Actual: %s", exTime.String(), m.Time.String())
}

func TestJournalParser_journalTime(t *testing.T) {
	defer setNoExecCommand()()

	jp := &JournalParser{
		Matches: []string{"FOO=bar"},
		GrokParser: JournalGrokParser{
			Parser: grok.Parser{
				CustomPatterns: "BAZPAT test %{WORD:myval}",
			},
			Patterns: map[string][]string{
				"MESSAGE": {"%{BAZPAT}"},
			},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, jp.Start(&acc))
	defer func() { assert.NoError(t, jp.Stop()) }()

	jeChan := jp.journalClient.jeChan
	now := time.Now()
	jeChan <- &journalEntry{
		time: now,
		fields: map[string][]byte{
			"MESSAGE": []byte("test baz"),
			"FOO":     []byte("bar"),
		},
	}
	acc.Wait(1)

	assert.Empty(t, acc.Errors)

	tags := map[string]string{}
	fields := map[string]interface{}{
		"myval": "baz",
	}
	acc.AssertContainsTaggedFields(t, "journalparser", fields, tags)

	assert.True(t, now.Equal(acc.Metrics[0].Time), "%s != %s", now, acc.Metrics[0].Time)
}
