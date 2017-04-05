package irqstat

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseInterrupts(t *testing.T) {
	include := []string{}
	interruptStr := `           CPU0       CPU1       
  0:        134          0   IO-APIC-edge      timer
  1:          7          3   IO-APIC-edge      i8042
NMI:          0          0   Non-maskable interrupts
LOC: 2338608687 2334309625   Local timer interrupts
MIS:          0
NET_RX:     867028		225
TASKLET:	205			0`

	parsed := []IRQ{
		IRQ{ID: "0", Type: "IO-APIC-edge", Device: "timer", Values: map[string]interface{}{"CPU0": int64(134), "CPU1": int64(0), "total": int64(134)}},
		IRQ{ID: "1", Type: "IO-APIC-edge", Device: "i8042", Values: map[string]interface{}{"CPU0": int64(7), "CPU1": int64(3), "total": int64(10)}},
		IRQ{ID: "NMI", Type: "Non-maskable interrupts", Values: map[string]interface{}{"CPU0": int64(0), "CPU1": int64(0), "total": int64(0)}},
		IRQ{ID: "LOC", Type: "Local timer interrupts", Values: map[string]interface{}{"CPU0": int64(2338608687), "CPU1": int64(2334309625), "total": int64(4672918312)}},
		IRQ{ID: "MIS", Values: map[string]interface{}{"CPU0": int64(0), "total": int64(0)}},
		IRQ{ID: "NET_RX", Values: map[string]interface{}{"CPU0": int64(867028), "CPU1": int64(225), "total": int64(867253)}},
		IRQ{ID: "TASKLET", Values: map[string]interface{}{"CPU0": int64(205), "CPU1": int64(0), "total": int64(205)}},
	}
	got, err := parseInterrupts(interruptStr, include)
	require.Equal(t, nil, err)
	require.NotEqual(t, 0, len(got))
	require.Equal(t, len(got), len(parsed))
	for i := 0; i < len(parsed); i++ {
		assert.Equal(t, parsed[i], got[i])
		for k, _ := range parsed[i].Values {
			assert.Equal(t, parsed[i].Values[k], got[i].Values[k])
		}
	}
}
