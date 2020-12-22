# Telegraf Assistant

## Overview
The Telegraf Assistant is a tool that allows Telegraf users to modify plugin configurations on the fly. 
From a local machine running the Telegraf Agent, the Assistant communicates with a remote cloud server 
via a long-lived websocket connection.

## Setup
Set the `INFLUX_TOKEN` environment variable to your token obtained from the InfluxDB Dashboard.

You may wish to add the token to your `.bash_profile` or `.bashrc`.

## API
The following 7 operations are supported:
- Start Plugin
- Stop Plugin
- Update Plugin
- Get Plugin Config
- Get Running Plugins
- Get All Plugins
- Get Plugin Schema

## Persistent Config

## Testing
To run our unit tests on the Assistant, run `go test` in the `assistant/` directory.
