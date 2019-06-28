# Fireboard Input Plugin

The fireboard plugin gathers the real time temperature data from fireboard
thermometers.  In order to use this input plugin, you'll need to sign up
to use their REST API, you can find more information on their website
here [https://docs.fireboard.io/reference/restapi.html]

### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
[[inputs.fireboard]]
  ## Specify auth token for your account
  ## https://docs.fireboard.io/reference/restapi.html#Authentication
  # authToken = "b4bb6e6a7b6231acb9f71b304edb2274693d8849"
  ## You can override the fireboard server URL if necessary
  # URL = https://fireboard.io/api/v1/devices.json
```

#### authToken

In lieu of requiring a username and password, this plugin requires the
authToken that you can generate using the Fireboard REST API as described
in their docs [https://docs.fireboard.io/reference/restapi.html#Authentication]

#### URL

While there should be no reason to override the URL, the option is available
in case Fireboard changes their site, etc.

### Metrics

The Fireboard REST API docs have good examples of the data that is available,
currently this input only returns the real time temperatures. Temperature 
values are included if they are less than a minute old.

- fireboard
  - tags:
    - channel
    - scale (1=celcius; 2=farenheit)
    - title (name of the Firebaord)
    - uuid (UUID of the Firebaord)
  - fields:
    - temperature (float, unit)

### Example Output

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```
fireboard,channel=2,host=patas-mbp,scale=2,title=telegraf-FireBoard,uuid=b55e766c-b308-49b5-93a4-df89fe31efd0 temperature=78.2 1561690040000000000
```