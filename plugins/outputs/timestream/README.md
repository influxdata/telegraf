# Timestream Output Plugin

The Timestream output plugin writes metrics to the [Amazon Timestream] service.

## Configuration

```toml
# Configuration for sending metrics to Amazon Timestream.
[[outputs.timestream]]
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Timestream database where the metrics will be inserted.
  ## The database must exist prior to starting Telegraf.
  database_name = "yourDatabaseNameHere"

  ## Specifies if the plugin should describe the Timestream database upon starting
  ## to validate if it has access necessary permissions, connection, etc., as a safety check.
  ## If the describe operation fails, the plugin will not start
  ## and therefore the Telegraf agent will not start.
  describe_database_on_start = false

  ## The mapping mode specifies how Telegraf records are represented in Timestream.
  ## Valid values are: single-table, multi-table.
  ## For example, consider the following data in line protocol format:
  ## weather,location=us-midwest,season=summer temperature=82,humidity=71 1465839830100400200
  ## airquality,location=us-west no2=5,pm25=16 1465839830100400200
  ## where weather and airquality are the measurement names, location and season are tags,
  ## and temperature, humidity, no2, pm25 are fields.
  ## In multi-table mode:
  ##  - first line will be ingested to table named weather
  ##  - second line will be ingested to table named airquality
  ##  - the tags will be represented as dimensions
  ##  - first table (weather) will have two records:
  ##      one with measurement name equals to temperature,
  ##      another with measurement name equals to humidity
  ##  - second table (airquality) will have two records:
  ##      one with measurement name equals to no2,
  ##      another with measurement name equals to pm25
  ##  - the Timestream tables from the example will look like this:
  ##      TABLE "weather":
  ##        time | location | season | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-midwest | summer | temperature | 82
  ##        2016-06-13 17:43:50 | us-midwest | summer | humidity | 71
  ##      TABLE "airquality":
  ##        time | location | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-west | no2 | 5
  ##        2016-06-13 17:43:50 | us-west | pm25 | 16
  ## In single-table mode:
  ##  - the data will be ingested to a single table, which name will be valueOf(single_table_name)
  ##  - measurement name will stored in dimension named valueOf(single_table_dimension_name_for_telegraf_measurement_name)
  ##  - location and season will be represented as dimensions
  ##  - temperature, humidity, no2, pm25 will be represented as measurement name
  ##  - the Timestream table from the example will look like this:
  ##      Assuming:
  ##        - single_table_name = "my_readings"
  ##        - single_table_dimension_name_for_telegraf_measurement_name = "namespace"
  ##      TABLE "my_readings":
  ##        time | location | season | namespace | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-midwest | summer | weather | temperature | 82
  ##        2016-06-13 17:43:50 | us-midwest | summer | weather | humidity | 71
  ##        2016-06-13 17:43:50 | us-west | NULL | airquality | no2 | 5
  ##        2016-06-13 17:43:50 | us-west | NULL | airquality | pm25 | 16
  ## In most cases, using multi-table mapping mode is recommended.
  ## However, you can consider using single-table in situations when you have thousands of measurement names.
  mapping_mode = "multi-table"

  ## Only valid and required for mapping_mode = "single-table"
  ## Specifies the Timestream table where the metrics will be uploaded.
  # single_table_name = "yourTableNameHere"

  ## Only valid and required for mapping_mode = "single-table"
  ## Describes what will be the Timestream dimension name for the Telegraf
  ## measurement name.
  # single_table_dimension_name_for_telegraf_measurement_name = "namespace"

  ## Specifies if the plugin should create the table, if the table do not exist.
  ## The plugin writes the data without prior checking if the table exists.
  ## When the table does not exist, the error returned from Timestream will cause
  ## the plugin to create the table, if this parameter is set to true.
  create_table_if_not_exists = true

  ## Only valid and required if create_table_if_not_exists = true
  ## Specifies the Timestream table magnetic store retention period in days.
  ## Check Timestream documentation for more details.
  create_table_magnetic_store_retention_period_in_days = 365

  ## Only valid and required if create_table_if_not_exists = true
  ## Specifies the Timestream table memory store retention period in hours.
  ## Check Timestream documentation for more details.
  create_table_memory_store_retention_period_in_hours = 24

  ## Only valid and optional if create_table_if_not_exists = true
  ## Specifies the Timestream table tags.
  ## Check Timestream documentation for more details
  # create_table_tags = { "foo" = "bar", "environment" = "dev"}

  ## Specify the maximum number of parallel go routines to ingest/write data
  ## If not specified, defaulted to 1 go routines
  max_write_go_routines = 25
```

### Batching

Timestream WriteInputRequest.CommonAttributes are used to efficiently write data
to Timestream.

### Multithreading

Single thread is used to write the data to Timestream, following general plugin
design pattern.

### Errors

In case of an attempt to write an unsupported by Timestream Telegraf Field type,
the field is dropped and error is emitted to the logs.

In case of receiving ThrottlingException or InternalServerException from
Timestream, the errors are returned to Telegraf, in which case Telegraf will
keep the metrics in buffer and retry writing those metrics on the next flush.

In case of receiving ResourceNotFoundException:

- If `create_table_if_not_exists` configuration is set to `true`, the plugin
  will try to create appropriate table and write the records again, if the table
  creation was successful.
- If `create_table_if_not_exists` configuration is set to `false`, the records
  are dropped, and an error is emitted to the logs.

In case of receiving any other AWS error from Timestream, the records are
dropped, and an error is emitted to the logs, as retrying such requests isn't
likely to succeed.

### Logging

Turn on debug flag in the Telegraf to turn on detailed logging (including
records being written to Timestream).

### Testing

Execute unit tests with:

```shell
go test -v ./plugins/outputs/timestream/...
```

[Amazon Timestream]: https://aws.amazon.com/timestream/
