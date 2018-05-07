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
* useast
* centralus
* westcentralus
* westeurope
* southeastasia

### resourceId

The resourceId used for Azure Monitor metrics.

### Configuration:

```
# Configuration for sending aggregate metrics to Azure Monitor
[[outputs.azuremonitor]]
## The resource ID against which metric will be logged.  If not
## specified, the plugin will attempt to retrieve the resource ID
## of the VM via the instance metadata service (optional if running 
## on an Azure VM with MSI)
#resource_id = "/subscriptions/<subscription_id>/resourceGroups/<resource_group>/providers/Microsoft.Compute/virtualMachines/<vm_name>"
## Azure region to publish metrics against.  Defaults to eastus.
## Leave blank to automatically query the region via MSI.
#region = "useast"

## Write HTTP timeout, formatted as a string.  If not provided, will default
## to 5s. 0s means no timeout (not recommended).
# timeout = "5s"

## Whether or not to use managed service identity.
#useManagedServiceIdentity = true

## Fill in the following values if using Active Directory Service
## Principal or User Principal for authentication.
## Subscription ID
#azureSubscription = ""
## Tenant ID
#azureTenant = ""
## Client ID
#azureClientId = ""
## Client secrete
#azureClientSecret = ""
```
