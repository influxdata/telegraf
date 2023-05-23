# Windows Management Instrumentation Input Plugin

This document presents the input plugin to read WMI classes on Windows
operating systems. With the win_wmi plugin, it is possible to
capture and filter virtually any configuration or metric value exposed
through the Windows Management Instrumentation ([WMI][WMIdoc])
service. At minimum, the telegraf service user must have permission
to [read][ACL] the WMI namespace that is being queried.

[ACL]: https://learn.microsoft.com/en-us/windows/win32/wmisdk/access-to-wmi-namespaces
[WMIdoc]: https://learn.microsoft.com/en-us/windows/win32/wmisdk/wmi-start-page

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Input plugin to query Windows Management Instrumentation
# This plugin ONLY supports Windows
[[inputs.win_wmi]]
  [[inputs.win_wmi.query]]
    # a string representing the WMI namespace to be queried
    namespace = "root\\cimv2"
    # a string representing the WMI class to be queried
    class_name = "Win32_Volume"
    # an array of strings representing the properties of the WMI class to be queried
    properties = ["Name", "Capacity", "FreeSpace"]
    # a string specifying a WHERE clause to use as a filter for the WQL
    filter = 'NOT Name LIKE "\\\\?\\%"'
    # WMI class properties which should be considered tags instead of fields
    tag_properties = ["Name"]
```

### namespace

A string representing the WMI namespace to be queried. For example,
`root\\cimv2`.

### class_name

A string representing the WMI class to be queried. For example,
`Win32_Processor`.

### properties

An array of strings representing the properties of the WMI class to be queried.

### filter

A string specifying a WHERE clause to use as a filter for the WMI Query
Language (WQL). See [WHERE Clause][WHERE] for more information.

[WHERE]: https://learn.microsoft.com/en-us/windows/win32/wmisdk/where-clause?source=recommendations

### tag_properties

Properties which should be considered tags instead of fields.

## Metrics

By default, a WMI class property's value is used as a metric field. If a class
property's value is specified in `tag_properties`, then the value is
instead included with the metric as a tag.

## Troubleshooting

### Errors

If you are getting an error about an invalid WMI namespace, class, or property,
use the `Get-WmiObject` or `Get-CimInstance` PowerShell commands in order to
verify their validity. For example:

```powershell
Get-WmiObject -Namespace root\cimv2 -Class Win32_Volume -Property Capacity, FreeSpace, Name -Filter 'NOT Name LIKE "\\\\?\\%"'
```

```powershell
Get-CimInstance -Namespace root\cimv2 -ClassName Win32_Volume -Property Capacity, FreeSpace, Name -Filter 'NOT Name LIKE "\\\\?\\%"'
```

### Data types

Some WMI classes will return the incorrect data type for a field. In those
cases, it is necessary to use a processor to convert the data type. For
example, the Capacity and FreeSpace properties of the Win32_Volume class must
be converted to integers:

```toml
[[processors.converter]]
  namepass = ["win_wmi_Win32_Volume"]
  [processors.converter.fields]
    integer = ["Capacity", "FreeSpace"]
```

## Example Output

### Physical Memory

This query provides metrics for the speed and capacity of each physical memory
device, along with tags describing the manufacturer, part number, and device
locator of each device.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "root\\cimv2"
    class_name = "Win32_PhysicalMemory"
    properties = [
      "Name",
      "Capacity",
      "DeviceLocator",
      "Manufacturer",
      "PartNumber",
      "Speed",
    ]
    tag_properties = ["Name","DeviceLocator","Manufacturer","PartNumber"]
```

Example Output:

```text
win_wmi_Win32_PhysicalMemory,DeviceLocator=DIMM1,Manufacturer=80AD000080AD,Name=Physical\ Memory,PartNumber=HMA82GU6DJR8N-XN\ \ \ \ ,host=foo Capacity=17179869184i,Speed=3200i 1654269272000000000
```

### Processor

This query provides metrics for the number of cores in each physical processor.
Since the Name property of the WMI class is included by default, the metrics
will also contain a tag value describing the model of each CPU.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "root\\cimv2"
    class_name = "Win32_Processor"
    properties = ["Name","NumberOfCores"]
    tag_properties = ["Name"]
