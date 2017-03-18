package irqstat

import "testing"

func TestParseInterrupts(t *testing.T) {
	include := []string{}
	interruptStr := `           CPU0       CPU1       
  0:        134          0   IO-APIC-edge      timer
  1:          7          3   IO-APIC-edge      i8042
NMI:          0          0   Non-maskable interrupts
LOC: 2338608687 2334309625   Local timer interrupts
MIS:          0`

	parsed := []IRQ{
		IRQ{ID: "0", Fields: map[string]interface{}{"CPU0": int64(134), "CPU1": int64(0)}, Tags: map[string]string{"type": "IO-APIC-edge", "device": "timer"}},
		IRQ{ID: "1", Fields: map[string]interface{}{"CPU0": int64(7), "CPU1": int64(3)}, Tags: map[string]string{"type": "IO-APIC-edge", "device": "i8042"}},
		IRQ{ID: "NMI", Fields: map[string]interface{}{"CPU0": int64(0), "CPU1": int64(0)}, Tags: map[string]string{"type": "Non-maskable interrupts"}},
		IRQ{ID: "LOC", Fields: map[string]interface{}{"CPU0": int64(2338608687), "CPU1": int64(2334309625)}, Tags: map[string]string{"type": "Local timer interrupts"}},
		IRQ{ID: "MIS", Fields: map[string]interface{}{"CPU0": int64(0)}},
	}

	got := parseInterrupts(interruptStr, include)
	if len(got) == 0 {
		t.Fatalf("want %+v, got %+v", parsed, got)
	}
	for i := 0; i < len(parsed); i++ {
		for k, _ := range parsed[i].Fields {
			if parsed[i].Fields[k] != got[i].Fields[k] {
				t.Fatalf("want %+v, got %+v", parsed[i].Fields[k], got[i].Fields[k])
			}
		}
		for k, _ := range parsed[i].Tags {
			if parsed[i].Tags[k] != got[i].Tags[k] {
				t.Fatalf("want %+v, got %+v", parsed[i].Tags[k], got[i].Tags[k])
			}
		}
	}
}
