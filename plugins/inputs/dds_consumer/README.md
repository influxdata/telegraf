# DDS Consumer Input Plugin

The DDS consumer plugin reads metrics over DDS by creating readers defined in [XML App Creation](https://community.rti.com/static/documentation/connext-dds/5.3.1/doc/manuals/connext_dds/xml_application_creation/RTI_ConnextDDS_CoreLibraries_XML_AppCreation_GettingStarted.pdf) configurations. This plugin converts received DDS data to JSON data and adds to a Telegraf output plugin.

## Configuration

```toml
[[inputs.dds_consumer]]
  ## XML configuration file path
  config_path = "example_configs/ShapeExample.xml"

  ## Configuration name for DDS Participant from a description in XML
  participant_config = "MyParticipantLibrary::Zero"

  ## Configuration name for DDS DataReader from a description in XML
  reader_config = "MySubscriber::MySquareReader"

  # Tag key is an array of keys that should be added as tags.
  tag_keys = ["color"]

  # Override the base name of the measurement
  name_override = "shapes"

  ## Data format to consume.
  data_format = "json"
```

## Prerequisites

- RTI Connext DDS Connector for Go must be installed on the system
- Valid DDS XML configuration file with defined participants, topics, and data readers
- Network connectivity to DDS domain

## How it Works

1. The plugin creates a DDS participant using the specified XML configuration
2. It establishes a DataReader using the configured reader settings
3. The plugin continuously waits for and processes incoming DDS samples
4. Each DDS sample is converted to JSON format
5. The JSON data is parsed into Telegraf metrics using the configured tag keys
6. Metrics are forwarded to configured Telegraf output plugins

## XML Configuration

The plugin requires an XML configuration file that defines:

- Domain participants
- Publishers and subscribers
- Topics and data types
- QoS settings
- DataReaders and DataWriters

Example XML structure:
```xml
<?xml version="1.0"?>
<dds>
    <types>
        <!-- Define your data types here -->
    </types>
    
    <domain_participant_library name="MyParticipantLibrary">
        <domain_participant name="Zero" domain_ref="MyDomainLibrary::MyDomain">
            <subscriber name="MySubscriber">
                <data_reader name="MySquareReader" topic_ref="Square"/>
            </subscriber>
        </domain_participant>
    </domain_participant_library>
</dds>
```

## Tag Processing

The `tag_keys` configuration allows you to specify which fields from the DDS data should be treated as tags rather than fields in the resulting metrics. This is useful for:

- Categorizing data by source, type, or other identifiers
- Creating time series with different tag combinations
- Filtering and grouping data in downstream processing

## Example Usage

Example configuration files are provided in this directory:
- `example_config.xml` - Sample DDS XML configuration
- `example_telegraf.conf` - Sample Telegraf configuration using the DDS plugin

### Generate config with DDS input & InfluxDB output plugins:

```bash
./telegraf --input-filter dds_consumer --output-filter influxdb config
```

### Generate a config file with DDS input & file output plugins:
```bash
./telegraf --input-filter dds_consumer --output-filter file config > dds_to_file.conf
```

### Run Telegraf with DDS consumer:
```bash
./telegraf --config ./dds_to_file.conf
```

When running with the `dds_consumer` plugin, ensure that the XML file for DDS configurations is located at the `config_path` specified in your Telegraf TOML config.

## Testing

You can test the plugin with the [RTI Shapes Demo](https://www.rti.com/free-trial/shapes-demo) by:

1. Publishing data with "Square" topic using Shapes Demo
2. Configuring the plugin to read from the "Square" topic
3. Observing metrics in your configured output

## Troubleshooting

### Common Issues

1. **DDS Domain Connectivity**: Ensure all participants are on the same DDS domain
2. **XML Configuration**: Verify XML file path and participant/reader names
3. **Topic Matching**: Confirm topic names and data types match between publishers and subscribers
4. **QoS Compatibility**: Check that QoS settings are compatible between writers and readers

### Debug Tips

- Enable debug logging in Telegraf to see detailed plugin operation
- Use RTI tools like `rtiddsping` to verify DDS connectivity
- Validate XML configuration with RTI tools before using with Telegraf
- Check that the RTI Connext DDS environment is properly configured

## Metrics

The plugin generates metrics based on the DDS data received. The metric name can be:
- Automatically derived from the DDS topic name
- Overridden using the `name_override` configuration
- Dynamically set based on data content (if using `json_name_key`)

Fields are created from all numeric data in the DDS samples, while configured tag keys become metric tags for categorization and filtering.