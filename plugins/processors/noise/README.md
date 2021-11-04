# Noise Processor

The *Noise* processor is used to add noise to numerical field values. For each field a noise is generated using Laplace distribution and added to the value.
Using the _scale_ parameter the effect of the noise can be modified. A smaller _scale_ value means that noise variation will be smaller and the resulting value will be more close to the real value.

### Configuration
```toml
[[processors.noise]]
  scale = 1.0
  include_fields = []
  exclude_fields = []
  namedrop = []
```

### Example
Add noise to each value the *Inputs.CPU*  plugin generates, except for the _usage\_steal_, _usage\_user_, _uptime\_format_, _usage\_idle_ field and all fields of the metrics _swap_, _disk_ and _net_:


```toml
[[inputs.cpu]]
  percpu = true
  totalcpu = true
  collect_cpu_time = false
  report_active = false

[[processors.noise]]
  scale = 1.0
  include_fields = []
  exclude_fields = ["usage_steal", "usage_user", "uptime_format", "usage_idle",
  namedrop = ["swap", "disk", "net"]
```

Result of noise added to the _cpu_ metric:

```diff
- cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:0 usage_guest_nice:0 usage_idle:94.3999999994412 usage_iowait:0 usage_irq:0.1999999999998181 usage_nice:0 usage_softirq:0.20000000000209184 usage_steal:0 usage_system:1.2000000000080036 usage_user:4.000000000014552]
+ cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:1.0078071583066057 usage_guest_nice:0.523063861602435 usage_idle:95.53920223476884 usage_iowait:0.5162661526251292 usage_irq:0.7138529816101375 usage_nice:0.6119678488887954 usage_softirq:0.5573585443688622 usage_steal:0.2006120911289802 usage_system:1.2954475820198437 usage_user:6.885664792615023]
```