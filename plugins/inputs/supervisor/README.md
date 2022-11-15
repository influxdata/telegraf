# Supervisor Input Plugin

This plugin gathers information about processes that
running under supervisor using XML-RPC API.

Minimum tested version of supervisor: 3.3.2

## Supervisor configuration

This plugin needs an HTTP server to be enabled in supervisor,
also it's recommended to enable basic authentication on the
HTTP server. When using basic authentication make sure to
include the username and password in the plugin's url setting.
Here is an example of the `inet_http_server` section in supervisor's
config that will work with default plugin configuration:

```ini
[inet_http_server]
port = 127.0.0.1:9001
username = user
password = pass
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Gathers information about processes that running under supervisor using XML-RPC API
[[inputs.supervisor]]
  ## Url of supervisor's XML-RPC endpoint if basic auth enabled in supervisor http server,
  ## than you have to add credentials to url (ex. http://login:pass@localhost:9001/RPC2)
  # url="http://localhost:9001/RPC2"
  ## With settings below you can manage gathering additional information about processes
  ## If both of them empty, then all additional information will be collected.
  ## Currently supported supported additional metrics are: pid, rc
  # metrics_include = []
  # metrics_exclude = ["pid", "rc"]
```

### Optional metrics

You can control gathering of some supervisor's metrics (processes PIDs
and exit codes) by setting metrics_include and metrics_exclude parameters
in configuration file.

### Server tag

Server tag is used to identify metrics source server. You have an option
to use host:port pair of supervisor's http endpoint by default or you
can use supervisor's identification string, which is set in supervisor's
configuration file.

## Metrics

- supervisor_processes
  - Tags:
    - source (Hostname or IP address of supervisor's instance)
    - port (Port number of supervisor's HTTP server)
    - id (Supervisor's identification string)
    - name (Process name)
    - group (Process group)
  - Fields:
    - state (int, see reference)
    - uptime (int, seconds)
    - pid (int, optional)
    - exitCode (int, optional)

- supervisor_instance
  - Tags:
    - source (Hostname or IP address of supervisor's instance)
    - port (Port number of supervisor's HTTP server)
    - id (Supervisor's identification string)
  - Fields:
    - state (int, see reference)

### Supervisor process state field reference table

|Statecode|Statename|                                            Description                                                 |
|--------|----------|--------------------------------------------------------------------------------------------------------|
|    0   |  STOPPED |             The process has been stopped due to a stop request or has never been started.              |
|   10   | STARTING |                             The process is starting due to a start request.                            |
|   20   |  RUNNING |                                       The process is running.                                          |
|   30   |  BACKOFF |The process entered the STARTING state but subsequently exited too quickly to move to the RUNNING state.|
|   40   | STOPPING |                           The process is stopping due to a stop request.                               |
|   100  |  EXITED  |                 The process exited from the RUNNING state (expectedly or unexpectedly).                |
|   200  |   FATAL  |                            The process could not be started successfully.                              |
|  1000  |  UNKNOWN |                  The process is in an unknown state (supervisord programming error).                   |

### Supervisor instance state field reference

|Statecode| Statename  |                  Description                 |
|---------|------------|----------------------------------------------|
|    2    |    FATAL   |  Supervisor has experienced a serious error. |
|    1    |   RUNNING  |         Supervisor is working normally.      |
|    0    | RESTARTING |  Supervisor is in the process of restarting. |
|   -1    |  SHUTDOWN  |Supervisor is in the process of shutting down.|

## Example Output

```shell
supervisor_processes,group=ExampleGroup,id=supervisor,port=9001,process=ExampleProcess,source=localhost state=20i,uptime=75958i 1659786637000000000
supervisor_instance,id=supervisor,port=9001,source=localhost state=1i 1659786637000000000
```
