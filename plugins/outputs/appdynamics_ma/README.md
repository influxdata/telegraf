# AppDynamics Machine Agent (MA) Output Plugin

This output plugin writes to an [AppDynamics Machine Agent](https://docs.appdynamics.com/display/PRO45/Standalone+Machine+Agent), specifically to the agent's [HTTP Listener](https://docs.appdynamics.com/display/PRO45/Standalone+Machine+Agent+HTTP+Listener).

### Configuration:

```toml
  ## AppDynamics Machine Agent Host (Required)
  host = "http://127.0.0.1"

  ## AppDynamics Machine Agent HTTP Listener Port (Required)
  port = "8293"

  ## AppDynamics Metric Path (Required)
  metricPath = "Custom Metrics|Telegraf|"
```

### Host:

The default `host` value `"http://127.0.0.1"` assumes a locally running Machine Agent.  
However, if you desire, you may configure this value to the IP address or resolvable hostname where the agent is running.
Please leave the `http://` prefix, and wrap the entire `host` value in double quotes.

### Port:

The default `port` value `"8293"` is the default port used by the Machine Agent HTTP Listener.
If you override the MA HTTP Port, make the change accordingly here.
Please wrap the `port` value in double quotes.

### Metric Path:

The default `metricPath` value `"Custom Metrics|Telegraf|"` assumes you are sending metrics to a Machine agent using [Server Visibility](https://docs.appdynamics.com/display/PRO45/Server+Visibility).
If you are not using Server Visibility, or need further explanation of how to configure a metric path, please consult this [Knowledge Base Article](https://community.appdynamics.com/t5/Knowledge-Base/How-do-I-troubleshoot-missing-custom-metrics-or-extensions/ta-p/28695).
Please wrap the entire `metricPath` value in double quotes.
