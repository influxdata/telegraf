# Scheduling monitor Input Plugin
The `sched_monitor` plugin gathers scheduling metrics such as cpu time, and voluntary/involuntary context switches for configured CPUs.
It is recommended to use for tracking performance issues driven by scheduling pressure or general compute resource distribution efficiency.
Using it to collect data from a noisy, general purpose CPU may create large volumes of data. This plugin is ideal for tracing scheduling dynamics
on an isolated CPU.
As this plugin's functionality relies on the 'proc' subsystem, it only works on Linux

#### Configuration
```toml
[[inputs.sched_monitor]]
  ## CPUs to monitor
  cpu_list = 
  ## Filter operating system threads
  exclude_kernel = false
```

### Metrics
sched_monitor,cmd=irq/68-nvidia,cpu=0,host=qdlt-pc cpu_time=773715i,ctx_swtch=8i,invl_ctx_swtch=0i 1571432791000000000
sched_monitor,cmd=kworker/0:1,cpu=0,host=qdlt-pc cpu_time=36815i,ctx_swtch=6i,invl_ctx_swtch=0i 1571432791000000000
sched_monitor,cmd=jbd2/nvme0n1p2,cpu=0,host=qdlt-pc cpu_time=84029i,ctx_swtch=3i,invl_ctx_swtch=1i 1571432791000000000
sched_monitor,cmd=kworker/0:1H,cpu=0,host=qdlt-pc cpu_time=12548i,ctx_swtch=2i,invl_ctx_swtch=0i 1571432791000000000
sched_monitor,cmd=Chrome_IOThread,cpu=0,host=qdlt-pc cpu_time=92369i,ctx_swtch=1i,invl_ctx_swtch=0i 1571432791000000000

