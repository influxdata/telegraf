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
- CPUx: the amount of interrupts for the IRQ handled by that CPU
- total: total amount of interrupts for all CPUs

### Tags
- irq: the IRQ
- type: the type of interrupt
- device: the name of the device that is located at that IRQ

### Example Output
```
./telegraf --config ~/interrupts_config.conf --test
* Plugin: inputs.interrupts, Collection 1
> interrupts,irq=0,type=IO-APIC,device=2-edge\ timer,host=hostname CPU0=23i,total=23i 1489346531000000000
> interrupts,irq=1,host=hostname,type=IO-APIC,device=1-edge\ i8042 CPU0=9i,total=9i 1489346531000000000
> interrupts,irq=30,type=PCI-MSI,device=65537-edge\ virtio1-input.0,host=hostname CPU0=1i,total=1i 1489346531000000000
> soft_interrupts,irq=NET_RX,host=hostname CPU0=280879i,total=280879i 1489346531000000000
```
