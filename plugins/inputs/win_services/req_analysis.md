#  Win_services telegraf input plugin analysis
This is an analysis of a telegraf feature requirement demanding a plugin for collecting windows services info,
originally requested in [telegraf issue #2714](https://github.com/influxdata/telegraf/issues/2714)

## Feature
### Use Cases
- Admin needs to monitor the states of selected windows services on the host, along with a few additional properties (display name, start-up mode)
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
Admin privileges are required to read service info. However, as telegraf mostly runs as a service under the Local System account, it should be no problem.

### Deployment
Feature request mentions the monitoring of 5000 servers. This either means:
* deploying telegraf on each monitored host, which is the preferred option, as the other plugins can be used to monitor other stuff on the host,
* plugin has to monitor multiple servers, which would lead to a more complex plugin (service input plugin) along with complex configuration

## Implementation
### Storing service info
#### Measurement
There are two options to define what a measurement might be
1. Store all service info in single measurement, e.g. win_services, configurable
2. Store service info per service

Option 1. has the biggest benefit in that a user can easily query info about all services, e.g. all services in a stopped state, but this leads to a lot of data in a single measurement, as service measurements have to have the same schema for it to make sense to use this.

Option 2. diversifies the data but makes it difficult to query multiple services

_Q: What is the best practise? I would prefer option 1.. Also, it could be configurable._

#### Properties

Basic properties to monitor and store, as initially requested, are:
 * service name
 * display name
 * startup mode
 * actual state

 There are also additional properties available(see [SERVICE_STATUS_PROCESS](https://msdn.microsoft.com/en-us/library/windows/desktop/ms685992(v=vs.85).aspx) and [QUERY_SERVICE_CONFIG](https://msdn.microsoft.com/en-us/library/windows/desktop/ms684950(v=vs.85).aspx) ), but these are quite advanced and not so suitable for monitoring.

 Display name and startup mode are rather static properties and they will change rarely, but as most probable output is influxdb it will handle this with compaction.

According to the use cases, conditions in queries can be based on _service name_ or a property (state, startup mode).
Services most probably won't be filtered according to display name.

So the mapping to a tag or a field would be:

* **Field** = actual state, startup mode, 
* **Tag** = service name, displayName

 Let's use the following key name and 'types':

 Property|Key | Type
  ---- |----- | ---
 service name| service_name | string
 display name| display_name | string
 startup mode| startup_mode | string
 actual state| state | string

The keys _startupMode_ and _state_ will be a human readable representation of the attribute.

 ### Configuration
 * User must be able to set what services will be monitored:
  ````
  # Case-insensitive name of services to monitor. Empty of all services
  service_names = [
    "LanmanServer",
    "TermService"
  ]
 ````
Services in examples should be available on all Windows editions and versions.

 ### Storing Errors
  There are basically two possible errors:
  * Invalid service name given in configuration
  * A service requires special privileges

  This should be reported as a warning, not as a complete error of the Gather function

   As stated by Daniel, best practice is to use Accumulator.AddError and log it every time. Telegraf should handle multiple instances of the same errors.

 ### Caching
  Most service info is almost static and it could be cached. But as all the info about requested services that Windows Service Manager stores in memory, even a full listing, takes just 8ms on Windows 10 on a Core i5 (2 cores) laptop,
  so caching seems like overhead.
