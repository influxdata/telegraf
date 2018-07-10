## Azure Monitor Custom Metrics Output for Telegraf

This plugin will send custom metrics to Azure Monitor.

All metrics are written as summarized values: min, max, sum, count. The Telegraf field name is appended to the metric name. All Telegraf tags are set as the metric dimensions.

## Azure Authentication

This plugin can use one of several different types of credentials to authenticate
with the Azure Monitor Custom Metrics ingestion API endpoint. In the following
order the plugin will attempt to authenticate.
1. Managed Service Identity (MSI) token
    - This is the prefered authentication method.
    - Note: MSI is only available to ARM-based resources.
2. AAD Application Tokens (Service Principals)
    - Primarily useful if Telegraf is writing metrics for other resources. [More information](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-application-objects).
    - A Service Principal or User Principal needs to be assigned the `Monitoring Contributor` roles.
3. AAD User Tokens (User Principals)
    - Allows Telegraf to authenticate like a user. It is best to use this method for development.

## Config

For this output plugin to function correctly the following variables
must be configured.

* resourceId
* region

### region

The region is the Azure region that you wish to connect to.
Examples include but are not limited to:
* eastus
* centralus
* westcentralus
* westeurope
* southeastasia

### resourceId

The resourceId used for Azure Monitor metrics.

### Configuration:

```
## Write HTTP timeout, formatted as a string.  If not provided, will default
## to 5s. 0s means no timeout (not recommended).
# timeout = "5s"

## Azure Monitor doesn't have a string value type, so convert string
## fields to dimensions (a.k.a. tags) if enabled. Azure Monitor allows
## a maximum of 10 dimensions so Telegraf will only send the first 10
## alphanumeric dimensions.
#strings_as_dimensions = false

## *The following two fields must be set or be available via the
## Instance Metadata service on Azure Virtual Machines.*
## The Azure Resource ID against which metric will be logged, e.g.
## "/subscriptions/<subscription_id>/resourceGroups/<resource_group>/providers/Microsoft.Compute/virtualMachines/<vm_name>"
#resource_id = ""
## Azure Region to publish metrics against, e.g. eastus, southcentralus.
#region = ""
```
