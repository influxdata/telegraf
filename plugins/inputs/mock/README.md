# Mock Data

The mock input plugin generates random data based on a selection of different
algorithms. For example, it can produce random data between a set of values,
fake stock data, sine waves, and step-wise values.

Additionally, users can set the measurement name and tags used to whatever is
required to mock their situation.

## Configuration

The mock plugin only requires that:

1) Metric name is set
2) One of the below data field algorithms is defined

Below is a sample config to generate one of each of the four types:

```toml
# Generate metrics for test and demonstration purposes
[[inputs.mock]]
  ## Set the metric name to use for reporting
  metric_name = "mock"

  ## Optional string key-value pairs of tags to add to all metrics
  # [inputs.mock.tags]
  # "key" = "value"

  ## One or more mock data fields *must* be defined.
  ##
  ## [[inputs.mock.random]]
  ##   name = "rand"
  ##   min = 1.0
  ##   max = 6.0
  ## [[inputs.mock.sine_wave]]
  ##   name = "wave"
  ##   amplitude = 1.0
  ##   period = 0.5
  ## [[inputs.mock.step]]
  ##   name = "plus_one"
  ##   start = 0.0
  ##   step = 1.0
  ## [[inputs.mock.stock]]
  ##   name = "abc"
  ##   price = 50.00
  ##   volatility = 0.2
```

## Available Algorithms

The available algorithms for generating mock data include:

* Random Float - generate a random float, inclusive of min and max
* Sine Wave - produce a sine wave with a certain amplitude and period
* Step - always add the step value, negative values accepted
* Stock - generate fake, stock-like price values based on a volatility variable

## Example Output

The following example shows all available algorithms configured with an
additional two tags as well:

```s
mock_sensors,building=5A,site=FTC random=4.875966794516125,abc=50,wave=0,plus_one=0 1632170840000000000
mock_sensors,building=5A,site=FTC random=5.738651873834452,abc=45.095549448434774,wave=5.877852522924732,plus_one=1 1632170850000000000
mock_sensors,building=5A,site=FTC random=1.0429328917205203,abc=51.928560083072924,wave=9.510565162951535,plus_one=2 1632170860000000000
mock_sensors,building=5A,site=FTC random=5.290188595384418,abc=44.41090520217027,wave=9.510565162951536,plus_one=3 1632170870000000000
mock_sensors,building=5A,site=FTC random=2.0724967227069135,abc=47.212167806890314,wave=5.877852522924733,plus_one=4 1632170880000000000
```
