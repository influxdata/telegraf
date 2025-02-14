package interrupts

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

// =====================================================================================
//	Setup and helper functions
// =====================================================================================

func expectCPUAsTags(m *testutil.Accumulator, t *testing.T, measurement string, irq irq) {
	for idx, value := range irq.cpus {
		m.AssertContainsTaggedFields(t, measurement,
			map[string]interface{}{"count": value},
			map[string]string{"irq": irq.id, "type": irq.typ, "device": irq.device, "cpu": fmt.Sprintf("cpu%d", idx)},
		)
	}
}

func expectCPUAsFields(m *testutil.Accumulator, t *testing.T, measurement string, irq irq) {
	fields := map[string]interface{}{}
	total := int64(0)
	for idx, count := range irq.cpus {
		fields[fmt.Sprintf("CPU%d", idx)] = count
		total += count
	}
	fields["total"] = total

	m.AssertContainsTaggedFields(t, measurement, fields, map[string]string{"irq": irq.id, "type": irq.typ, "device": irq.device})
}

func setup(t *testing.T, irqString string, cpuAsTags bool) (*testutil.Accumulator, []irq) {
	f := bytes.NewBufferString(irqString)
	irqs, err := parseInterrupts(f)
	require.NoError(t, err)
	require.NotEmpty(t, irqs)

	acc := new(testutil.Accumulator)
	reportMetrics("soft_interrupts", irqs, acc, cpuAsTags)

	return acc, irqs
}

// =====================================================================================
//	Soft interrupts
// =====================================================================================

const softIrqsString = `            CPU0       		CPU1
						0:           134 	    	   0   IO-APIC-edge      timer
						1:   	       7	           3   IO-APIC-edge      i8042
						NMI:           0    	   	   0   Non-maskable interrupts
						LOC:  2338608687 	  2334309625   Local timer interrupts
						MIS:           0
						NET_RX:   867028			 225
						TASKLET:	 205			   0`

var softIrqsExpectedArgs = []irq{
	{id: "0", typ: "IO-APIC-edge", device: "timer", cpus: []int64{134, 0}},
	{id: "1", typ: "IO-APIC-edge", device: "i8042", cpus: []int64{7, 3}},
	{id: "NMI", typ: "Non-maskable interrupts", cpus: []int64{0, 0}},
	{id: "MIS", cpus: []int64{0}},
	{id: "NET_RX", cpus: []int64{867028, 225}},
	{id: "TASKLET", cpus: []int64{205, 0}},
}

func TestCpuAsTagsSoftIrqs(t *testing.T) {
	acc, irqs := setup(t, softIrqsString, true)
	reportMetrics("soft_interrupts", irqs, acc, true)

	for _, irq := range softIrqsExpectedArgs {
		expectCPUAsTags(acc, t, "soft_interrupts", irq)
	}
}

func TestCpuAsFieldsSoftIrqs(t *testing.T) {
	acc, irqs := setup(t, softIrqsString, false)
	reportMetrics("soft_interrupts", irqs, acc, false)

	for _, irq := range softIrqsExpectedArgs {
		expectCPUAsFields(acc, t, "soft_interrupts", irq)
	}
}

// =====================================================================================
//	HW interrupts, tests #4470
// =====================================================================================

const hwIrqsString = `     CPU0       CPU1       CPU2       CPU3
				 16:          0          0          0          0  bcm2836-timer   0 Edge      arch_timer
				 17:  127224250  118424219  127224437  117885416  bcm2836-timer   1 Edge      arch_timer
				 21:          0          0          0          0  bcm2836-pmu     9 Edge      arm-pmu
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
				IPI0:         0          0          0          0  CPU wakeup interrupts
				IPI1:         0          0          0          0  Timer broadcast interrupts
				IPI2:  23564958   23464876   23531165   23040826  Rescheduling interrupts
				IPI3:    148438     639704     644266     588150  Function call interrupts
				IPI4:         0          0          0          0  CPU stop interrupts
				IPI5:   4348149    1843985    3819457    1822877  IRQ work interrupts
				IPI6:         0          0          0          0  completion interrupts`

