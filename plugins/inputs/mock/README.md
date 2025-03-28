# Mock Data Input Plugin

The plugin generates mock-metrics based on different algorithms like sine-wave
functions, random numbers and more with the configured names and tags. Those
metrics are usefull during testing (e.g. processors) or if random data is
required.

⭐ Telegraf v1.22.0
🏷️ testing
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Generate metrics for test and demonstration purposes
[[inputs.mock]]
  ## Set the metric name to use for reporting
  metric_name = "mock"

  ## Optional string key-value pairs of tags to add to all metrics
  # [inputs.mock.tags]
  # "key" = "value"

  ## One or more mock data fields *must* be defined.
  # [[inputs.mock.constant]]
  #   name = "constant"
  #   value = value_of_any_type
  # [[inputs.mock.random]]
  #   name = "rand"
  #   min = 1.0
  #   max = 6.0
  # [[inputs.mock.sine_wave]]
  #   name = "wave"
  #   amplitude = 1.0
  #   period = 0.5
  #   phase = 20.0
  #   base_line = 0.0
  # [[inputs.mock.step]]
  #   name = "plus_one"
  #   start = 0.0
  #   step = 1.0
  # [[inputs.mock.stock]]
  #   name = "abc"
  #   price = 50.00
  #   volatility = 0.2
```

The mock plugin only requires that:

1) Metric name is set
2) One of the data field algorithms is defined

## Available Algorithms

The available algorithms for generating mock data include:

* `constant`: generate a field with the given value of type string, float, int
  or bool
* `random`: generate a random float, inclusive of min and max
* `sine_wave`: produce a sine wave with a certain amplitude, period and baseline
* `step`: always add the step value, negative values accepted
* `stock`: generate fake, stock-like price values based on a volatility variable

## Metrics

Metrics are entirely based on the user's own configuration and settings.

## Example Output

The following example shows all available algorithms configured with an
additional two tags as well:

```text
mock_sensors,building=5A,site=FTC random=4.875966794516125,abc=50,wave=0,plus_one=0 1632170840000000000
mock_sensors,building=5A,site=FTC random=5.738651873834452,abc=45.095549448434774,wave=5.877852522924732,plus_one=1 1632170850000000000
mock_sensors,building=5A,site=FTC random=1.0429328917205203,abc=51.928560083072924,wave=9.510565162951535,plus_one=2 1632170860000000000
mock_sensors,building=5A,site=FTC random=5.290188595384418,abc=44.41090520217027,wave=9.510565162951536,plus_one=3 1632170870000000000
mock_sensors,building=5A,site=FTC random=2.0724967227069135,abc=47.212167806890314,wave=5.877852522924733,plus_one=4 1632170880000000000
```
