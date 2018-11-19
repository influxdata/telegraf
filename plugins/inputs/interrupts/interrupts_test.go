package interrupts

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{ID: "0", Type: "IO-APIC-edge", Device: "timer", CPU: 0, Count: int64(134)},
		{ID: "0", Type: "IO-APIC-edge", Device: "timer", CPU: 1, Count: int64(0)},
		{ID: "1", Type: "IO-APIC-edge", Device: "i8042", CPU: 0, Count: int64(7)},
		{ID: "1", Type: "IO-APIC-edge", Device: "i8042", CPU: 1, Count: int64(3)},
		{ID: "NMI", Type: "Non-maskable interrupts", CPU: 0, Count: int64(0)},
		{ID: "NMI", Type: "Non-maskable interrupts", CPU: 1, Count: int64(0)},
		{ID: "LOC", Type: "Local timer interrupts", CPU: 0, Count: int64(2338608687)},
		{ID: "LOC", Type: "Local timer interrupts", CPU: 1, Count: int64(2334309625)},
		{ID: "MIS", CPU: 0, Count: int64(0)},
		{ID: "NET_RX", CPU: 0, Count: int64(867028)},
		{ID: "NET_RX", CPU: 1, Count: int64(225)},
		{ID: "TASKLET", CPU: 0, Count: int64(205)},
		{ID: "TASKLET", CPU: 1, Count: int64(0)},
	}
	got, err := parseInterrupts(f)
	require.Equal(t, nil, err)
	require.NotEqual(t, 0, len(got))
	require.Equal(t, len(got), len(parsed))
	for i := 0; i < len(parsed); i++ {
		assert.Equal(t, parsed[i], got[i])
	}
}

