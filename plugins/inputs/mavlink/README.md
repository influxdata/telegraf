# MavLink Input Plugin

The `mavlink` plugin connects to a MavLink flight controller and translates all incoming messages into metrics.

The purpose of this plugin is to allow Telegraf to be used to ingest live flight metrics from unmanned systems (drones, planes, boats, etc.)

This plugin does not control the rate of messages received or send any Mavlink commands to the flight controller; in order to set message rates, another system will have to [set the message intervals.](https://mavlink.io/en/mavgen_python/howto_requestmessages.html)

Warning: This input plugin potentially generates a large amount of data! Use the Configuration to limit the set of messages or the rate, or use another telegraf plugin to filter the output.

## Connection Setup

## Configuration

### Example Configuration

## Metrics

Each supported Mavlink message translates to one metric group, and fields on the Mavlink message are converted to fields in telegraf.

## Example Output