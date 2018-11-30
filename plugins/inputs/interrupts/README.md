# Interrupts Input Plugin

The interrupts plugin gathers metrics about IRQs from `/proc/interrupts` and `/proc/softirqs`.

### Configuration
```
[[inputs.interrupts]]
  # To report cpus as tags instead of fields use cpu_as_tags
    # cpu_as_tags = false
  #
  ## To filter which IRQs to collect, make use of tagpass / tagdrop, i.e.
  # [inputs.interrupts.tagdrop]
    # irq = [ "NET_RX", "TASKLET" ]
```

### Measurements
There are two measurements reported by this plugin.
- `interrupts` gathers metrics from the `/proc/interrupts` file
- `soft_interrupts` gathers metrics from the `/proc/softirqs` file

### Fields
For cpu_as_tags=false (default):
- CPUx: the amount of interrupts for the IRQ handled by the CPU
- Total: sum of interrupts for the IRS for all CPUs
For cpu_as_tags=true ():
- Count: the amount of interrupts for the IRQ handled by CPU described in CPU tag

### Tags
- irq: the IRQ
- type: the type of interrupt
- device: the name of the device that is located at that IRQ
- cpu: the CPU (when cpus_as_tags=true)

### Example Output
```
./telegraf --config ~/interrupts_config.conf --test
For cpus_as_tags=false (default):
* Plugin: inputs.interrupts, Collection 1
> interrupts,irq=0,type=IO-APIC,device=2-edge\ timer,host=hostname,cpu=cpu0 count=23i 1489346531000000000
> interrupts,irq=1,host=hostname,type=IO-APIC,device=1-edge\ i8042,cpu=cpu0 count=9i 1489346531000000000
> interrupts,irq=30,type=PCI-MSI,device=65537-edge\ virtio1-input.0,host=hostname,cpu=cpu1 count=1i 1489346531000000000
> soft_interrupts,irq=NET_RX,host=hostname,cpu=cpu0 count=280879i 1489346531000000000

For cpus_as_tags=true:
> interrupts,cpu=cpu6,host=hostname,irq=PIW,type=Posted-interrupt\ wakeup\ event count=0i 1543539773000000000
> interrupts,cpu=cpu7,host=hostname,irq=PIW,type=Posted-interrupt\ wakeup\ event count=0i 1543539773000000000
> soft_interrupts,cpu=cpu0,host=hostname,irq=HI count=246441i 1543539773000000000
> soft_interrupts,cpu=cpu1,host=hostname,irq=HI count=159154i 1543539773000000000

```