// Tests #4470
func TestParseInterruptsBad(t *testing.T) {
	interruptStr := `           CPU0       CPU1       CPU2       CPU3
	16:          0          0          0          0  bcm2836-timer   0 Edge      arch_timer
	17:  127224250  118424219  127224437  117885416  bcm2836-timer   1 Edge      arch_timer
	21:          0          0          0          0  bcm2836-pmu   9 Edge      arm-pmu
	23:    1549514          0          0          0  ARMCTRL-level   1 Edge      3f00b880.mailbox
	24:          2          0          0          0  ARMCTRL-level   2 Edge      VCHIQ doorbell
	46:          0          0          0          0  ARMCTRL-level  48 Edge      bcm2708_fb dma
	48:          0          0          0          0  ARMCTRL-level  50 Edge      DMA IRQ
	50:          0          0          0          0  ARMCTRL-level  52 Edge      DMA IRQ
	51:        208          0          0          0  ARMCTRL-level  53 Edge      DMA IRQ
	54:     883002          0          0          0  ARMCTRL-level  56 Edge      DMA IRQ
	59:          0          0          0          0  ARMCTRL-level  61 Edge      bcm2835-auxirq
	62:  521451447          0          0          0  ARMCTRL-level  64 Edge      dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1
	86:     857597          0          0          0  ARMCTRL-level  88 Edge      mmc0
	87:       4938          0          0          0  ARMCTRL-level  89 Edge      uart-pl011
	92:       5669          0          0          0  ARMCTRL-level  94 Edge      mmc1
   FIQ:              usb_fiq
   IPI0:          0          0          0          0  CPU wakeup interrupts
   IPI1:          0          0          0          0  Timer broadcast interrupts
   IPI2:   23564958   23464876   23531165   23040826  Rescheduling interrupts
   IPI3:     148438     639704     644266     588150  Function call interrupts
   IPI4:          0          0          0          0  CPU stop interrupts
   IPI5:    4348149    1843985    3819457    1822877  IRQ work interrupts
   IPI6:          0          0          0          0  completion interrupts`
	f := bytes.NewBufferString(interruptStr)
	parsed := []IRQ{
		{ID: "16", Type: "bcm2836-timer", Device: "0 Edge arch_timer", CPU: 0, Count: int64(0)},
		{ID: "16", Type: "bcm2836-timer", Device: "0 Edge arch_timer", CPU: 1, Count: int64(0)},
		{ID: "16", Type: "bcm2836-timer", Device: "0 Edge arch_timer", CPU: 2, Count: int64(0)},
		{ID: "16", Type: "bcm2836-timer", Device: "0 Edge arch_timer", CPU: 3, Count: int64(0)},
		{ID: "17", Type: "bcm2836-timer", Device: "1 Edge arch_timer", CPU: 0, Count: int64(127224250)},
		{ID: "17", Type: "bcm2836-timer", Device: "1 Edge arch_timer", CPU: 1, Count: int64(118424219)},
		{ID: "17", Type: "bcm2836-timer", Device: "1 Edge arch_timer", CPU: 2, Count: int64(127224437)},
		{ID: "17", Type: "bcm2836-timer", Device: "1 Edge arch_timer", CPU: 3, Count: int64(117885416)},
		{ID: "21", Type: "bcm2836-pmu", Device: "9 Edge arm-pmu", CPU: 0, Count: int64(0)},
		{ID: "21", Type: "bcm2836-pmu", Device: "9 Edge arm-pmu", CPU: 1, Count: int64(0)},
		{ID: "21", Type: "bcm2836-pmu", Device: "9 Edge arm-pmu", CPU: 2, Count: int64(0)},
		{ID: "21", Type: "bcm2836-pmu", Device: "9 Edge arm-pmu", CPU: 3, Count: int64(0)},
		{ID: "23", Type: "ARMCTRL-level", Device: "1 Edge 3f00b880.mailbox", CPU: 0, Count: int64(1549514)},
		{ID: "23", Type: "ARMCTRL-level", Device: "1 Edge 3f00b880.mailbox", CPU: 1, Count: int64(0)},
		{ID: "23", Type: "ARMCTRL-level", Device: "1 Edge 3f00b880.mailbox", CPU: 2, Count: int64(0)},
		{ID: "23", Type: "ARMCTRL-level", Device: "1 Edge 3f00b880.mailbox", CPU: 3, Count: int64(0)},
		{ID: "24", Type: "ARMCTRL-level", Device: "2 Edge VCHIQ doorbell", CPU: 0, Count: int64(2)},
		{ID: "24", Type: "ARMCTRL-level", Device: "2 Edge VCHIQ doorbell", CPU: 1, Count: int64(0)},
		{ID: "24", Type: "ARMCTRL-level", Device: "2 Edge VCHIQ doorbell", CPU: 2, Count: int64(0)},
		{ID: "24", Type: "ARMCTRL-level", Device: "2 Edge VCHIQ doorbell", CPU: 3, Count: int64(0)},
		{ID: "46", Type: "ARMCTRL-level", Device: "48 Edge bcm2708_fb dma", CPU: 0, Count: int64(0)},
		{ID: "46", Type: "ARMCTRL-level", Device: "48 Edge bcm2708_fb dma", CPU: 1, Count: int64(0)},
		{ID: "46", Type: "ARMCTRL-level", Device: "48 Edge bcm2708_fb dma", CPU: 2, Count: int64(0)},
		{ID: "46", Type: "ARMCTRL-level", Device: "48 Edge bcm2708_fb dma", CPU: 3, Count: int64(0)},
		{ID: "48", Type: "ARMCTRL-level", Device: "50 Edge DMA IRQ", CPU: 0, Count: int64(0)},
		{ID: "48", Type: "ARMCTRL-level", Device: "50 Edge DMA IRQ", CPU: 1, Count: int64(0)},
		{ID: "48", Type: "ARMCTRL-level", Device: "50 Edge DMA IRQ", CPU: 2, Count: int64(0)},
		{ID: "48", Type: "ARMCTRL-level", Device: "50 Edge DMA IRQ", CPU: 3, Count: int64(0)},
		{ID: "50", Type: "ARMCTRL-level", Device: "52 Edge DMA IRQ", CPU: 0, Count: int64(0)},
		{ID: "50", Type: "ARMCTRL-level", Device: "52 Edge DMA IRQ", CPU: 1, Count: int64(0)},
		{ID: "50", Type: "ARMCTRL-level", Device: "52 Edge DMA IRQ", CPU: 2, Count: int64(0)},
		{ID: "50", Type: "ARMCTRL-level", Device: "52 Edge DMA IRQ", CPU: 3, Count: int64(0)},
		{ID: "51", Type: "ARMCTRL-level", Device: "53 Edge DMA IRQ", CPU: 0, Count: int64(208)},
		{ID: "51", Type: "ARMCTRL-level", Device: "53 Edge DMA IRQ", CPU: 1, Count: int64(0)},
		{ID: "51", Type: "ARMCTRL-level", Device: "53 Edge DMA IRQ", CPU: 2, Count: int64(0)},
		{ID: "51", Type: "ARMCTRL-level", Device: "53 Edge DMA IRQ", CPU: 3, Count: int64(0)},
		{ID: "54", Type: "ARMCTRL-level", Device: "56 Edge DMA IRQ", CPU: 0, Count: int64(883002)},
		{ID: "54", Type: "ARMCTRL-level", Device: "56 Edge DMA IRQ", CPU: 1, Count: int64(0)},
		{ID: "54", Type: "ARMCTRL-level", Device: "56 Edge DMA IRQ", CPU: 2, Count: int64(0)},
		{ID: "54", Type: "ARMCTRL-level", Device: "56 Edge DMA IRQ", CPU: 3, Count: int64(0)},
		{ID: "59", Type: "ARMCTRL-level", Device: "61 Edge bcm2835-auxirq", CPU: 0, Count: int64(0)},
		{ID: "59", Type: "ARMCTRL-level", Device: "61 Edge bcm2835-auxirq", CPU: 1, Count: int64(0)},
		{ID: "59", Type: "ARMCTRL-level", Device: "61 Edge bcm2835-auxirq", CPU: 2, Count: int64(0)},
		{ID: "59", Type: "ARMCTRL-level", Device: "61 Edge bcm2835-auxirq", CPU: 3, Count: int64(0)},
		{ID: "62", Type: "ARMCTRL-level", Device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", CPU: 0, Count: int64(521451447)},
		{ID: "62", Type: "ARMCTRL-level", Device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", CPU: 1, Count: int64(0)},
		{ID: "62", Type: "ARMCTRL-level", Device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", CPU: 2, Count: int64(0)},
		{ID: "62", Type: "ARMCTRL-level", Device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", CPU: 3, Count: int64(0)},
		{ID: "86", Type: "ARMCTRL-level", Device: "88 Edge mmc0", CPU: 0, Count: int64(857597)},
		{ID: "86", Type: "ARMCTRL-level", Device: "88 Edge mmc0", CPU: 1, Count: int64(0)},
		{ID: "86", Type: "ARMCTRL-level", Device: "88 Edge mmc0", CPU: 2, Count: int64(0)},
		{ID: "86", Type: "ARMCTRL-level", Device: "88 Edge mmc0", CPU: 3, Count: int64(0)},
		{ID: "87", Type: "ARMCTRL-level", Device: "89 Edge uart-pl011", CPU: 0, Count: int64(4938)},
		{ID: "87", Type: "ARMCTRL-level", Device: "89 Edge uart-pl011", CPU: 1, Count: int64(0)},
		{ID: "87", Type: "ARMCTRL-level", Device: "89 Edge uart-pl011", CPU: 2, Count: int64(0)},
		{ID: "87", Type: "ARMCTRL-level", Device: "89 Edge uart-pl011", CPU: 3, Count: int64(0)},
		{ID: "92", Type: "ARMCTRL-level", Device: "94 Edge mmc1", CPU: 0, Count: int64(5669)},
		{ID: "92", Type: "ARMCTRL-level", Device: "94 Edge mmc1", CPU: 1, Count: int64(0)},
		{ID: "92", Type: "ARMCTRL-level", Device: "94 Edge mmc1", CPU: 2, Count: int64(0)},
		{ID: "92", Type: "ARMCTRL-level", Device: "94 Edge mmc1", CPU: 3, Count: int64(0)},
		{ID: "IPI0", Type: "CPU wakeup interrupts", CPU: 0, Count: int64(0)},
		{ID: "IPI0", Type: "CPU wakeup interrupts", CPU: 1, Count: int64(0)},
		{ID: "IPI0", Type: "CPU wakeup interrupts", CPU: 2, Count: int64(0)},
		{ID: "IPI0", Type: "CPU wakeup interrupts", CPU: 3, Count: int64(0)},
		{ID: "IPI1", Type: "Timer broadcast interrupts", CPU: 0, Count: int64(0)},
		{ID: "IPI1", Type: "Timer broadcast interrupts", CPU: 1, Count: int64(0)},
		{ID: "IPI1", Type: "Timer broadcast interrupts", CPU: 2, Count: int64(0)},
		{ID: "IPI1", Type: "Timer broadcast interrupts", CPU: 3, Count: int64(0)},
		{ID: "IPI2", Type: "Rescheduling interrupts", CPU: 0, Count: int64(23564958)},
		{ID: "IPI2", Type: "Rescheduling interrupts", CPU: 1, Count: int64(23464876)},
		{ID: "IPI2", Type: "Rescheduling interrupts", CPU: 2, Count: int64(23531165)},
		{ID: "IPI2", Type: "Rescheduling interrupts", CPU: 3, Count: int64(23040826)},
		{ID: "IPI3", Type: "Function call interrupts", CPU: 0, Count: int64(148438)},
		{ID: "IPI3", Type: "Function call interrupts", CPU: 1, Count: int64(639704)},
		{ID: "IPI3", Type: "Function call interrupts", CPU: 2, Count: int64(644266)},
		{ID: "IPI3", Type: "Function call interrupts", CPU: 3, Count: int64(588150)},
		{ID: "IPI4", Type: "CPU stop interrupts", CPU: 0, Count: int64(0)},
		{ID: "IPI4", Type: "CPU stop interrupts", CPU: 1, Count: int64(0)},
		{ID: "IPI4", Type: "CPU stop interrupts", CPU: 2, Count: int64(0)},
		{ID: "IPI4", Type: "CPU stop interrupts", CPU: 3, Count: int64(0)},
		{ID: "IPI5", Type: "IRQ work interrupts", CPU: 0, Count: int64(4348149)},
		{ID: "IPI5", Type: "IRQ work interrupts", CPU: 1, Count: int64(1843985)},
		{ID: "IPI5", Type: "IRQ work interrupts", CPU: 2, Count: int64(3819457)},
		{ID: "IPI5", Type: "IRQ work interrupts", CPU: 3, Count: int64(1822877)},
		{ID: "IPI6", Type: "completion interrupts", CPU: 0, Count: int64(0)},
		{ID: "IPI6", Type: "completion interrupts", CPU: 1, Count: int64(0)},
		{ID: "IPI6", Type: "completion interrupts", CPU: 2, Count: int64(0)},
		{ID: "IPI6", Type: "completion interrupts", CPU: 3, Count: int64(0)},
	}
	got, err := parseInterrupts(f)
	require.Equal(t, nil, err)
	require.NotEqual(t, 0, len(got))
	require.Equal(t, len(got), len(parsed))
	for i := 0; i < len(parsed); i++ {
		assert.Equal(t, parsed[i], got[i])
	}
}
