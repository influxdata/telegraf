# Irqstat Input Plugin

The irqstat plugin gathers metrics about the interrupt types and associated values from `/proc/interrupts` and `/proc/softirqs` for each CPU present on a system.

### Configuration
```
[[inputs.irqstat]]
  include = ["0", "1", "30", "NET_RX"]
```

The above configuration would result in an output similar to:
```
./telegraf -config ~/irqstat_config.conf -test
* Plugin: inputs.irqstat, Collection 1
> interrupts,irq=30,type=PCI-MSI,device=65537-edge\ virtio1-input.0,host=hostname CPU0=1i,total=1i 1489346531000000000
> interrupts,irq=1,host=hostname,type=IO-APIC,device=1-edge\ i8042 CPU0=9i,total=9i 1489346531000000000
> soft_interrupts,irq=NET_RX,host=hostname CPU0=280879i,total=280879i 1489346531000000000
> interrupts,irq=0,type=IO-APIC,device=2-edge\ timer,host=hostname CPU0=23i,total=23i 1489346531000000000
```

# Measurements

There are two measurements reported by this plugin.
- `interrupts` reports metrics from the `/proc/interrupts` file
- `soft_interrupts` reports metrics from the `/proc/softirqs` file

Depending on the content of each file there will multiple tags and fields for each measurement
- <strong>Fields:</strong>
 - CPUx: the IRQ value based on CPU number
 - Total: total IRQ value of all CPUs
- <strong>Tags:</strong>
 - IRQ: the IRQ
 - Type: the type associated with the IRQ
 - Device: the device associated with the IRQ
