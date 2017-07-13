# Telegraf Plugin: win_services
Input plugin to report Windows services info: service name, display name, state, startup mode

### Configuration:

```toml
[[inputs.win_services]]
  ## Name of services to monitor. Set empty to monitor all the available services on the host
  service_names = [
    "LanmanServer",
    "TermService",
  ]
```

### Measurements & Fields:

- win_services
  - display_name

### Tags:

- All measurements have the following tags:
    - service_name
    - state
    - startup_mode

The `state` tag can have following values:
* _service_stopped_         
* _service_start_pending_   
* _service_stop_pending_    
* _service_running_         
* _service_continue_pending_
* _service_pause_pending_   
* _service_paused_
          
The `startup_mode` tag can have following values:
* _service_boot_start_  
* _service_system_start_
* _service_auto_start_  
* _service_demand_start_
* _service_disabled_    

### Example Output:

Using default configuration:

When run with:
```
E:\Telegraf>telegraf.exe -config telegraf.conf -test
```
It produces:
```
* Plugin: inputs.win_services, Collection 1
> win_services,state=service_running,startup_mode=service_auto_start,host=WIN2008R2H401,service_name=LanmanServer display_name="Server" 1499947615000000000
> win_services,service_name=TermService,state=service_stopped,startup_mode=service_demand_start,host=WIN2008R2H401 display_name="Remote Desktop Services" 1499947615000000000
```
