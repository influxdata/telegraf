# Azure Monitor Output Plugin

__The Azure Monitor custom metrics service is currently in preview and not
available in a subset of Azure regions.__

This plugin will send custom metrics to Azure Monitor. Azure Monitor has a
metric resolution of one minute. To handle this in Telegraf, the Azure Monitor
output plugin will automatically aggregates metrics into one minute buckets,
which are then sent to Azure Monitor on every flush interval.

The metrics from each input plugin will be written to a separate Azure Monitor
namespace, prefixed with `Telegraf/` by default. The field name for each metric
is written as the Azure Monitor metric name. All field values are written as a
summarized set that includes: min, max, sum, count. Tags are written as a
dimension on each Azure Monitor metric.

## Configuration

```toml
# Send aggregate metrics to Azure Monitor
[[outputs.azure_monitor]]
  ## Timeout for HTTP writes.
  # timeout = "20s"

  ## Set the namespace prefix, defaults to "Telegraf/<input-name>".
  # namespace_prefix = "Telegraf/"

  ## Azure Monitor doesn't have a string value type, so convert string
  ## fields to dimensions (a.k.a. tags) if enabled. Azure Monitor allows
  ## a maximum of 10 dimensions so Telegraf will only send the first 10
  ## alphanumeric dimensions.
  # strings_as_dimensions = false

  ## Both region and resource_id must be set or be available via the
  ## Instance Metadata service on Azure Virtual Machines.
  #
  ## Azure Region to publish metrics against.
  ##   ex: region = "southcentralus"
  # region = ""
  #
  ## The Azure Resource ID against which metric will be logged, e.g.
  ##   ex: resource_id = "/subscriptions/<subscription_id>/resourceGroups/<resource_group>/providers/Microsoft.Compute/virtualMachines/<vm_name>"
  # resource_id = ""

  ## Optionally, if in Azure US Government, China, or other sovereign
  ## cloud environment, set the appropriate REST endpoint for receiving
  ## metrics. (Note: region may be unused in this context)
  # endpoint_url = "https://monitoring.core.usgovcloudapi.net"
```

## Setup

1. [Register the `microsoft.insights` resource provider in your Azure
   subscription][resource provider].
1. If using Managed Service Identities to authenticate an Azure VM, [enable
   system-assigned managed identity][enable msi].
1. Use a region that supports Azure Monitor Custom Metrics, For regions with
   Custom Metrics support, an endpoint will be available with the format
   `https://<region>.monitoring.azure.com`.

[resource provider]: https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-supported-services

[enable msi]: https://docs.microsoft.com/en-us/azure/active-directory/managed-service-identity/qs-configure-portal-windows-vm

### Region and Resource ID

The plugin will attempt to discover the region and resource ID using the Azure
VM Instance Metadata service. If Telegraf is not running on a virtual machine or
the VM Instance Metadata service is not available, the following variables are
required for the output to function.

* region
* resource_id

### Authentication

This plugin uses one of several different types of authenticate methods. The
preferred authentication methods are different from the *order* in which each
authentication is checked. Here are the preferred authentication methods:

1. Managed Service Identity (MSI) token: This is the preferred authentication
   method. Telegraf will automatically authenticate using this method when
   running on Azure VMs.
2. AAD Application Tokens (Service Principals)

    * Primarily useful if Telegraf is writing metrics for other resources.
      [More information][principal].
    * A Service Principal or User Principal needs to be assigned the `Monitoring
      Metrics Publisher` role on the resource(s) metrics will be emitted
      against.

3. AAD User Tokens (User Principals)

    * Allows Telegraf to authenticate like a user. It is best to use this method
      for development.

[principal]: https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-application-objects

The plugin will authenticate using the first available of the following
configurations:

1. **Client Credentials**: Azure AD Application ID and Secret. Set the following
   environment variables:

    * `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    * `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    * `AZURE_CLIENT_SECRET`: Specifies the app secret to use.

1. **Client Certificate**: Azure AD Application ID and X.509 Certificate.

    * `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    * `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    * `AZURE_CERTIFICATE_PATH`: Specifies the certificate Path to use.
    * `AZURE_CERTIFICATE_PASSWORD`: Specifies the certificate password to use.

1. **Resource Owner Password**: Azure AD User and Password. This grant type is
   *not recommended*, use device login instead if you need interactive login.

    * `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.
    * `AZURE_CLIENT_ID`: Specifies the app client ID to use.
    * `AZURE_USERNAME`: Specifies the username to use.
    * `AZURE_PASSWORD`: Specifies the password to use.

1. **Azure Managed Service Identity**: Delegate credential management to the
   platform. Requires that code is running in Azure, e.g. on a VM. All
   configuration is handled by Azure. See [Azure Managed Service Identity][msi]
   for more details. Only available when using the [Azure Resource
   Manager][arm].

[msi]: https://docs.microsoft.com/en-us/azure/active-directory/msi-overview
[arm]: https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-overview

**Note: As shown above, the last option (#4) is the preferred way to
authenticate when running Telegraf on Azure VMs.

## Dimensions

Azure Monitor only accepts values with a numeric type. The plugin will drop
fields with a string type by default. The plugin can set all string type fields
as extra dimensions in the Azure Monitor custom metric by setting the
configuration option `strings_as_dimensions` to `true`.

Keep in mind, Azure Monitor allows a maximum of 10 dimensions per metric. The
plugin will deterministically dropped any dimensions that exceed the 10
dimension limit.

To convert only a subset of string-typed fields as dimensions, enable
`strings_as_dimensions` and use the [`fieldpass` or `fielddrop`
processors][conf-processor] to limit the string-typed fields that are sent to
the plugin.

[conf-processor]: https://docs.influxdata.com/telegraf/v1.7/administration/configuration/#processor-configuration
