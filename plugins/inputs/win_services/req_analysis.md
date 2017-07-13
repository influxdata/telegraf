#  Win_services telegraf input plugin analysis
This is analysis of a telegraf feature requirement demanding plugin for collecting windows services info,
originally requested in [telegraf issue #2714](https://github.com/influxdata/telegraf/issues/2714)

## Feature
### Use Cases
- Admin needs to monitor state of selected windows services on the host, along with a few additional properties (display name, start-up mode)
- Admin wants to query info of monitored services
- Admin wants to query what services have defined properties

### Feature requirements
 * telegraf input plugin
 * store service name, display name, state and startup mode
 * configure what services should be monitored
 
### Platform requirements
 * WindowsXP and higher 
 * Windows Server 2003 and higher
 
 Defined by Windows Service API, used for querying services, availability

### Admin rights
Admin privileges are required to read service info. However, as telegraf mostly run as a service running under local system account, it should be no problem.

### Deployment
Feature request mentions monitoring of 5000 servers. This either means:
* deploying telegraf on each monitored host, what is the preferred option, as the other plugins can be used to monitor other stuff on the host,
* plugin has to monitor multiple servers, what would lead to more complex plugin (service input plugin) along with complex configuration

## Implementation
### Storing service info
#### Measurement
There are two options to define what measurement would be
1. Store all service info in single measurement, e.g. win_services, configurable
2. Store services info per service

Option 1. has the biggest benefit that user can easily query info about all services, e.g. all services in stopped state, but this leads to a lot of data in single measurement. As services measurement has the same schema ti make sense to use this.

Option 2. diversifies the data but makes it difficult to query multiple services

_Q: What is the best practise? is would prefer option 1.. Also, it could be configurable._

#### Properties

Basic properties to monitor and store, as initially requested, are:
 * service name
 * display name
 * startup mode 
 * actual state
 
 They are also additional properties available(see [SERVICE_STATUS_PROCESS](https://msdn.microsoft.com/en-us/library/windows/desktop/ms685992(v=vs.85).aspx) and [QUERY_SERVICE_CONFIG](https://msdn.microsoft.com/en-us/library/windows/desktop/ms684950(v=vs.85).aspx) ), but those are quite advanced and not so suitable for monitoring.
 
 Display name and startup mode, are rather static properties and they will change rarely, but as most probable output is influxdb it will handle this with compaction.
 
According to use cases, condition in queries can be based on _service name_ or a property (state, startup mode).
Services will won't be most probably filtered according to display name. 

So mapping to tag or field would be:

* **Field** = displayName
* **Tag** = service name, startup mode, actual state

 Let's use following key name and 'types':
 
 Property|Key | Type
  ---- |----- | ---
 service name| service_name | string
 display name| display_name | string
 startup mode| startup_mode | string
 actual state| state | string

Keys _startupMode_ and _state_ will be  a human readable representation of the attribute.

 ### Configuration
 * User must be able to set what services to be monitored:
  ````
  # Case-insensitive name of services to monitor. Empty of all services
  service_names = [
    "LanmanServer",
    "TermService"
  ]
 ````
Services in example should be available on all Windows edition and versions. 
 
 ### Storing Errors
  There are basically two possible errors:
  * Invalid service name given in configuration
  * A service require special privileges
  
  This should be reported as a warning, not complete error of the Gather function
    
   As stated by Daniel, best practice is to use Accumulator.AddError and log it every time. Telegraf should handle multiple same errors.
      
 ### Caching
  Most service info is almost static and it could be cached. But as all the info about requested services Windows Service Manager stores in memory, even full listing takes just 8ms on Windows 10 on Core i5 (2 cores) laptop,
  so caching seems overhead.