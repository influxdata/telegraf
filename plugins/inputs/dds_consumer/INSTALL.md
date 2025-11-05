# DDS Plugin Installation Guide

This guide provides step-by-step instructions for installing and configuring the DDS Consumer input plugin for Telegraf.

## Prerequisites

1. **RTI Connext DDS**: You must have RTI Connext DDS installed on your system
   - Download from [RTI Website](https://www.rti.com/free-trial)
   - Follow RTI's installation instructions for your platform

2. **Go Environment**: Ensure Go is properly installed and configured
   - Go 1.18 or later is recommended

## Installation Steps

### 1. Set RTI Environment Variables

Before running Telegraf, ensure RTI environment is configured:

#### Linux/macOS:
```bash
# Add to your .bashrc or .profile
export NDDSHOME=/path/to/rti_connext_dds
export RTI_LICENSE_FILE=/path/to/rti_license.dat
export LD_LIBRARY_PATH=$NDDSHOME/lib/<arch>:$LD_LIBRARY_PATH
```

#### Windows:
```cmd
set NDDSHOME=C:\path\to\rti_connext_dds
set RTI_LICENSE_FILE=C:\path\to\rti_license.dat
set PATH=%NDDSHOME%\bin;<arch>;%PATH%
```

### 3. Build Telegraf with DDS Plugin

```bash
cd /path/to/telegraf
go build ./cmd/telegraf
```

### 4. Verify Installation

Test that the plugin is available:

```bash
./telegraf --input-list | grep dds_consumer
```

## Configuration

### 1. Create DDS XML Configuration

Create an XML file describing your DDS configuration (see `example_config.xml`):

```xml
<?xml version="1.0"?>
<dds>
    <types>
        <!-- Define your data types -->
    </types>
    <domain_participant_library name="MyParticipantLibrary">
        <!-- Define participants and readers -->
    </domain_participant_library>
</dds>
```

### 2. Configure Telegraf

Add the DDS consumer plugin to your Telegraf configuration:

```toml
[[inputs.dds_consumer]]
  config_path = "/path/to/your/dds_config.xml"
  participant_config = "MyParticipantLibrary::MyParticipant"
  reader_config = "MySubscriber::MyReader"
  tag_keys = ["field1", "field2"]
  name_override = "my_metrics"
  data_format = "json"
```

### 3. Test Configuration

```bash
./telegraf --config /path/to/your/telegraf.conf --test
```

## Troubleshooting

### Common Issues

1. **RTI License Error**: Ensure `RTI_LICENSE_FILE` points to a valid license
2. **Library Not Found**: Verify `LD_LIBRARY_PATH` includes RTI libraries
3. **XML Parse Error**: Validate XML syntax and participant/reader names
4. **Network Issues**: Check DDS domain ID and network connectivity

### Debug Commands

```bash
# Test DDS connectivity
rtiddsping -domainId 0

# Check environment
echo $NDDSHOME
echo $RTI_LICENSE_FILE

# Validate XML configuration
rtiddsgen -validate your_config.xml
```

## Platform-Specific Notes

### Linux
- May require additional network configuration for multicast
- Check firewall settings for DDS traffic

### Windows
- Ensure Visual C++ Redistributable is installed
- Check Windows Firewall for DDS traffic

### Docker
- Use host networking mode for DDS multicast
- Mount RTI installation and license file

## Support

For RTI Connext DDS specific issues:
- [RTI Community Forums](https://community.rti.com)
- [RTI Documentation](https://community.rti.com/documentation)

For Telegraf specific issues:
- [Telegraf GitHub Repository](https://github.com/influxdata/telegraf)
- [InfluxData Community](https://community.influxdata.com)