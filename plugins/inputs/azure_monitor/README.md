# Azure Monitor Input Plugin

The `azure_monitor` plugin, gathers metrics of each Azure
resource using Azure Monitor API. Uses **Logz.io
azure-monitor-metrics-receiver** package -
an SDK wrapper for Azure Monitor SDK.

## Azure Credential

This plugin uses `client_id`, `client_secret` and `tenant_id`
for authentication (access token), and `subscription_id`
is for accessing Azure resources.

## Property Locations

`subscription_id` can be found under **Overview**->**Essentials** in
the Azure portal for your application/service.

`client_id` and `client_secret` can be obtained by registering an
application under Azure Active Directory.

`tenant_id` can be found under **Azure Active Directory**->**Properties**.

resource target `resource_id` can be found under
**Overview**->**Essentials**->**JSON View** (link) in the Azure
portal for your application/service.

## More Information

To see a table of resource types and their metrics, please use this link:

`https://docs.microsoft.com/en-us/azure/azure-monitor/
essentials/metrics-supported`

## Rate Limits

Azure API read limit is 12000 requests per hour.
Please make sure the total number of metrics you are requesting is proportional
to your time interval.

## Usage

Use `resource_targets` to collect metrics from specific resources using
resource id.

Use `resource_group_targets` to collect metrics from resources under the
resource group with resource type.

Use `subscription_targets` to collect metrics from resources under the
subscription with resource type.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather Azure resources metrics from Azure Monitor API
[[inputs.azure_monitor]]
  # can be found under Overview->Essentials in the Azure portal for your application/service
  subscription_id = "<<SUBSCRIPTION_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_id = "<<CLIENT_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_secret = "<<CLIENT_SECRET>>"
  # can be found under Azure Active Directory->Properties
  tenant_id = "<<TENANT_ID>>"

  # resource target #1 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    # can be found undet Overview->Essentials->JSON View in the Azure portal for your application/service
    # must start with 'resourceGroups/...' ('/subscriptions/xxxxxxxx-xxxx-xxxx-xxx-xxxxxxxxxxxx'
    # must be removed from the beginning of Resource ID property value)
    resource_id = "<<RESOURCE_ID>>"
    # the metric names to collect
    # leave the array empty to use all metrics available to this resource
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    # metrics aggregation type value to collect
    # can be 'Total', 'Count', 'Average', 'Minimum', 'Maximum'
    # leave the array empty to collect all aggregation types values for each metric
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource target #2 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    resource_id = "<<RESOURCE_ID>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #1 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    # the resource group name
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      # the resource type
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #2 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # subscription target #1 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # subscription target #2 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
```

## Metrics

* azure_monitor_<<RESOURCE_NAMESPACE>>_<<METRIC_NAME>>
  * fields:
    * total (float64)
    * count (float64)
    * average (float64)
    * minimum (float64)
    * maximum (float64)
  * tags:
    * namespace
    * resource_group
    * resource_name
    * subscription_id
    * resource_region
    * unit

## Example Output

```shell
> azure_monitor_microsoft_storage_storageaccounts_used_capacity,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=9065573,maximum=9065573,minimum=9065573,timeStamp="2021-11-08T09:52:00Z",total=9065573 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_transactions,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Count average=1,count=6,maximum=1,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=6 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_ingress,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=5822.333333333333,count=6,maximum=5833,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=34934 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_egress,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=840.1666666666666,count=6,maximum=841,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=5041 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_success_server_latency,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=MilliSeconds average=12.833333333333334,count=6,maximum=30,minimum=8,timeStamp="2021-11-08T09:52:00Z",total=77 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_success_e2e_latency,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=MilliSeconds average=12.833333333333334,count=6,maximum=30,minimum=8,timeStamp="2021-11-08T09:52:00Z",total=77 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_availability,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Percent average=100,count=6,maximum=100,minimum=100,timeStamp="2021-11-08T09:52:00Z",total=600 1636368744000000000
> azure_monitor_microsoft_storage_storageaccounts_used_capacity,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=9065573,maximum=9065573,minimum=9065573,timeStamp="2021-11-08T09:52:00Z",total=9065573 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_transactions,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Count average=1,count=6,maximum=1,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=6 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_ingress,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=5822.333333333333,count=6,maximum=5833,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=34934 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_egress,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Bytes average=840.1666666666666,count=6,maximum=841,minimum=0,timeStamp="2021-11-08T09:52:00Z",total=5041 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_success_server_latency,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=MilliSeconds average=12.833333333333334,count=6,maximum=30,minimum=8,timeStamp="2021-11-08T09:52:00Z",total=77 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_success_e2e_latency,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=MilliSeconds average=12.833333333333334,count=6,maximum=30,minimum=8,timeStamp="2021-11-08T09:52:00Z",total=77 1636368745000000000
> azure_monitor_microsoft_storage_storageaccounts_availability,host=Azure-MBP,namespace=Microsoft.Storage/storageAccounts,resource_group=azure-rg,resource_name=azuresa,resource_region=eastus,subscription_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,unit=Percent average=100,count=6,maximum=100,minimum=100,timeStamp="2021-11-08T09:52:00Z",total=600 1636368745000000000
```
