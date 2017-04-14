package interrupts

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseInterrupts(t *testing.T) {
	interruptStr := `           CPU0       CPU1
  0:        134          0   IO-APIC-edge      timer
  1:          7          3   IO-APIC-edge      i8042
NMI:          0          0   Non-maskable interrupts
LOC: 2338608687 2334309625   Local timer interrupts
MIS:          0
NET_RX:     867028		225
TASKLET:	205			0`
	f := bytes.NewBufferString(interruptStr)
	parsed := []IRQ{
		IRQ{
			ID: "0", Type: "IO-APIC-edge", Device: "timer",
			Cpus: []int64{int64(134), int64(0)}, Total: int64(134),
		},
		IRQ{
			ID: "1", Type: "IO-APIC-edge", Device: "i8042",
			Cpus: []int64{int64(7), int64(3)}, Total: int64(10),
		},
		IRQ{
			ID: "NMI", Type: "Non-maskable interrupts",
			Cpus: []int64{int64(0), int64(0)}, Total: int64(0),
		},
		IRQ{
			ID: "LOC", Type: "Local timer interrupts",
			Cpus:  []int64{int64(2338608687), int64(2334309625)},
			Total: int64(4672918312),
		},
		IRQ{
			ID: "MIS", Cpus: []int64{int64(0)}, Total: int64(0),
		},
		IRQ{
			ID: "NET_RX", Cpus: []int64{int64(867028), int64(225)},
			Total: int64(867253),
		},
		IRQ{
			ID: "TASKLET", Cpus: []int64{int64(205), int64(0)},
			Total: int64(205),
		},
	}
	got, err := parseInterrupts(f)
	require.Equal(t, nil, err)
	require.NotEqual(t, 0, len(got))
	require.Equal(t, len(got), len(parsed))
	for i := 0; i < len(parsed); i++ {
		assert.Equal(t, parsed[i], got[i])
		for k := 0; k < len(parsed[i].Cpus); k++ {
			assert.Equal(t, parsed[i].Cpus[k], got[i].Cpus[k])
		}
	}
}
