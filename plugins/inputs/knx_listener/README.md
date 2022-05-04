# KNX Input Plugin

The KNX input plugin that listens for messages on the KNX home-automation bus.
This plugin connects to the KNX bus via a KNX-IP interface.
Information about supported KNX message datapoint types can be found at the
underlying "knx-go" project site (<https://github.com/vapourismo/knx-go>).

## Configuration

This is a sample config for the plugin.

```toml
# Listener capable of handling KNX bus messages provided through a KNX-IP Interface.
[[inputs.knx_listener]]
  ## Type of KNX-IP interface.
  ## Can be either "tunnel" or "router".
  # service_type = "tunnel"

  ## Address of the KNX-IP interface.
  service_address = "localhost:3671"

  ## Measurement definition(s)
  # [[inputs.knx_listener.measurement]]
  #   ## Name of the measurement
  #   name = "temperature"
  #   ## Datapoint-Type (DPT) of the KNX messages
  #   dpt = "9.001"
  #   ## List of Group-Addresses (GAs) assigned to the measurement
  #   addresses = ["5/5/1"]

  # [[inputs.knx_listener.measurement]]
  #   name = "illumination"
  #   dpt = "9.004"
  #   addresses = ["5/5/3"]
```

### Measurement configurations

Each measurement contains only one datapoint-type (DPT) and assigns a list of
addresses to this measurement. You can, for example group all temperature sensor
messages within a "temperature" measurement. However, you are free to split
messages of one datapoint-type to multiple measurements.

**NOTE: You should not assign a group-address (GA) to multiple measurements!**

## Metrics

Received KNX data is stored in the named measurement as configured above using
the "value" field. Additional to the value, there are the following tags added
to the datapoint:

- "groupaddress": KNX group-address corresponding to the value
- "unit":         unit of the value
- "source":       KNX physical address sending the value

To find out about the datatype of the datapoint please check your KNX project,
the KNX-specification or the "knx-go" project for the corresponding DPT.

## Example Output

This section shows example output in Line Protocol format.

```shell
illumination,groupaddress=5/5/4,host=Hugin,source=1.1.12,unit=lux value=17.889999389648438 1582132674999013274
temperature,groupaddress=5/5/1,host=Hugin,source=1.1.8,unit=°C value=17.799999237060547 1582132663427587361
windowopen,groupaddress=1/0/1,host=Hugin,source=1.1.3 value=true 1582132630425581320
```
