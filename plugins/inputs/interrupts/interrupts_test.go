package interrupts

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// =====================================================================================
//	Setup and helper functions
// =====================================================================================

func expectCPUAsTags(m *testutil.Accumulator, t *testing.T, measurement string, irq IRQ) {
	for idx, value := range irq.Cpus {
		m.AssertContainsTaggedFields(t, measurement, map[string]interface{}{"count": value}, map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device, "cpu": fmt.Sprintf("cpu%d", idx)})
	}
}

func expectCPUAsFields(m *testutil.Accumulator, t *testing.T, measurement string, irq IRQ) {
	fields := map[string]interface{}{}
	total := int64(0)
	for idx, count := range irq.Cpus {
		fields[fmt.Sprintf("CPU%d", idx)] = count
		total += count
	}
	fields["total"] = total

	m.AssertContainsTaggedFields(t, measurement, fields, map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device})
}

func setup(t *testing.T, irqString string, cpuAsTags bool) (*testutil.Accumulator, []IRQ) {
	f := bytes.NewBufferString(irqString)
	irqs, err := parseInterrupts(f)
	require.Equal(t, nil, err)
	require.NotEqual(t, 0, len(irqs))

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

var softIrqsExpectedArgs = []IRQ{
	{ID: "0", Type: "IO-APIC-edge", Device: "timer", Cpus: []int64{134, 0}},
	{ID: "1", Type: "IO-APIC-edge", Device: "i8042", Cpus: []int64{7, 3}},
	{ID: "NMI", Type: "Non-maskable interrupts", Cpus: []int64{0, 0}},
	{ID: "MIS", Cpus: []int64{0}},
	{ID: "NET_RX", Cpus: []int64{867028, 225}},
	{ID: "TASKLET", Cpus: []int64{205, 0}},
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

var hwIrqsExpectedArgs = []IRQ{
	{ID: "16", Type: "bcm2836-timer", Device: "0 Edge arch_timer", Cpus: []int64{0, 0, 0, 0}},
	{ID: "17", Type: "bcm2836-timer", Device: "1 Edge arch_timer", Cpus: []int64{127224250, 118424219, 127224437, 117885416}},
	{ID: "21", Type: "bcm2836-pmu", Device: "9 Edge arm-pmu", Cpus: []int64{0, 0, 0, 0}},
	{ID: "23", Type: "ARMCTRL-level", Device: "1 Edge 3f00b880.mailbox", Cpus: []int64{1549514, 0, 0, 0}},
	{ID: "24", Type: "ARMCTRL-level", Device: "2 Edge VCHIQ doorbell", Cpus: []int64{2, 0, 0, 0}},
	{ID: "46", Type: "ARMCTRL-level", Device: "48 Edge bcm2708_fb dma", Cpus: []int64{0, 0, 0, 0}},
	{ID: "48", Type: "ARMCTRL-level", Device: "50 Edge DMA IRQ", Cpus: []int64{0, 0, 0, 0}},
	{ID: "50", Type: "ARMCTRL-level", Device: "52 Edge DMA IRQ", Cpus: []int64{0, 0, 0, 0}},
	{ID: "51", Type: "ARMCTRL-level", Device: "53 Edge DMA IRQ", Cpus: []int64{208, 0, 0, 0}},
	{ID: "54", Type: "ARMCTRL-level", Device: "56 Edge DMA IRQ", Cpus: []int64{883002, 0, 0, 0}},
	{ID: "59", Type: "ARMCTRL-level", Device: "61 Edge bcm2835-auxirq", Cpus: []int64{0, 0, 0, 0}},
	{ID: "62", Type: "ARMCTRL-level", Device: "64 Edge dwc_otg, dwc_otg_pcd, dwc_otg_hcd:usb1", Cpus: []int64{521451447, 0, 0, 0}},
	{ID: "86", Type: "ARMCTRL-level", Device: "88 Edge mmc0", Cpus: []int64{857597, 0, 0, 0}},
	{ID: "87", Type: "ARMCTRL-level", Device: "89 Edge uart-pl011", Cpus: []int64{4938, 0, 0, 0}},
	{ID: "92", Type: "ARMCTRL-level", Device: "94 Edge mmc1", Cpus: []int64{5669, 0, 0, 0}},
	{ID: "IPI0", Type: "CPU wakeup interrupts", Cpus: []int64{0, 0, 0, 0}},
	{ID: "IPI1", Type: "Timer broadcast interrupts", Cpus: []int64{0, 0, 0, 0}},
	{ID: "IPI2", Type: "Rescheduling interrupts", Cpus: []int64{23564958, 23464876, 23531165, 23040826}},
	{ID: "IPI3", Type: "Function call interrupts", Cpus: []int64{148438, 639704, 644266, 588150}},
	{ID: "IPI4", Type: "CPU stop interrupts", Cpus: []int64{0, 0, 0, 0}},
	{ID: "IPI5", Type: "IRQ work interrupts", Cpus: []int64{4348149, 1843985, 3819457, 1822877}},
	{ID: "IPI6", Type: "completion interrupts", Cpus: []int64{0, 0, 0, 0}},
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
