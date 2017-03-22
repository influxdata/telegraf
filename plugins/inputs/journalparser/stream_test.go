package journalparser

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalStreamer_readField(t *testing.T) {
	buf := bytes.NewBuffer([]byte("FOO=fooval\nBAR\n\x06\x00\x00\x00\x00\x00\x00\x00barval\n\n"))
	js := &journalStreamer{
		reader: bufio.NewReader(buf),
	}

	k, v, err := js.readField()
	assert.NoError(t, err)
	assert.Equal(t, "FOO", k)
	assert.Equal(t, "fooval", string(v))

	k, v, err = js.readField()
	assert.NoError(t, err)
	assert.Equal(t, "BAR", k)
	assert.Equal(t, "barval", string(v))

	k, v, err = js.readField()
	assert.NoError(t, err)
	assert.Equal(t, "", k)

	k, v, err = js.readField()
	assert.Error(t, err) // EOF
}

func TestJournalStreamer_readEntry(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`__REALTIME_TIMESTAMP=1492979183632466
FOO=fooval

__REALTIME_TIMESTAMP=1492979183632467
BAR=barval

`))
	js := &journalStreamer{
		reader: bufio.NewReader(buf),
	}

	je, err := js.readEntry()
	require.NoError(t, err)
	assert.WithinDuration(t, time.Unix(0, 1492979183632466000), je.time, time.Duration(0))
	assert.Len(t, je.fields, 2)
	assert.Equal(t, "1492979183632466", string(je.fields["__REALTIME_TIMESTAMP"]))
	assert.Equal(t, "fooval", string(je.fields["FOO"]))

	je, err = js.readEntry()
	require.NoError(t, err)
	assert.WithinDuration(t, time.Unix(0, 1492979183632467000), je.time, time.Duration(0))
	assert.Len(t, je.fields, 2)
	assert.Equal(t, "1492979183632467", string(je.fields["__REALTIME_TIMESTAMP"]))
	assert.Equal(t, "barval", string(je.fields["BAR"]))

	_, err = js.readEntry()
	assert.Error(t, err) // EOF
}

func jcGetEntry(jc *journalClient) *journalEntry {
	select {
	case je := <-jc.jeChan:
		return je
	default:
		return nil
	}
}

func TestJournalStreamer_dispatch(t *testing.T) {
	jc1, _ := newJournalClient([]string{"FOO=fooval"})
	jc2, _ := newJournalClient([]string{"FOO=fooval1", "FOO=fooval2", "BAR=barval"})
	js := &journalStreamer{
		clients: []*journalClient{jc1, jc2},
	}

	je := &journalEntry{
		time: time.Unix(0, 1492979183630000000),
		fields: map[string][]byte{
			"FOO": []byte("fooval"),
			"BAR": []byte("barval"),
		},
	}
	js.dispatch(je)
	je1 := jcGetEntry(jc1)
	assert.Equal(t, je, je1)
	je2 := jcGetEntry(jc2)
	assert.Nil(t, je2)

	je = &journalEntry{
		time: time.Unix(0, 1492979183630000001),
		fields: map[string][]byte{
			"FOO": []byte("fooval1"),
		},
	}
	js.dispatch(je)
	je1 = jcGetEntry(jc1)
	assert.Nil(t, je1)
	je2 = jcGetEntry(jc2)
	assert.Nil(t, je2)

	je = &journalEntry{
		time: time.Unix(0, 1492979183630000002),
		fields: map[string][]byte{
			"FOO": []byte("fooval2"),
			"BAR": []byte("barval"),
		},
	}
	js.dispatch(je)
	je1 = jcGetEntry(jc1)
	assert.Nil(t, je1)
	je2 = jcGetEntry(jc2)
	assert.Equal(t, je, je2)
}

func TestJournalStreamer_run(t *testing.T) {
	buf := []byte(`__REALTIME_TIMESTAMP=1492979183630000
FOO=fooval

__REALTIME_TIMESTAMP=x
FOO=fooval

__REALTIME_TIMESTAMP=1492979183630001
FOO=fooval

`)
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		reader:  bufio.NewReader(bytes.NewBuffer(buf)),
		clients: []*journalClient{jc},
	}

	js.wg.Add(1)
	js.run(nil)

	je := jcGetEntry(jc)
	require.NotNil(t, je)
	assert.WithinDuration(t, time.Unix(0, 1492979183630000000), je.time, time.Duration(0))

	je = jcGetEntry(jc)
	require.NotNil(t, je)
	assert.WithinDuration(t, time.Unix(0, 1492979183630001000), je.time, time.Duration(0))

	assert.Equal(t, je, js.lastEntry)
}

func TestJournalStreamer_run_after(t *testing.T) {
	buf := []byte(`__REALTIME_TIMESTAMP=1492979183630000
FOO=fooval

__REALTIME_TIMESTAMP=x
FOO=fooval

__REALTIME_TIMESTAMP=1492979183630001
FOO=fooval

`)
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		reader:  bufio.NewReader(bytes.NewBuffer(buf)),
		clients: []*journalClient{jc},
	}

	after := time.Unix(0, 1492979183630000000)
	js.wg.Add(1)
	js.run(&after)

	je := jcGetEntry(jc)
	require.NotNil(t, je)
	assert.WithinDuration(t, time.Unix(0, 1492979183630001000), je.time, time.Duration(0))
}

func TestJournalStreamer_start(t *testing.T) {
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		clients: []*journalClient{jc},
	}
	require.NoError(t, js.start())
	defer js.stop()

	je := <-jc.jeChan
	assert.Equal(t, "1", string(je.fields["N"]))
	je = <-jc.jeChan
	assert.Equal(t, "2", string(je.fields["N"]))

	// test adding a new client

	js.stopWait()
	jc2, _ := newJournalClient([]string{"BAR=barval"})
	js.clients = append(js.clients, jc2)
	require.NoError(t, js.start())

	je = <-jc.jeChan
	assert.Equal(t, "4", string(je.fields["N"]))
	je = <-jc2.jeChan
	assert.Equal(t, "5", string(je.fields["N"]))
}

func TestJournalStreamer_start_pathUser(t *testing.T) {
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		path:    "user",
		clients: []*journalClient{jc},
	}
	js.start()
	defer js.stop()

	assert.Contains(t, js.cmd.Args, "--user")
}

func TestJournalStreamer_start_pathDir(t *testing.T) {
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		path:    "/",
		clients: []*journalClient{jc},
	}
	js.start()
	defer js.stop()

	assert.Contains(t, js.cmd.Args, "--directory")
}

func TestJournalStreamer_start_pathFile(t *testing.T) {
	jc, _ := newJournalClient([]string{"FOO=fooval"})
	js := &journalStreamer{
		path:    "/dev/null",
		clients: []*journalClient{jc},
	}
	js.start()
	defer js.stop()

	assert.Contains(t, js.cmd.Args, "--file")
}

func TestJournalStreamer_NewClient(t *testing.T) {
	js := &journalStreamer{}
	jc, err := js.NewClient([]string{"FOO=fooval"})
	require.NoError(t, err)

	assert.NotNil(t, js.cmd.Process)
	require.Len(t, js.clients, 1)
	jc2 := js.clients[0]
	assert.Equal(t, jc, jc2)
	require.Len(t, jc.matches, 1)
	assert.Equal(t, "FOO=fooval", jc.matches[0])
}

func TestJournalStreamer_RemoveClient(t *testing.T) {
	//TODO
}
