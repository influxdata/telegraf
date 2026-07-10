package concat

import (
	"bytes"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// newRand returns a deterministically seeded RNG so test runs are reproducible.
// The seed is logged so a failing run can be replayed.
func newRand(t *testing.T) *mathrand.Rand {
	t.Helper()
	const seed = 42
	t.Logf("random seed: %d", seed)
	return mathrand.New(mathrand.NewSource(seed))
}

func randomBytes(t *testing.T, rng *mathrand.Rand, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rng.Read(b)
	require.NoError(t, err)
	return b
}

func randomMessages(t *testing.T, rng *mathrand.Rand, n, minLen, maxLen int, constBytes []byte) [][]byte {
	t.Helper()
	messages := make([][]byte, n)
	for i := 0; i < n; i++ {
		length := minLen + rng.Intn(maxLen-minLen+1) + 6
		lenbytes := make([]byte, 2)
		lenbytes[0] = uint8(length >> 8)
		lenbytes[1] = uint8(length & 0xff)
		messages[i] = append(constBytes, uint8(i), lenbytes[0], lenbytes[1], 0)
		checksum := uint8(0)
		for j := 0; j < 5; j++ {
			checksum ^= messages[i][j]
		}
		messages[i] = append(messages[i], checksum)
		messages[i] = append(messages[i], randomBytes(t, rng, length-6)...)
	}
	return messages
}

// rawParser is a minimal embedded parser that wraps each message payload in a
// metric so tests can verify the bytes the splitter handed over.
type rawParser struct{}

func (*rawParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	m := metric.New("raw", nil, map[string]interface{}{"payload": string(buf)}, time.Unix(0, 0))
	return []telegraf.Metric{m}, nil
}

func (*rawParser) ParseLine(line string) (telegraf.Metric, error) {
	return metric.New("raw", nil, map[string]interface{}{"payload": line}, time.Unix(0, 0)), nil
}

func (*rawParser) SetDefaultTags(map[string]string) {}

func payload(t *testing.T, m telegraf.Metric) string {
	t.Helper()
	v, ok := m.GetField("payload")
	require.True(t, ok, "metric is missing the payload field")
	s, ok := v.(string)
	require.True(t, ok, "payload field is not a string")
	return s
}

func TestConcatParserDelim(t *testing.T) {
	p := &Parser{
		Splitter: Splitter{
			HeaderLength: 6,
			MessageLengthField: LengthFieldConfig{
				Offset:     2,
				Bytes:      0,
				Endianness: "be",
			},
			ConstBytes: map[string]string{
				"0": "0x73",
				"1": "0x74",
				"2": "0x61",
				"3": "0x72",
				"4": "0x74",
			},
		},
	}
	p.SetParser(&rawParser{})
	require.NoError(t, p.Init())

	rng := newRand(t)

	t.Run("single message", func(t *testing.T) {
		constBytes := []byte{'s', 't', 'a', 'r', 't'}
		msgs := randomMessages(t, rng, 2, 10, 20, constBytes)
		// In delimiter mode a message is only emitted once the next header is
		// seen, so feed a following message to flush the first.
		metrics, err := p.Parse(msgs[0])
		require.NoError(t, err)
		require.Empty(t, metrics, "expected no metrics before next header")

		metrics, err = p.Parse(msgs[1])
		require.NoError(t, err)
		require.Len(t, metrics, 1)
		require.Equal(t, string(msgs[0][6:]), payload(t, metrics[0]))
	})

	t.Run("multiple messages", func(t *testing.T) {
		const (
			n              = 10000
			minLen         = 100
			maxLen         = 2000
			allowedDropped = 0.1
			allowedGhost   = 0.1
		)
		constBytes := []byte{'s', 't', 'a', 'r', 't'}
		messages := randomMessages(t, rng, n, minLen, maxLen, constBytes)

		stream := bytes.Join(messages, nil)
		var metrics []telegraf.Metric
		for len(stream) > 0 {
			pkt := rng.Intn(2048) + 1
			if pkt > len(stream) {
				pkt = len(stream)
			}
			parsed, err := p.Parse(stream[:pkt])
			require.NoError(t, err)
			metrics = append(metrics, parsed...)
			stream = stream[pkt:]
		}
		// Reconcile emitted metrics against the source messages. Payloads are
		// random variable-length blobs.
		dropped := 0
		ghostMetrics := 0
		mi := 0
		for _, m := range metrics {
			buf := payload(t, m)
			matched := false
			for j := mi; j < len(messages); j++ {
				if buf == string(messages[j][6:]) {
					dropped += j - mi
					mi = j + 1
					matched = true
					break
				}
			}
			if !matched {
				ghostMetrics++
			}
		}
		// Any messages left unmatched after the last metric were also dropped.
		// In delimiter mode the final message stays buffered until a following
		// header arrives, so it is expected to show up here.
		dropped += len(messages) - mi

		t.Logf("out of: %d, dropped metrics: %d, ghost metrics: %d", n, dropped, ghostMetrics)
		require.LessOrEqualf(t, float64(dropped)/float64(n), allowedDropped,
			"dropped %d of %d metrics", dropped, n)
		require.LessOrEqualf(t, float64(ghostMetrics)/float64(n), allowedGhost,
			"ghost metrics: %d", ghostMetrics)
	})
}
func TestConcatParserLength(t *testing.T) {
	p := &Parser{
		Splitter: Splitter{
			HeaderLength: 6,
			MessageLengthField: LengthFieldConfig{
				Offset:     2,
				Bytes:      2,
				Endianness: "be",
			},
			ConstBytes: map[string]string{
				"0": "0x53",
			},
			Checksum: ChecksumConfig{
				Strategy: "xor",
				Range:    []int{0, 4},
				Offset:   5,
				Bytes:    1,
			},
		},
	}
	p.SetParser(&rawParser{})
	require.NoError(t, p.Init())

	rng := newRand(t)

	t.Run("single message", func(t *testing.T) {
		constBytes := []byte{'S'}
		m := randomMessages(t, rng, 1, 10, 20, constBytes)[0]
		metrics, err := p.Parse(m)
		require.NoError(t, err)
		require.Len(t, metrics, 1)
		require.Equal(t, string(m[6:]), payload(t, metrics[0]))
	})

	t.Run("multiple messages", func(t *testing.T) {
		const (
			n              = 10000
			minLen         = 100
			maxLen         = 200
			allowedDropped = 0.1
			allowedGhost   = 0.1
		)
		messages := randomMessages(t, rng, n, minLen, maxLen, []byte{'S'})
		stream := bytes.Join(messages, nil)
		var metrics []telegraf.Metric
		for len(stream) > 0 {
			pkt := rng.Intn(2048) + 1
			if pkt > len(stream) {
				pkt = len(stream)
			}
			parsed, err := p.Parse(stream[:pkt])
			require.NoError(t, err)
			metrics = append(metrics, parsed...)
			stream = stream[pkt:]
		}
		// Reconcile emitted metrics against the source messages. Payloads are
		// random variable-length blobs.
		dropped := 0
		ghostMetrics := 0
		mi := 0
		for _, m := range metrics {
			buf := payload(t, m)
			matched := false
			for j := mi; j < len(messages); j++ {
				if buf == string(messages[j][6:]) {
					dropped += j - mi
					mi = j + 1
					matched = true
					break
				}
			}
			if !matched {
				ghostMetrics++
			}
		}
		// Any messages left unmatched after the last metric were also dropped.
		dropped += len(messages) - mi

		t.Logf("out of: %d, dropped metrics: %d, ghost metrics: %d", n, dropped, ghostMetrics)
		require.LessOrEqualf(t, float64(dropped)/float64(n), allowedDropped,
			"dropped %d of %d metrics", dropped, n)
		require.LessOrEqualf(t, float64(ghostMetrics)/float64(n), allowedGhost,
			"ghost metrics: %d", ghostMetrics)
	})
}