```

Example Output:

```text
win_wmi_Win32_Processor,Name=Intel(R)\ Core(TM)\ i9-10900\ CPU\ @\ 2.80GHz,host=foo NumberOfCores=10i 1654269272000000000
```

### Computer System

This query provides metrics for the number of socketted processors, number of
logical cores on each processor, and the total physical memory in the computer.
The metrics include tag values for the domain, manufacturer, and model of the
computer.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "root\\cimv2"
    class_name = "Win32_ComputerSystem"
    properties = [
      "Name",
      "Domain",
      "Manufacturer",
      "Model",
      "NumberOfLogicalProcessors",
      "NumberOfProcessors",
      "TotalPhysicalMemory"
    ]
    tag_properties = ["Name","Domain","Manufacturer","Model"]
```

Example Output:

```text
win_wmi_Win32_ComputerSystem,Domain=company.com,Manufacturer=Lenovo,Model=X1\ Carbon,Name=FOO,host=foo NumberOfLogicalProcessors=20i,NumberOfProcessors=1i,TotalPhysicalMemory=34083926016i 1654269272000000000
```

### Operating System

This query provides metrics for the paging file's free space, the operating
system's free virtual memory, the operating system SKU installed on the
computer, and the Windows product type. The OS architecture is included as a
tagged value to describe whether the installation is 32-bit or 64-bit.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    class_name = "Win32_OperatingSystem"
    namespace = "root\\cimv2"
    properties = [
      "Name",
      "Caption",
      "FreeSpaceInPagingFiles",
      "FreeVirtualMemory",
      "OperatingSystemSKU",
      "OSArchitecture",
      "ProductType"
    ]
    tag_properties = ["Name","Caption","OSArchitecture"]
```

Example Output:

```text
win_wmi_Win32_OperatingSystem,Caption=Microsoft\ Windows\ 10\ Enterprise,InstallationType=Client,Name=Microsoft\ Windows\ 10\ Enterprise|C:\WINDOWS|\Device\Harddisk0\Partition3,OSArchitecture=64-bit,host=foo FreeSpaceInPagingFiles=5203244i,FreeVirtualMemory=16194496i,OperatingSystemSKU=4i,ProductType=1i 1654269272000000000
```

### Failover Clusters

This query provides a boolean metric describing whether Dynamic Quorum is
enabled for the cluster. The tag values for the metric also include the name of
the Windows Server Failover Cluster and the type of Quorum in use.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "root\\mscluster"
    class_name = "MSCluster_Cluster"
    properties = [
      "Name",
      "QuorumType",
      "DynamicQuorumEnabled"
    ]
    tag_properties = ["Name","QuorumType"]
```

Example Output:

```text
win_wmi_MSCluster_Cluster,Name=testcluster1,QuorumType=Node\ and\ File\ Share\ Majority,host=testnode1 DynamicQuorumEnabled=1i 1671553260000000000
```

### Bitlocker

This query provides a list of volumes which are eligible for bitlocker
encryption and their compliance status. Because the MBAM_Volume class does not
include a Name property, the ExcludeNameKey configuration is included. The
VolumeName property is included in the metric as a tagged value.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "root\\Microsoft\\MBAM"
    class_name = "MBAM_Volume"
    properties = [
      "Compliant",
      "VolumeName"
    ]
    tag_properties = ["VolumeName"]
```

Example Output:

```text
win_wmi_MBAM_Volume,VolumeName=C:,host=foo Compliant=1i 1654269272000000000
```

### SQL Server

This query provides metrics which contain tags describing the version and SKU
of SQL Server. These properties are useful for creating a dashboard of your SQL
Server inventory, which includes the patch level and edition of SQL Server that
is installed.

```toml
[[inputs.win_wmi]]
  name_prefix = "win_wmi_"
  [[inputs.win_wmi.query]]
    namespace = "Root\\Microsoft\\SqlServer\\ComputerManagement15"
    class_name = "SqlServiceAdvancedProperty"
    properties = [
      "PropertyName",
      "ServiceName",
      "PropertyStrValue",
      "SqlServiceType"
    ]
    filter = "ServiceName LIKE 'MSSQLSERVER' AND SqlServiceType = 1 AND (PropertyName LIKE 'FILEVERSION' OR PropertyName LIKE 'SKUNAME')"
    tag_properties = ["PropertyName","ServiceName","PropertyStrValue"]
```

Example Output:

```text
win_wmi_SqlServiceAdvancedProperty,PropertyName=FILEVERSION,PropertyStrValue=2019.150.4178.1,ServiceName=MSSQLSERVER,host=foo,sqlinstance=foo SqlServiceType=1i 1654269272000000000
win_wmi_SqlServiceAdvancedProperty,PropertyName=SKUNAME,PropertyStrValue=Developer\ Edition\ (64-bit),ServiceName=MSSQLSERVER,host=foo,sqlinstance=foo SqlServiceType=1i 1654269272000000000
```
