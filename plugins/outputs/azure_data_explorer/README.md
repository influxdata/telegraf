# Azure Data Explorer

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