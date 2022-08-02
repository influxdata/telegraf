# Supervisor Input Plugin

This plugin gather information about processes that running under supervisor using XML-RPC API

Plugin minimum tested version: 3.3.2

## Supervisor configuration

This plugin needs TCP HTTP server to be enabled for collecting information. Here is example of
`inet_http_server` section in supervisor config that will work with default plugin configuration.

```ini
[inet_http_server]
port=127.0.0.1:9001
```

It is also recommended to setup basic authentication to http server as described [here](http://supervisord.org/configuration.html#inet-http-server-section-values).

## Configuration

```toml
[inputs.supervisor]
  ## Url of supervisor's XML-RPC endpoint if basic auth enabled in supervisor http server,
  ## than you have to add credentials to url (ex. http://login:pass@localhost:9001/RPC2)
  # url="http://localhost:9001/RPC2"
  ## Use supervisor identification string as server tag
  # use_identification_tag = false
  ## With settings below you can manage gathering additional information about processes
  ## If both of them empty, then all additional information will be collected.
  ## Currently supported supported additional metrics are: pid, rc
  # metrics_include = []
  # metrics_exclude = ["pid", "rc"]
```

### Optional metrics

You can control gathering of some supervisor's metrics (processes PIDs and exit codes) by setting metrics_include
and metrics_exclude parameters in configuration file.

### Server tag

Server tag is used to identify metrics source server. You have an option to use host:port pair of supervisor's http
endpoint by default or you can use supervisor's identification string, which is set in supervisor's configuration file.

## Metrics

- supervisor_processes
  - Tags:
    - server (Supervisor address or identification string)
    - name (Process name)
    - group (Process group)
  - Fields:
    - state (int, see reference)
    - uptime (int, seconds)
    - pid (int, optional)
    - exitCode (int, optional)

- supervisor_instance
  - Tags:
    - server (Supervisor address or identification string)
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
supervisor_processes,host=localhost,group=ExampleGroup,process=ExampleProcess,server=localhost:9001 exitCode=0i,pid=12345i,state=20i,uptime=4812i
supervisor_instance,host=localhost,server=localhost:9001 state=1
```
