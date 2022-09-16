# Timestream Output Plugin

The Timestream output plugin writes metrics to the [Amazon Timestream] service.

## Authentication

This plugin uses a credential chain for Authentication with Timestream
API endpoint. In the following order the plugin will attempt to authenticate.

1. Web identity provider credentials via STS if `role_arn` and `web_identity_token_file` are specified
1. [Assumed credentials via STS] if `role_arn` attribute is specified (source credentials are evaluated from subsequent rules). The `endpoint_url` attribute is used only for Timestream service. When fetching credentials, STS global endpoint will be used.
1. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
1. Shared profile from `profile` attribute
1. [Environment Variables]
1. [Shared Credentials]
1. [EC2 Instance Profile]

## Configuration

```toml @sample.conf
# Configuration for sending metrics to Amazon Timestream.
[[outputs.timestream]]
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order:
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

  ## Specifies how the data is organized in Timestream.
  ## Valid values are: single-table, multi-table.
  ## When mapping_mode is set to single-table, all of the data is stored in a single table.
  ## When mapping_mode is set to multi-table, the data is organized and stored in multiple tables.
  ## The default is multi-table.
  mapping_mode = "multi-table"

  ## Specifies if the plugin should create the table, if the table does not exist.
  create_table_if_not_exists = true

  ## Specifies the Timestream table magnetic store retention period in days.
  ## Check Timestream documentation for more details.
  ## NOTE: This property is valid when create_table_if_not_exists = true.
  create_table_magnetic_store_retention_period_in_days = 365

  ## Specifies the Timestream table memory store retention period in hours.
  ## Check Timestream documentation for more details.
  ## NOTE: This property is valid when create_table_if_not_exists = true.
  create_table_memory_store_retention_period_in_hours = 24

  ## Specifies how the data is written into Timestream.
  ## Valid values are: true, false
  ## When use_multi_measure_records is set to true, all of the tags and fields are stored
  ## as a single row in a Timestream table.
  ## When use_multi_measure_record is set to false, Timestream stores each field in a
  ## separate table row, thereby storing the tags multiple times (once for each field).
  ## The recommended setting is true.
  ## The default is false.
  use_multi_measure_records = "false"

  ## Specifies the measure_name to use when sending multi-measure records.
  ## NOTE: This property is valid when use_multi_measure_records=true and mapping_mode=multi-table
  measure_name_for_multi_measure_records = "telegraf_measure"

  ## Specifies the name of the table to write data into
  ## NOTE: This property is valid when mapping_mode=single-table.
  # single_table_name = ""

  ## Specifies the name of dimension when all of the data is being stored in a single table
  ## and the measurement name is transformed into the dimension value
  ## (see Mapping data from Influx to Timestream for details)
  ## NOTE: This property is valid when mapping_mode=single-table.
  # single_table_dimension_name_for_telegraf_measurement_name = "namespace"

  ## Only valid and optional if create_table_if_not_exists = true
  ## Specifies the Timestream table tags.
  ## Check Timestream documentation for more details
  # create_table_tags = { "foo" = "bar", "environment" = "dev"}

  ## Specify the maximum number of parallel go routines to ingest/write data
  ## If not specified, defaulted to 1 go routines
  max_write_go_routines = 25

  ## Please see README.md to know how line protocol data is mapped to Timestream
  ##
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

### Mapping data from Influx to Timestream

When writing data from Influx to Timestream,
data is written by default as follows:

```
 1. The timestamp is written as the time field.
 2. Tags are written as dimensions.
 3. Fields are written as measures.
 4. Measurements are written as table names.
 ```

 For example, consider the following data in line protocol format:
> weather,location=us-midwest,season=summer temperature=82,humidity=71 1465839830100400200
> airquality,location=us-west no2=5,pm25=16 1465839830100400200

where:    
  `weather` and `airquality` are the measurement names,
  `location` and `season` are tags,
  `temperature`, `humidity`, `no2`, `pm25` are fields.

When you choose to create a separate table for each measurement and store
multiple fields in a single table row, the data will be written into
Timestream as:
  1. The plugin will create 2 tables, namely, weather and airquality (mapping_mode=multi-table).
  2. The tables may contain multiple fields in a single table row (use_multi_measure_records=true).
  3. The table weather will contain the following columns and data:

  ```
    time | location | season | measure_name | temperature | humidity
    2016-06-13 17:43:50 | us-midwest | summer | <measure_name_for_multi_measure_records> | 82 | 71
  ```

  4. The table airquality will contain the following columns and data:

  ```
    time | location | measure_name | no2 | pm25
    2016-06-13 17:43:50 | us-west | <measure_name_for_multi_measure_records> | 5 | 16
  ```

  NOTE:
  `<measure_name_for_multi_measure_records>` represents the actual
  value of that property.


You can also choose to create a separate table per measurement and store
each field in a separate row per table. In that case:

  1. The plugin will create 2 tables, namely, weather and airquality (mapping_mode=multi-table).
  2. Each table row will contain a single field only (use_multi_measure_records=false).
  3. The table weather will contain the following columns and data:

  ```
    time | location | season | measure_name | measure_value::bigint
    2016-06-13 17:43:50 | us-midwest | summer | temperature | 82
    2016-06-13 17:43:50 | us-midwest | summer | humidity | 71
  ```

  4. The table airquality will contain the following columns and data:

  ```
    time | location | measure_name | measure_value::bigint
    2016-06-13 17:43:50 | us-west | no2 | 5
    2016-06-13 17:43:50 | us-west | pm25 | 16
  ```

You can also choose to store all the measurements in a single table and
store all fields in a single table row. In that case:

 1. This plugin will create a table with name <single_table_name> (mapping_mode=single-table).
 2. The table may contain multiple fields in a single table row (use_multi_measure_records=true).
 3. The table will contain the following column and data:

  ```
   time | location | season | <single_table_dimension_name_for_telegraf_measurement_name>
   | measure_name | temperature | humidity | no2 | pm25
   2016-06-13 17:43:50 | us-midwest | summer | weather | <measure_name_for_multi_measure_records>
   | 82 | 71 | null | null
   2016-06-13 17:43:50 | us-west | null | airquality | <measure_name_for_multi_measure_records>
   | null | null | 5 | 16
  ```

  NOTE:
  `<single_table_name>` represents the actual value of that property.
  `<single_table_dimension_name_for_telegraf_measurement_name>` represents
  the actual value of that property.
  `<measure_name_for_multi_measure_records>` represents the actual value of
  that property.


Furthermore, you can choose to store all the measurements in a single table
and store each field in a separate table row. In that case:

   1. Timestream will create a table with name <single_table_name> (mapping_mode=single-table).
   2. Each table row will contain a single field only (use_multi_measure_records=false).
   3. The table will contain the following column and data:

   ```
    time | location | season | namespace | measure_name | measure_value::bigint
    2016-06-13 17:43:50 | us-midwest | summer | weather | temperature | 82
    2016-06-13 17:43:50 | us-midwest | summer | weather | humidity | 71
    2016-06-13 17:43:50 | us-west | NULL | airquality | no2 | 5
    2016-06-13 17:43:50 | us-west | NULL | airquality | pm25 | 16
   ```
    NOTE:
    `<single_table_name>` represents the actual value of that property.
    `<single_table_dimension_name_for_telegraf_measurement_name>` represents the actual value
    of that property.
    `<measure_name_for_multi_measure_records>` represents the actual value of that property.


### References

```
[Amazon Timestream]: https://aws.amazon.com/timestream/
[Assumed credentials via STS]: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/stscreds
[Environment Variables]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables
[Shared Credentials]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file
[EC2 Instance Profile]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html
```
