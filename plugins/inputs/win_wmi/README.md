# Windows Management Interface Input Plugin

This document presents the input plugin to read WMI classes on Windows
operating systems. With the win_wmi plugin, an administrator is enabled to
capture and filter virtually any configuration or metric value exposed through
the Windows Management Instrumentation service.

If a WMI class property's value is a string, then the string is included with
the metric as a tag. If a WMI class property's value is an integer, then the
integer is used as a metric field.

If telegraf is configured with a logfile and the plugin's configuration
contains an invalid namespace, class, or property, an error is logged.

## Basics

The examples contained in this file have been utilized and exercised in Windows
environments. There are many other useful classes to monitor if you know what
to look for.

### Namespace

A string representing the WMI namespace to be queried. For example,
`root\\cimv2`.

### ClassName

A string representing the WMI class to be queried. For example,
`Win32_Processor`.

### Properties

An array of strings representing the properties of the WMI class to be queried.
By default, the `Name` property is included in the query. However, some classes
do not contain a `Name` property, in which cases the `ExcludeNameKey`
configuration should be utilized.

### ExcludeNameKey

By default, a WMI class's `Name` property is included as a tag value in order
for each metric to have a unique identifier. However, some WMI classes do not
have a `Name` property. In such cases, ExcludeNameKey should be set to `True`.

## Configuration

```toml @sample.conf
  ## By default, this plugin returns no results.
  ## Uncomment the example below or write your own as you see fit.
  ## The "Name" property of a WMI class is automatically included unless
  ## excludenamekey is true.
  ## If the WMI property's value is a string, then it is used as a tag.
  ## If the WMI property's value is a type of int, then it is used as a field.
  ## [[inputs.win_wmi]]
  ##   namespace = "root\\cimv2"
  ##   classname = "Win32_Volume"
  ##   properties = ["Capacity", "FreeSpace"]
  ##   filter = 'NOT Name LIKE "\\\\?\\%"'
  ##   excludenamekey = false
  ##   name_prefix = "win_wmi_"
```

### Generic Queries

This query provides metrics for the speed and capacity of each physical memory
device, along with tags describing the manufacturer, part number, and device
locator of each device.

```toml
[[inputs.win_wmi]]
  namespace = "root\\cimv2"
  classname = "Win32_PhysicalMemory"
  properties = [
    "Capacity",
    "DeviceLocator",
    "Manufacturer",
    "PartNumber",
    "Speed",
  ]
  name_prefix = "win_wmi_"
```

This query provides metrics for the number of cores in each physical processor.
Since the Name property of the WMI class is included by default, the metrics
will also contain a tag value describing the model of each CPU.

```toml
[[inputs.win_wmi]]
  namespace = "root\\cimv2"
  classname = "Win32_Processor"
  properties = [
    "NumberOfCores"
  ]
  name_prefix = "win_wmi_"
```

This query provides metrics for the number of socketted processors, number of
logical cores on each processor, and the total physical memory in the computer.
The metrics include tag values for the domain, manufacturer, and model of the
computer.

```toml
[[inputs.win_wmi]]
  namespace = "root\\cimv2"
  classname = "Win32_ComputerSystem"
  properties = [
    "Domain",
    "Manufacturer",
    "Model",
    "NumberOfLogicalProcessors",
    "NumberOfProcessors",
    "TotalPhysicalMemory"
  ]
  name_prefix = "win_wmi_"
```

This query provides metrics for the paging file's free space, the operating
system's free virtual memory, the operating system SKU installed on the
computer, and the Windows product type. The OS architecture is included as a
tagged value to describe whether the installation is 32-bit or 64-bit.

```toml
[[inputs.win_wmi]]
classname = "Win32_OperatingSystem"
name_prefix = "win_wmi_"
namespace = "root\\cimv2"
properties = [
  "Caption",
  "FreeSpaceInPagingFiles",
  "FreeVirtualMemory",
  "OperatingSystemSKU",
  "OSArchitecture",
  "ProductType"
]
```

### Failover Clusters

This query provides a boolean metric describing whether Dynamic Quorum is
enabled for the cluster. The tag values for the metric also include the name of
the Windows Server Failover Cluster and the type of Quorum in use.

```toml
[[inputs.win_wmi]]
  namespace = "root\\mscluster"
  classname = "MSCluster_Cluster"
  properties = [
    "QuorumType",
    "DynamicQuorumEnabled"
  ]
  name_prefix = "win_wmi_MSCluster"
```

