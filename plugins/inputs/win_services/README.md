# Telegraf Plugin: win_services
Input plugin to report Windows services info.

It requires that Telegraf must be running under the administrator privileges.
### Configuration:

```toml
[[inputs.win_services]]
  ## Names of the services to monitor. Leave empty to monitor all the available services on the host
  service_names = [
    "LanmanServer",
    "TermService",
  ]
```

### Measurements & Fields:

- win_services
    - state
    - startup_mode

The `state` field can have the following values:
* _stopped_         
* _start_pending_   
* _stop_pending_    
* _running_         
* _continue_pending_
* _pause_pending_   
* _paused_

The `startup_mode` field can have the following values:
* _boot_start_  
* _system_start_
* _auto_start_  
* _demand_start_
* _disabled_

### Tags:

- All measurements have the following tags:
    - service_name
    - display_name

### Example Output:
```
* Plugin: inputs.win_services, Collection 1
> win_services,host=WIN2008R2H401,display_name=Server,service_name=LanmanServer state="running",startup_mode="auto_start" 1500040669000000000
> win_services,display_name=Remote\ Desktop\ Services,service_name=TermService,host=WIN2008R2H401 state="stopped",startup_mode="demand_start" 1500040669000000000
```
### TICK Scripts

A sample TICK script for a notification about a not running service.
It sends a notification whenever any service changes its state to be not _running_ and when it changes that state back to _running_. 
The notification is sent via an HTTP POST call.

```
stream
    |from()
        .database('telegraf')
        .retentionPolicy('autogen')
        .measurement('win_services')
        .groupBy('host','service_name')
    |alert()
        .crit(lambda: "state" != 'running')
        .stateChangesOnly()
        .message('Service {{ index .Tags "service_name" }} on Host {{ index .Tags "host" }} is {{ index .Fields "state" }} ')
        .post('http://localhost:666/alert/service')
```
