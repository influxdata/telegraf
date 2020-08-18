package shim

import (
	"bufio"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/stretchr/testify/require"
)

func TestProcessorShim(t *testing.T) {
	p := &testProcessor{}

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	s := New()
	// inject test into shim
	s.stdin = stdinReader
	s.stdout = stdoutWriter
	err := s.AddProcessor(p)
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := s.RunProcessor()
		require.NoError(t, err)
		wg.Done()
	}()

	serializer, _ := serializers.NewInfluxSerializer()
	parser, _ := parsers.NewInfluxParser()

	m, _ := metric.New("thing",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"v": 1,
		},
		time.Now(),
	)
	b, err := serializer.Serialize(m)
	require.NoError(t, err)
	_, err = stdinWriter.Write(b)
	require.NoError(t, err)
	err = stdinWriter.Close()
	require.NoError(t, err)

	r := bufio.NewReader(stdoutReader)
	out, err := r.ReadString('\n')
	require.NoError(t, err)
	mOut, err := parser.ParseLine(out)
	require.NoError(t, err)

	val, ok := mOut.GetTag("hi")
	require.True(t, ok)
	require.Equal(t, "mom", val)

	go ioutil.ReadAll(r)
	wg.Wait()
}

type testProcessor struct{}

func (p *testProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		metric.AddTag("hi", "mom")
	}
	return in
}

func (p *testProcessor) SampleConfig() string {
	return ""
}

func (p *testProcessor) Description() string {
	return ""
}
