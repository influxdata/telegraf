# Noise Processor

The *Noise* processor is used to add noise to numerical field values. For each field a noise is generated using a defined probability densitiy function and added to the value. The function type can be configured as _Laplace_, _Gaussian_ or _Uniform_.
Depending on the function, various parameters need to be configured:

### Laplace:
* `noise_type = laplace`
* `scale`: also referred to as _diversity_ parameter, regulates the width & height of the function, a bigger `scale` value means a higher probability of larger noise, default set to 1.0
* `mu`: location of the curve, default set to 0.0

### Gaussian:
* `noise_type = gaussian`
* `mu`: mean value, default set to 0.0
* `sigma`: standard deviation, default set to 0.1

### Uniform
* `noise_type = uniform`
* `min`: minimal interval value, default set to -1.0
* `max`: maximal interval value, default set to 1.0

### Configuration
Depending on the choice of the distribution function, the respective parameters must be set. Default settings are `noise_type = "laplace"` with `mu = 0.0` and `scale = 1.0`:

```toml
[[processors.noise]]
  scale = 1.0
  min = -1.0
  max = 1.0
  mu = 0.0
  sigma = 0.1
  noise_type = "laplace"
  noise_log = true
  include_fields = []
  exclude_fields = []
```

If the generated noise values should be saved as a new field value, the option `noise_log` can be enabled.

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
  mu = 0.0
  noise_type = "laplace"
  noise_log = true
  include_fields = []
  exclude_fields = ["usage_steal", "usage_user", "uptime_format", "usage_idle" ]
  namedrop = ["swap", "disk", "net"]
```

Result of noise added to the _cpu_ metric:

```diff
- cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:0 usage_guest_nice:0 usage_idle:94.3999999994412 usage_iowait:0 usage_irq:0.1999999999998181 usage_nice:0 usage_softirq:0.20000000000209184 usage_steal:0 usage_system:1.2000000000080036 usage_user:4.000000000014552]
+ cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:1.0078071583066057 usage_guest_nice:0.523063861602435 usage_idle:95.53920223476884 usage_iowait:0.5162661526251292 usage_irq:0.7138529816101375 usage_nice:0.6119678488887954 usage_softirq:0.5573585443688622 usage_steal:0.2006120911289802 usage_system:1.2954475820198437 usage_user:6.885664792615023]
```