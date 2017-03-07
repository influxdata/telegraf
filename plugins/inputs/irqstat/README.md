# Irqstat Input Plugin

The irqstat plugin gathers metrics about the interrupt types and associated values for each CPU present on a system.

### Configuration
```
[[inputs.irqstat]]
  include = ["0", "1"]
```

The above configuration would result in an output similar to:
```
./telegraf -config ~/irqstat_config.conf -test
* Plugin: inputs.irqstat, Collection 1
> irqstat,host=hostname,cpu=CPU0 1=9i,0=22i 1488751337000000000
```

# Measurements

There is only one measurement reported by this plugin, `irqstat`:
- <strong>Fields:</strong> IRQs
- <strong>Tags:</strong> CPUs