### Bitlocker

This query provides a list of volumes which are eligible for bitlocker
encryption and their compliance status. Because the MBAM_Volume class does not
include a Name property, the ExcludeNameKey configuration is included. The
VolumeName property is included in the metric as a tagged value.

```toml
[[inputs.win_wmi]]
  namespace = "root\\Microsoft\\MBAM"
  classname = "MBAM_Volume"
  properties = [
    "Compliant",
    "VolumeName"
  ]
  excludenamekey = true
  name_prefix = "win_wmi_"
```

### SQL Server

This query provides metrics which contain tags describing the version and SKU
of SQL Server. These properties are useful for creating a dashboard of your SQL
Server inventory, which includes the patch level and edition of SQL Server that
is installed.

```toml
[[inputs.win_wmi]]
  namespace = "Root\\microsoft\\sqlserver\\ComputerManagement15"
  classname = "SqlServiceAdvancedProperty"
  properties = [
    "PropertyName",
    "ServiceName",
    "PropertyStrValue",
    "SqlServiceType"
  ]
  filter = "ServiceName LIKE 'MSSQLSERVER' AND SqlServiceType = 1 AND (PropertyName LIKE 'FILEVERSION' OR PropertyName LIKE 'SKUNAME')"
  excludenamekey = true
  name_prefix = "win_wmi_"
```

## Troubleshooting

If you are getting an error about an invalid WMI namespace, class, or property,
use the `Get-WmiObject` or `Get-CimInstance` PowerShell commands in order to
verify their validity. For example:

```powershell
Get-WmiObject -Namespace root\cimv2 -Class Win32_Volume -Property Capacity, FreeSpace, Name -Filter 'NOT Name LIKE "\\\\?\\%"'
```

```powershell
Get-CimInstance -Namespace root\cimv2 -ClassName Win32_Volume -Property Capacity, FreeSpace, Name -Filter 'NOT Name LIKE "\\\\?\\%"'
```

## Example Output

Some values are changed for anonymity.

```text
> win_wmi_MBAM_Volume,VolumeName=C:,host=foo Compliant=1i 1654269272000000000
> win_wmi_MSFT_NetAdapter,Name=Ethernet,host=foo Speed=1000000000i 1654269272000000000
> win_wmi_SqlServiceAdvancedProperty,PropertyName=FILEVERSION,PropertyStrValue=2019.150.4178.1,ServiceName=MSSQLSERVER,host=foo,sqlinstance=foo SqlServiceType=1i 1654269272000000000
> win_wmi_SqlServiceAdvancedProperty,PropertyName=SKUNAME,PropertyStrValue=Developer\ Edition\ (64-bit),ServiceName=MSSQLSERVER,host=foo,sqlinstance=foo SqlServiceType=1i 1654269272000000000
> win_wmi_Win32_ComputerSystem,Domain=company.com,Manufacturer=Lenovo,Model=X1\ Carbon,Name=FOO,host=foo NumberOfLogicalProcessors=20i,NumberOfProcessors=1i,TotalPhysicalMemory=34083926016i 1654269272000000000
> win_wmi_Win32_OperatingSystem,Caption=Microsoft\ Windows\ 10\ Enterprise,InstallationType=Client,Name=Microsoft\ Windows\ 10\ Enterprise|C:\WINDOWS|\Device\Harddisk0\Partition3,OSArchitecture=64-bit,SiteCode=NYC,host=foo FreeSpaceInPagingFiles=5203244i,FreeVirtualMemory=16194496i,OperatingSystemSKU=4i,ProductType=1i 1654269272000000000
> win_wmi_Win32_PhysicalMemory,DeviceLocator=DIMM1,Manufacturer=80AD000080AD,Name=Physical\ Memory,PartNumber=HMA82GU6DJR8N-XN\ \ \ \ ,host=foo Capacity=17179869184i,Speed=3200i 1654269272000000000
> win_wmi_Win32_Processor,Name=Intel(R)\ Core(TM)\ i9-10900\ CPU\ @\ 2.80GHz,host=foo NumberOfCores=10i 1654269272000000000
> win_wmi_Win32_Volume,Name=C:,host=foo Capacity=511870046208i,FreeSpace=276193509376i 1654269272000000000
```
