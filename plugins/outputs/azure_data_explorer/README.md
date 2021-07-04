# Azure Data Explorer output plugin

This plugin writes metrics collected by any of the input plugins of Telegraf to [Azure Data Explorer](https://azure.microsoft.com/en-au/services/data-explorer/). 

## Pre-requisites:
- [Create Azure Data Explorer cluster and database](https://docs.microsoft.com/en-us/azure/data-explorer/create-cluster-database-portal)
- VM/compute or container to host Telegraf - it could be hosted locally where an app/services to be monitored are deployed or remotely on a dedicated monitoring compute/container.


## Configuration:

```toml
[[outputs.azure_data_explorer]]
  ## The URI property of the Azure Data Explorer resource on Azure
  ## ex: https://myadxresource.australiasoutheast.kusto.windows.net
  # endpoint_url = ""

  ## The Azure Data Explorer database that the metrics will be ingested into.
  ## The plugin will NOT generate this database automatically, it's expected that this database already exists before ingestion.
  ## ex: "exampledatabase"
  # database = ""

  ## Timeout for Azure Data Explorer operations
  # timeout = "15s"

```

## Metrics Grouping

The plugin will group the metrics by the metric name, and will send each group of metrics to an Azure Data Explorer table. If the table doesn't exist the plugin will create the table, if the table exists then the plugin will try to merge the Telegraf metric schema to the existing table. For more information about the merge process check the [`.create-merge` documentation](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/create-merge-table-command).

The table name will match the `name` property of the metric, this means that the name of the metric should comply with the Azure Data Explorer table naming constraints in case you plan to add a prefix to the metric name.

## Tables Schema

The schema of the Azure Data Explorer table will match the structure of the Telegraf `Metric` object. The corresponding Azure Data Explorer command would be like the following:
```
.create-merge table ['table-name']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime)
```

The corresponding table mapping would be like the following:
```
.create-or-alter table ['table-name'] ingestion json mapping 'table-name_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'
```

**Note**: This plugin will automatically create Azure Data Explorer tables and corresponding table mapping as per the above mentioned commands. Since the `Metric` object is a complex type, the only output format supported is JSON, so make sure to set the `data_format` configuration in `telegraf.conf` to `json`.

## Authentiation

### Supported Authentication Methods
This plugin provides several types of authentication. The plugin will check the existence of several specific environment variables, and consequently will choose the right method. 

These methods are:


1. AAD Application Tokens (Service Principals with secrets or certificates).

    For guidance on how to create and register an App in Azure Active Directory check [this article](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app#register-an-application), and for more information on the Service Principals check [this article](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals).


2. AAD User Tokens 
    - Allows Telegraf to authenticate like a user. It is best to use this method
      for development.

3. Managed Service Identity (MSI) token
    - If you are running Telegraf from Azure VM or infrastructure, then this is the prefered authentication method. 

[principal]: https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-application-objects

Whichever method, the designated Principal needs to be assigned the `Database User` role on the Database level in the Azure Data Explorer. This role will allow the plugin to create the required tables and ingest data into it.

### Choosing The Authentication Method

The plugin will authenticate using the first available of the
following configurations, **it's important to understand that the assessment, and consequently choosing the authentication method, will happen in order as below**:

1. **Client Credentials**: Azure AD Application ID and Secret.

    Set the following environment variables:

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
   configuration is handled by Azure. See [Azure Managed Service Identity][msi]
   for more details. Only available when using the [Azure Resource Manager][arm].

[msi]: https://docs.microsoft.com/en-us/azure/active-directory/msi-overview
[arm]: https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-overview


## Querying
With all above configurations, you will have data stored in following standard format for each metric type stored as an Azure Data Explorer table -
ColumnName | ColumnType
---------- | ----------
fields	|	dynamic
name	|	string
tags	|	dynamic
timestamp	|	datetime

As "fields" and "tags" are of dynamic data type so following multiple ways to query this data -
1. **Query JSON attributes directly**: This is one of the coolest feature of Azure Data Explorer so you can run query like this -
   ```
   Tablename
   | where fields.size_kb == 9120
   ```
2. **Use [Update policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/updatepolicy)**: to transform data, in this case, to flatten dynamic data type columns**: This is the recommended performant way for querying over large data volumes compared to querying directly over JSON attributes. 
      ```
      // Function to transform data
      .create-or-alter function Transform_TargetTableName() {
             SourceTableName 
             | extend clerk_type = tags.clerk_type
             | extend host = tags.host
      } 

      // Create the destination table (if it doesn't exist already)
      .set-or-append TargetTableName <| Transform_TargetTableName() | limit 0

      // Apply update policy on destination table
      .alter table TargetTableName policy update
      @'[{"IsEnabled": true, "Source": "SourceTableName", "Query": "Transform_TargetTableName()", "IsTransactional": false, "PropagateIngestionProperties": false}]'

      ```
      There are two ways to flatten dynamic columns as explained below. You can use either of these ways in above mentioned update policy function - 'Transform_TargetTableName()'
    - Use [bag_unpack plugin](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/query/bag-unpackplugin) to unpack the dynamic columns as shown below. This method will unpack all columns, it could lead to issues in case source schema changes.
       ```
       Tablename
       | evaluate bag_unpack(tags)
       | evaluate bag_unpack(fields)
       ```
    
    - Use [extend](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/query/extendoperator) operator as shown below. This is the best way provided you know what columns are needed in the final destination table. Another benefit of this method is even if schema changes, it will not break your queries or dashboards.
       ```
       Tablename
       | extend clerk_type = tags.clerk_type
       | extend host = tags.host
       ```