var hwIrqsExpectedArgs = []irq{
	{id: "16", typ: "bcm2836-timer", device: "0 Edge arch_timer", cpus: []int64{0, 0, 0, 0}},
	{id: "17", typ: "bcm2836-timer", device: "1 Edge arch_timer", cpus: []int64{127224250, 118424219, 127224437, 117885416}},
	{id: "21", typ: "bcm2836-pmu", device: "9 Edge arm-pmu", cpus: []int64{0, 0, 0, 0}},
	{id: "23", typ: "ARMCTRL-level", device: "1 Edge 3f00b880.mailbox", cpus: []int64{1549514, 0, 0, 0}},
	{id: "24", typ: "ARMCTRL-level", device: "2 Edge VCHIQ doorbell", cpus: []int64{2, 0, 0, 0}},
	{id: "46", typ: "ARMCTRL-level", device: "48 Edge bcm2708_fb dma", cpus: []int64{0, 0, 0, 0}},
	{id: "48", typ: "ARMCTRL-level", device: "50 Edge DMA IRQ", cpus: []int64{0, 0, 0, 0}},
	{id: "50", typ: "ARMCTRL-level", device: "52 Edge DMA IRQ", cpus: []int64{0, 0, 0, 0}},
	{id: "51", typ: "ARMCTRL-level", device: "53 Edge DMA IRQ", cpus: []int64{208, 0, 0, 0}},
	{id: "54", typ: "ARMCTRL-level", device: "56 Edge DMA IRQ", cpus: []int64{883002, 0, 0, 0}},
	{id: "59", typ: "ARMCTRL-level", device: "61 Edge bcm2835-auxirq", cpus: []int64{0, 0, 0, 0}},
	{id: "62", typ: "ARMCTRL-level", device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", cpus: []int64{521451447, 0, 0, 0}},
	{id: "86", typ: "ARMCTRL-level", device: "88 Edge mmc0", cpus: []int64{857597, 0, 0, 0}},
	{id: "87", typ: "ARMCTRL-level", device: "89 Edge uart-pl011", cpus: []int64{4938, 0, 0, 0}},
	{id: "92", typ: "ARMCTRL-level", device: "94 Edge mmc1", cpus: []int64{5669, 0, 0, 0}},
	{id: "IPI0", typ: "CPU wakeup interrupts", cpus: []int64{0, 0, 0, 0}},
	{id: "IPI1", typ: "Timer broadcast interrupts", cpus: []int64{0, 0, 0, 0}},
	{id: "IPI2", typ: "Rescheduling interrupts", cpus: []int64{23564958, 23464876, 23531165, 23040826}},
	{id: "IPI3", typ: "Function call interrupts", cpus: []int64{148438, 639704, 644266, 588150}},
	{id: "IPI4", typ: "CPU stop interrupts", cpus: []int64{0, 0, 0, 0}},
	{id: "IPI5", typ: "IRQ work interrupts", cpus: []int64{4348149, 1843985, 3819457, 1822877}},
	{id: "IPI6", typ: "completion interrupts", cpus: []int64{0, 0, 0, 0}},
}

func TestCpuAsTagsHwIrqs(t *testing.T) {
	acc, irqs := setup(t, hwIrqsString, true)
	reportMetrics("interrupts", irqs, acc, true)

	for _, irq := range hwIrqsExpectedArgs {
		expectCPUAsTags(acc, t, "interrupts", irq)
	}
}

func TestCpuAsFieldsHwIrqs(t *testing.T) {
	acc, irqs := setup(t, hwIrqsString, false)
	reportMetrics("interrupts", irqs, acc, false)

	for _, irq := range hwIrqsExpectedArgs {
		expectCPUAsFields(acc, t, "interrupts", irq)
	}
}
