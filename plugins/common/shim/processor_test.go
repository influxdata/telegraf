package shim

import (
	"bufio"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/serializers"
)

func TestProcessorShim(t *testing.T) {
	testSendAndRecieve(t, "f1", "fv1")
}

func TestProcessorShimWithLargerThanDefaultScannerBufferSize(t *testing.T) {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, bufio.MaxScanTokenSize*2)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	testSendAndRecieve(t, "f1", string(b))
}

func testSendAndRecieve(t *testing.T, fieldKey string, fieldValue string) {
	p := &testProcessor{"hi", "mom"}

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

	m := metric.New("thing",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"v":      1,
			fieldKey: fieldValue,
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

	val, ok := mOut.GetTag(p.tagName)
	require.True(t, ok)
	require.Equal(t, p.tagValue, val)
	val2, ok := mOut.Fields()[fieldKey]
	require.True(t, ok)
	require.Equal(t, fieldValue, val2)
	go func() {
		_, _ = io.ReadAll(r)
	}()
	wg.Wait()
}

type testProcessor struct {
	tagName  string
	tagValue string
}

func (p *testProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, m := range in {
		m.AddTag(p.tagName, p.tagValue)
	}
	return in
}

func (p *testProcessor) SampleConfig() string {
	return ""
}

func (p *testProcessor) Description() string {
	return ""
}
