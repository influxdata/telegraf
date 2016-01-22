# win_perfcounters readme

The way this plugin works is that on load of Telegraf, the plugin will be handed configuration from Telegraf. This configuration is parsed and then tested for validity such as if the Object, Instance and Counter existing. If it does not match at startup, it will not be fetched. Exceptions to this are in cases where you query for all instances "*". By default the plugin does not return _Total when it is querying for all (*) as this is redundant.

## Basics

The examples contained in this file have been found on the internet as counters used when performance monitoring Active Directory and IIS in perticular. There are a lot other good objects to monitor, if you know what to look for. This file is likely to be updated in the future with more examples for useful configurations for separate scenarios.

### Entry
A new configuration entry consists of the TOML header to start with, `[[inputs.win_perfcounters.object]]`. This must follow before other plugins configuration, beneath the main win_perfcounters entry, `[[inputs.win_perfcounters]]`.

Following this is 3 required key/value pairs and the three optional parameters and their usage.

### ObjectName
**Required**

ObjectName is the Object to query for, like Processor, DirectoryServices, LogicalDisk or similar.

Example: `ObjectName = "LogicalDisk"`

### Instances
**Required**

Instances (this is an array) is the instances of a counter you would like returned, it can be one or more values. 

Example, `Instances = ["C:","D:","E:"]` will return only for the instances C:, D: and E: where relevant. To get all instnaces of a Counter, use ["*"] only. By default any results containing _Total are stripped, unless this is specified as the wanted instance. Alternatively see the option IncludeTotal below.

Some Objects does not have instances to select from at all, here only one option is valid if you want data back, and that is to specify `Instances = ["------"]`.

### Counters
**Required**

Counters (this is an array) is the counters of the ObjectName you would like returned, it can also be one or more values.

Example: `Counters = ["% Idle Time", "% Disk Read Time", "% Disk Write Time"]`
This must be specified for every counter you want the results of, it is not possible to ask for all counters in the ObjectName.

### Measurement
*Optional*

This key is optional, if it is not set it will be win_perfcounters. In InfluxDB this is the key by which the returned data is stored underneath, so for ordering your data in a good manner, this is a good key to set with where you want your IIS and Disk results stored, separate from Processor results.

Example: `Measurement = "win_disk"

### IncludeTotal
*Optional* 

This key is optional, it is a simple bool. If it is not set to true or included it is treated as false.
This key only has an effect if Instances is set to "*" and you would also like all instances containg _Total returned, like "_Total", "0,_Total" and so on where applicable (Processor Information is one example).

### WarnOnMissing
*Optional*

This key is optional, it is a simple bool. If it is not set to true or included it is treated as false.
This only has an effect on the first execution of the plugin, it will print out any ObjectName/Instance/Counter combinations asked for that does not match. Useful when debugging new configurations.

### FailOnMissing
*Internal*

This key should not be used, it is for testing purposes only. It is a simple bool, if it is not set to true or included this is treaded as false. If this is set to true, the plugin will abort and end prematurely if any of the combinations of ObjectName/Instances/Counters are invalid.

## Examples

### Generic Queries
```

  [[inputs.win_perfcounters.object]]
    # Processor usage, alternative to native, reports on a per core.
    ObjectName = "Processor"
    Instances = ["*"]
    Counters = ["% Idle Time", "% Interrupt Time", "% Privileged Time", "% User Time", "% Processor Time"]
    Measurement = "win_cpu"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # Disk times and queues
    ObjectName = "LogicalDisk"
    Instances = ["*"]
    Counters = ["% Idle Time", "% Disk Time","% Disk Read Time", "% Disk Write Time", "% User Time", "Current Disk Queue Length"]
    Measurement = "win_disk"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    ObjectName = "System"
    Counters = ["Context Switches/sec","System Calls/sec"]
    Instances = ["------"]
    Measurement = "win_system"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # Example query where the Instance portion must be removed to get data back, such as from the Memory object.
    ObjectName = "Memory"
    Counters = ["Available Bytes","Cache Faults/sec","Demand Zero Faults/sec","Page Faults/sec","Pages/sec","Transition Faults/sec","Pool Nonpaged Bytes","Pool Paged Bytes"]
    Instances = ["------"] # Use 6 x - to remove the Instance bit from the query.
    Measurement = "win_mem"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```

### Active Directory Domain Controller
```
  [[inputs.win_perfcounters.object]]
    ObjectName = "DirectoryServices"
    Instances = ["*"]
    Counters = ["Base Searches/sec","Database adds/sec","Database deletes/sec","Database modifys/sec","Database recycles/sec","LDAP Client Sessions","LDAP Searches/sec","LDAP Writes/sec"]
    Measurement = "win_ad" # Set an alternative measurement to win_perfcounters if wanted.
    #Instances = [""] # Gathers all instances by default, specify to only gather these
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    ObjectName = "Security System-Wide Statistics"
    Instances = ["*"]
    Counters = ["NTLM Authentications","Kerberos Authentications","Digest Authentications"]
    Measurement = "win_ad"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    ObjectName = "Database"
    Instances = ["*"]
    Counters = ["Database Cache % Hit","Database Cache Page Fault Stalls/sec","Database Cache Page Faults/sec","Database Cache Size"]
    Measurement = "win_db"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```

### DFS Namespace + Domain Controllers
```
  [[inputs.win_perfcounters.object]]
    # AD, DFS N, Useful if the server hosts a DFS Namespace or is a Domain Controller
    ObjectName = "DFS Namespace Service Referrals"
    Instances = ["*"]
    Counters = ["Requests Processed","Requests Failed","Avg. Response Time"]
    Measurement = "win_dfsn"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
    #WarnOnMissing = false # Print out when the performance counter is missing, either of object, counter or instance.
