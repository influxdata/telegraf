# Supervisor Input Plugin

This plugin gather information about processes that running under supervisor using XML-RPC API

Plugin minimum tested version: 1.17.0

### Supervisor configuration

This plugin needs TCP HTTP server to be enabled without basic auth for collecting information. Here is example of
`inet_http_server` section in supervisor config that will work with default plugin configuration.

```
[inet_http_server]
port=127.0.0.1:9001
```



### Plugin configuration

```toml
[inputs.supervisor]
  ## Url of supervisor's XML-RPC endpoint
  # url="http://localhost:9001/RPC2"
  ## Use supervisor identification string as server tag
  use_identification_tag = false
  ## Gather PID of running processes
  gather_pid = false
  ## Gather exit codes of processes
  gather_exit_code = false
```

#### Optional metrics

By default this plugin doesn't collect any information about processes pids and exit codes, you can enable it by setting
`gather_pid` and `gather_exit_code` options in configuration file. 

#### Server tag
Server tag is used to identify metrics source server. You have an option to use host:port pair of supervisor's http
endpoint by default or you can use supervisor's identification string, which is set in supervisor's configuration file. 

### Metrics

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
    
+ supervisor_instance
    - Tags:
        - server (Supervisor address or identification string)
    - Fields:
        - state (int, see reference) 

#### Supervisor process state field reference table

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

#### Supervisor instance state field reference

|Statecode| Statename  |                  Description                 |
|---------|------------|----------------------------------------------|
|    2    |    FATAL   |  Supervisor has experienced a serious error. |
|    1    |   RUNNING  |        Supervisor is working normally.       |
|    0    | RESTARTING |  Supervisor is in the process of restarting. |
|   -1    |  SHUTDOWN  |Supervisor is in the process of shutting down.|
### Example Output

```
supervisor_processes,host=localhost,group=ExampleGroup,process=ExampleProcess,server=localhost:9001 exitCode=0i,pid=12345i,state=20i,uptime=4812i
supervisor_instance,host=localhost,server=localhost:9001 state=1
```
