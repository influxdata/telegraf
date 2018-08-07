## Azure Monitor Custom Metrics Output for Telegraf

This plugin will send custom metrics to Azure Monitor. Azure Monitor has a
metric resolution of one minute. To handle this in Telegraf, the Azure Monitor
output plugin will automatically aggregates metrics into one minute buckets,
which are then sent to Azure Monitor on every flush interval.

The metrics from each input plugin will be written to a separate Azure Monitor
namespace, prefixed with `Telegraf/` by default. The field name for each
metric is written as the Azure Monitor metric name. All field values are
written as a summarized set that includes: min, max, sum, count. Tags are
written as a dimension on each Azure Monitor metric.

Since Azure Monitor only accepts numeric values, string-typed fields are
dropped by default. There is a configuration option (`strings_as_dimensions`)
to retain fields that contain strings as extra dimensions. Azure Monitor
allows a maximum of 10 dimensions per metric so any dimensions over that
amount will be deterministically dropped.

## Initial Setup

1. [Register your Azure subscription with the `microsoft.insights` resource
   provider.](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-supported-services#portal)
2. [Consult this chart to identify which regions support Azure Monitor.](https://azure.microsoft.com/en-us/global-infrastructure/services/)
3. Only some Azure Monitor regions support Custom Metrics. For regions with
   Custom Metrics support, an endpoint will be available with the format
   `https://<region>.monitoring.azure.com`. The following regions are
   currently known to be supported:
    - West Central US, e.g. `https://westcentralus.monitoring.azure.com`
    - South Central US, e.g. `https://southcentralus.monitoring.azure.com`

## Azure Authentication

This plugin uses one of several different types of authenticate methods. The
preferred authentication methods are different from the *order* in which each
authentication is checked. Here are the preferred authentication methods:

1. Managed Service Identity (MSI) token
    - This is the prefered authentication method. Telegraf will automatically
      authenticate using this method when running on Azure VMs.
2. AAD Application Tokens (Service Principals)
    - Primarily useful if Telegraf is writing metrics for other resources. [More
      information](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-application-objects).
    - A Service Principal or User Principal needs to be assigned the `Monitoring
      Contributor` roles.
3. AAD User Tokens (User Principals)
    - Allows Telegraf to authenticate like a user. It is best to use this method
      for development.

The plugin will attempt to authenticate with the first available of the
following configurations in this order:

1. **Client Credentials**: Azure AD Application ID and Secret.

    Set the following Telegraf configuration variables:

    - `azure_tenant_id`: Specifies the Tenant to which to authenticate.
    - `azure_client_id`: Specifies the app client ID to use.
    - `azure_client_secret`: Specifies the app secret to use.

    Or set the following environment variables:

    - `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    - `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    - `AZURE_CLIENT_SECRET`: Specifies the app secret to use.

2. **Client Certificate**: Azure AD Application ID and X.509 Certificate.

    - `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    - `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    - `AZURE_CERTIFICATE_PATH`: Specifies the certificate Path to use.
    - `AZURE_CERTIFICATE_PASSWORD`: Specifies the certificate password to use.

3. **Resource Owner Password**: Azure AD User and Password. This grant type is
   *not recommended*, use device login instead if you need interactive login.

    - `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    - `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    - `AZURE_USERNAME`: Specifies the username to use.
    - `AZURE_PASSWORD`: Specifies the password to use.

4. **Azure Managed Service Identity**: Delegate credential management to the
   platform. Requires that code is running in Azure, e.g. on a VM. All
   configuration is handled by Azure. See [Azure Managed Service
   Identity](https://docs.microsoft.com/en-us/azure/active-directory/msi-overview)
   for more details. Only available on ARM-based resources.

**Note: As shown above, the last option (#4) is the preferred way to
authenticate when running Telegraf on Azure VMs. Make sure you've followed the
[initial setup instructions](#initial-setup).**

## Config

The plugin will automatically attempt to discover the region and resource ID
using the Azure VM Instance Metadata service. If Telegraf is not running on a
virtual machine or the VM Instance Metadata service is not available, the
following variables are required for the output to function.

* region
* resource_id

### Configuration:

```
[[outputs.azure_monitor]]
  ## See the [Azure Monitor output plugin README](/plugins/outputs/azure_monitor/README.md)
  ## for details on authentication options.

  ## Write HTTP timeout, formatted as a string. Defaults to 20s.
  #timeout = "20s"

  ## Set the namespace prefix, defaults to "Telegraf/<input-name>".
  #namespace_prefix = "Telegraf/"

  ## Azure Monitor doesn't have a string value type, so convert string
  ## fields to dimensions (a.k.a. tags) if enabled. Azure Monitor allows
  ## a maximum of 10 dimensions so Telegraf will only send the first 10
  ## alphanumeric dimensions.
  #strings_as_dimensions = false

  ## *The following two fields must be set or be available via the
  ## Instance Metadata service on Azure Virtual Machines.*

  ## Azure Region to publish metrics against, e.g. eastus, southcentralus.
  #region = ""

  ## The Azure Resource ID against which metric will be logged, e.g.
  ## "/subscriptions/<subscription_id>/resourceGroups/<resource_group>/providers/Microsoft.Compute/virtualMachines/<vm_name>"
  #resource_id = ""
```