```
### DFS Replication + Domain Controllers
```
  [[inputs.win_perfcounters.object]]
    # AD, DFS R, Useful if the server hosts a DFS Replication folder or is a Domain Controller
    ObjectName = "DFS Replication Service Volumes"
    Instances = ["*"]
    Counters = ["Data Lookups","Database Commits"]
    Measurement = "win_dfsr"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
    #WarnOnMissing = false # Print out when the performance counter is missing, either of object, counter or instance.
```
### DNS Server + Domain Controllers
```
  [[inputs.win_perfcounters.object]]
    ObjectName = "DNS"
    Counters = ["Dynamic Update Received","Dynamic Update Rejected","Recursive Queries","Recursive Queries Failure","Secure Update Failure","Secure Update Received","TCP Query Received","TCP Response Sent","UDP Query Received","UDP Response Sent","Total Query Received","Total Response Sent"]
    Instances = ["------"]
    Measurement = "win_dns"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```
### IIS / ASP.NET 
```
  [[inputs.win_perfcounters.object]]
    # HTTP Service request queues in the Kernel before being handed over to User Mode.
    ObjectName = "HTTP Service Request Queues"
    Instances = ["*"]
    Counters = ["CurrentQueueSize","RejectedRequests"]
    Measurement = "win_http_queues"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # IIS, ASP.NET Applications
    ObjectName = "ASP.NET Applications"
    Counters = ["Cache Total Entries","Cache Total Hit Ratio","Cache Total Turnover Rate","Output Cache Entries","Output Cache Hits","Output Cache Hit Ratio","Output Cache Turnover Rate","Compilations Total","Errors Total/Sec","Pipeline Instance Count","Requests Executing","Requests in Application Queue","Requests/Sec"]
    Instances = ["*"]
    Measurement = "win_aspnet_app"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # IIS, ASP.NET
    ObjectName = "ASP.NET"
    Counters = ["Application Restarts","Request Wait Time","Requests Current","Requests Queued","Requests Rejected"]
    Instances = ["*"]
    Measurement = "win_aspnet"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # IIS, Web Service
    ObjectName = "Web Service"
    Counters = ["Get Requests/sec","Post Requests/sec","Connection Attempts/sec","Current Connections","ISAPI Extension Requests/sec"]
    Instances = ["*"]
    Measurement = "win_websvc"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # Web Service Cache / IIS
    ObjectName = "Web Service Cache"
    Counters = ["URI Cache Hits %","Kernel: URI Cache Hits %","File Cache Hits %"]
    Instances = ["*"]
    Measurement = "win_websvc_cache"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```
### Process
```
  [[inputs.win_perfcounters.object]]
    # Process metrics, in this case for IIS only
    ObjectName = "Process"
    Counters = ["% Processor Time","Handle Count","Private Bytes","Thread Count","Virtual Bytes","Working Set"]
    Instances = ["w3wp"]
    Measurement = "win_proc"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```
### .NET Montioring
  [[inputs.win_perfcounters.object]]
    # .NET CLR Exceptions, in this case for IIS only
    ObjectName = ".NET CLR Exceptions"
    Counters = ["# of Exceps Thrown / sec"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_exceptions"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # .NET CLR Jit, in this case for IIS only
    ObjectName = ".NET CLR Jit"
    Counters = ["% Time in Jit","IL Bytes Jitted / sec"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_jit"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # .NET CLR Loading, in this case for IIS only
    ObjectName = ".NET CLR Loading"
    Counters = ["% Time Loading"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_loading"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # .NET CLR LocksAndThreads, in this case for IIS only
    ObjectName = ".NET CLR LocksAndThreads"
    Counters = ["# of current logical Threads","# of current physical Threads","# of current recognized threads","# of total recognized threads","Queue Length / sec","Total # of Contentions","Current Queue Length"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_locks"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # .NET CLR Memory, in this case for IIS only
    ObjectName = ".NET CLR Memory"
    Counters = ["% Time in GC","# Bytes in all Heaps","# Gen 0 Collections","# Gen 1 Collections","# Gen 2 Collections","# Induced GC","Allocated Bytes/sec","Finalization Survivors","Gen 0 heap size","Gen 1 heap size","Gen 2 heap size","Large Object Heap size","# of Pinned Objects"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_mem"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

  [[inputs.win_perfcounters.object]]
    # .NET CLR Security, in this case for IIS only
    ObjectName = ".NET CLR Security"
    Counters = ["% Time in RT checks","Stack Walk Depth","Total Runtime Checks"]
    Instances = ["w3wp"]
    Measurement = "win_dotnet_security"
    #IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
```
