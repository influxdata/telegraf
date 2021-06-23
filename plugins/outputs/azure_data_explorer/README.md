# Azure Data Explorer

The Azure Data Explorer plugin sends Telegraf metrics to [Azure Data Explorer](https://azure.microsoft.com/en-au/services/data-explorer/#security). The plugin assumes that there is a Database already exist in the Azure Data Explorer server. 

### Configuration:

```toml
[[outputs.azure_data_explorer]]
  ## The URI property of the Azure Data Explorer resource on Azure
  ## ex: https://myadxresource.australiasoutheast.kusto.windows.net
  # endpoint_url = ""

  ## The Azure Data Explorer database that the mertrics will be ingested into.
  ## The plugin will NOT generated this database automatically, it's expected that this database already exists before ingestion.
  ## ex: "exampledatabase"
  # database = ""

  ## Client ID of the Azure Active Directory App (Service Principal). This Service Principal should have permissions on the Azure Data Explorer
  ## to create Tables and ingest data into these tables
  ## ex: client_id = "dc871111-1222-4eee-bwww-111111111111"
  # client_id = ""

  ## The Client Secret of the Service Principal above
  # client_secret = ""

  ## The Azure Tenant ID this Service Principal above belongs to
  # tenant_id = ""

  ## The data format in which the metrics data will be when sent to Azure Data Explorer. This option is required and has to be value of 'json'.
  # data_format = "json"

```

### Metrics Grouping

The plugin will group the metrics by the metric name, and will send each group of metrics to an Azure Data Explorer table. If the table doesn't exist the plugin will create the table, if the table exists then the plugin will try to merge the Telegraf metric schema to the existing table. For more information about the merge process check the [`.create-merge` documentation](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/create-merge-table-command).

The table name will match the `name` property of the metric, this means that the name of the metric should comply with the Azure Data Explorer table naming constraints. 

### Tables' Schema

The schema of the table will match the structure of the Telegraf `Metric` object. The corresponding Azure Data Explorer command would be like the following:
```
.create-merge table ['table-name']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime)
```

The corresponding table maping would be like the following:
```
.create-or-alter table ['table-name'] ingestion json mapping 'table-name_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'
```

**Note**: Since the `Metric` object is a complex type, the only output format supported is JSON, so make sure to set the `data_format` configuration in `telegraf.conf` to `json`.

### Authentiation and Permissions

The plugin uses Service Principal credentials to authenticate to the Azure Data Explorer server. For guidance on how to create and register an App in Azure Active Directory check [this article](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app#register-an-application), and for more information on the Service Principals check [this article](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals).

The Service Principal should be assigned the `Database User` role on the Database level in Azure Data Explorer. This role will allow the plugin to create the required tables and ingest data into it.