# Interrupts Input Plugin

The interrupts plugin gathers metrics about IRQs from `/proc/interrupts` and `/proc/softirqs`.

### Configuration
```
[[inputs.interrupts]]
  ## To filter which IRQs to collect, make use of tagpass / tagdrop, i.e.
  # [inputs.interrupts.tagdrop]
    # irq = [ "NET_RX", "TASKLET" ]
```

### Measurements
There are two measurements reported by this plugin.
- `interrupts` gathers metrics from the `/proc/interrupts` file
- `soft_interrupts` gathers metrics from the `/proc/softirqs` file

### Fields
- cpu: the amount of interrupts for the IRQ handled by the CPU in the cpu tag, or the total amount of interrupts for all CPUs (cpu-total)

### Tags
- cpu: the cpu associated with the interrupt, or "cpu-total" representing all CPU's
- irq: the IRQ
- type: the type of interrupt
- device: the name of the device that is located at that IRQ

### Example Output
```
./telegraf --config ~/interrupts_config.conf --test
* Plugin: inputs.interrupts, Collection 1
* Plugin: inputs.interrupts, Collection 1
> soft_interrupts,irq=HI,cpu=cpu0,host=etl cpu=0i 1502206961000000000
> soft_interrupts,irq=HI,cpu=cpu1,host=etl cpu=0i 1502206961000000000
> soft_interrupts,host=etl,cpu=cpu-total,irq=HI cpu=0i 1502206961000000000
> soft_interrupts,cpu=cpu0,host=etl,irq=TIMER cpu=3160625i 1502206961000000000
> soft_interrupts,irq=TIMER,cpu=cpu1,host=etl cpu=1863935i 1502206961000000000
> soft_interrupts,irq=TIMER,host=etl,cpu=cpu-total cpu=5024560i 1502206961000000000
> interrupts,irq=0,type=IO-APIC-edge,device=timer,cpu=cpu0,host=etl cpu=121i 1502206961000000000
> interrupts,cpu=cpu1,host=etl,irq=0,type=IO-APIC-edge,device=timer cpu=0i 1502206961000000000
> interrupts,type=IO-APIC-edge,device=timer,host=etl,cpu=cpu-total,irq=0 cpu=121i 1502206961000000000
> interrupts,irq=1,type=IO-APIC-edge,device=i8042,cpu=cpu0,host=etl cpu=10i 1502206961000000000
> interrupts,irq=1,type=IO-APIC-edge,device=i8042,cpu=cpu1,host=etl cpu=0i 1502206961000000000
> interrupts,cpu=cpu-total,irq=1,type=IO-APIC-edge,device=i8042,host=etl cpu=10i 1502206961000000000
```
