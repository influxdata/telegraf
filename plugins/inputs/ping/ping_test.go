package ping

import (
	"context"
	"math"
	"slices"
	"sync"
	"testing"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestNativeIDs(t *testing.T) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 100)

	tests := []struct {
		name        string
		initializer func() []uint16
		expected    func() []int
	}{
		{
			name: "empty",
			initializer: func() []uint16 {
				return make([]uint16, 0)
			},
			expected: func() []int {
				e := make([]int, 10*len(targets))
				for i := range e {
					e[i] = i
				}
				return e
			},
		},
		{
			name: "append",
			initializer: func() []uint16 {
				return []uint16{999}
			},
			expected: func() []int {
				e := make([]int, 10*len(targets))
				for i := range e {
					e[i] = 1000 + i
				}
				return e
			},
		},
		{
			name: "insert",
			initializer: func() []uint16 {
				e := make([]uint16, 0, 10*len(targets)+1)
				for i := range uint16(10 * len(targets)) {
					e = append(e, 2*i+1)
				}
				// We need max at the end to force fill-in
				e = append(e, math.MaxUint16)
				return e
			},
			expected: func() []int {
				e := make([]int, 10*len(targets))
				for i := range e {
					e[i] = 2 * i
				}
				return e
			},
		},
		{
			name: "insert highest",
			initializer: func() []uint16 {
				return []uint16{math.MaxUint16}
			},
			expected: func() []int {
				e := make([]int, 10*len(targets))
				for i := range e {
					e[i] = math.MaxUint16 - len(e) + i
				}
				return e
			},
		},
		{
			name: "complete fill",
			initializer: func() []uint16 {
				e := make([]uint16, 0, math.MaxUint16-10*len(targets))
				for i := range uint16(math.MaxUint16 - 10*len(targets)) {
					e = append(e, i)
				}
				return e
			},
			expected: func() []int {
				e := make([]int, 10*len(targets))
				for i := range e {
					e[i] = math.MaxUint16 - len(e) + i
				}
				return e
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the initial state
			initialIDs := tt.initializer()
			usedIDsCond.L.Lock()
			usedIDs = slices.Clone(initialIDs)
			usedIDsCond.L.Unlock()

			// Add a number of plugin instances that need to share the IDs
			var wg sync.WaitGroup
			var seenIDsMu sync.Mutex
			seenIDs := make([]int, 0, 10*len(targets))
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			for range 10 {
				wg.Add(1)
				go func() {
					defer wg.Done()

					plugin := &Ping{
						Method: "native",
						Urls:   targets,
						Log:    &testutil.Logger{},
						nativePingFunc: func(_ string, id int) (*pingStats, error) {
							seenIDsMu.Lock()
							seenIDs = append(seenIDs, id)
							seenIDsMu.Unlock()

							<-ctx.Done()

							return fakeResult()
						},
					}
					if err := plugin.Init(); err != nil {
						t.Errorf("initializing plugin failed: %v", err)
					}

					var acc testutil.Accumulator
					if err := plugin.Gather(&acc); err != nil {
						t.Errorf("running gather failed: %v", err)
					}
				}()
			}

			// Wait for all plugins to reach the pinging function to ensure we have seen
			// all IDs and all IDs are unique as they are not free'ed in between.
			require.Eventually(t, func() bool {
				seenIDsMu.Lock()
				defer seenIDsMu.Unlock()
				return len(seenIDs) >= 10*len(targets)
			}, 3*time.Second, 100*time.Millisecond)

			// Allow all pings to complete and wait for the plugins to finish
			cancel()
			wg.Wait()

			// Check the seen IDs
			expected := tt.expected()

			seenIDsMu.Lock()
			defer seenIDsMu.Unlock()
			require.Len(t, seenIDs, 10*len(targets))
			slices.Sort(seenIDs)
			require.Equal(t, expected, seenIDs)

			// Check that all acquired IDs have been free'ed
			usedIDsCond.L.Lock()
			defer usedIDsCond.L.Unlock()
			require.Equal(t, initialIDs, usedIDs)
		})
	}
}

func TestNativeIDsWaitOnFull(t *testing.T) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 10)

	// Add some used IDs to force filling in
	initialUsedIDs := make([]uint16, 0, math.MaxUint16)
	for i := range uint16(math.MaxUint16) {
		initialUsedIDs = append(initialUsedIDs, i)
	}
	initialUsedIDs = append(initialUsedIDs, math.MaxUint16)

	usedIDsCond.L.Lock()
	usedIDs = slices.Clone(initialUsedIDs)
	usedIDsCond.L.Unlock()

	// Add a number of plugin instances that need to share the IDs
	var wg sync.WaitGroup
	var seenIDsMu sync.Mutex
	seenIDs := make([]int, 0, len(targets))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()

		plugin := &Ping{
			Method: "native",
			Urls:   targets,
			Log:    &testutil.Logger{},
			nativePingFunc: func(_ string, id int) (*pingStats, error) {
				seenIDsMu.Lock()
				seenIDs = append(seenIDs, id)
				seenIDsMu.Unlock()

				<-ctx.Done()

				return fakeResult()
			},
		}
		if err := plugin.Init(); err != nil {
			t.Errorf("initializing plugin failed: %v", err)
		}

		var acc testutil.Accumulator
		if err := plugin.Gather(&acc); err != nil {
			t.Errorf("running gather failed: %v", err)
		}
	}()

	// All pingers should wait since the IDs are full
	require.Never(t, func() bool {
		seenIDsMu.Lock()
		defer seenIDsMu.Unlock()
		return len(seenIDs) > 0
	}, time.Second, 100*time.Millisecond)

	// Free the required amount of IDs
	for i := range uint16(len(targets)) {
		freeNativePingID(i)
	}

	// Wait for all plugins to reach the pinging function to ensure we have seen
	// all IDs and all IDs are unique as they are not free'ed in between.
	require.Eventually(t, func() bool {
		seenIDsMu.Lock()
		defer seenIDsMu.Unlock()
		return len(seenIDs) >= len(targets)
	}, 3*time.Second, 100*time.Millisecond)

	// Allow all pings to complete and wait for the plugins to finish
	cancel()
	wg.Wait()

	// Check the seen IDs
	expected := make([]int, len(targets))
	for i := range expected {
		expected[i] = i
	}

	seenIDsMu.Lock()
	defer seenIDsMu.Unlock()
	require.Len(t, seenIDs, len(targets))
	slices.Sort(seenIDs)
	require.Equal(t, expected, seenIDs)

	// Check that all IDs have been free'ed
	usedIDsCond.L.Lock()
	defer usedIDsCond.L.Unlock()
	require.Equal(t, initialUsedIDs[len(targets):], usedIDs)
}

func Benchmark(b *testing.B) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 100)

	plugin := &Ping{
		Method: "native",
		Urls:   targets,
		Log:    &testutil.Logger{},
	}
	require.NoError(b, plugin.Init())

	acc := &testutil.Accumulator{Discard: true}
	for b.Loop() {
		require.NoError(b, plugin.Gather(acc))
	}
}

func fakeResult() (*pingStats, error) {
	return &pingStats{
		Statistics: ping.Statistics{
			PacketsSent: 5,
			PacketsRecv: 5,
			Rtts: []time.Duration{
				3 * time.Millisecond,
				4 * time.Millisecond,
				1 * time.Millisecond,
				5 * time.Millisecond,
				2 * time.Millisecond,
			},
		},
		ttl: 1,
	}, nil
}
