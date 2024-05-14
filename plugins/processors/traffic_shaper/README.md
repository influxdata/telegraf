# Traffic Shaper Processor

An in-memory traffic shaper processor which evens out traffic so that output traffic is uniform

Example of uneven traffic distribution ![traffic_distribution](./docs/traffic_distribution.png)
After applying traffic shaper the output traffic distribution is uniform

## Configuration

```toml @sample.conf
# Traffic Shaper shapes the traffic from non-uniform distribution to uniform distribution
[[processors.traffic_shaper]]

  ## No of samples to be emitted per time unit, default is seconds
  ## This should be used in conjunction with number of telegraf instances.
  samples = 20000

  ## Buffer Size
  ## If buffer is full the incoming metrics will be dropped
  buffer_size = 1000000
```
