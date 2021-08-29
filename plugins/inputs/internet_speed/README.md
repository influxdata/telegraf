# Internet Speed Monitor

- [Internet Speed Monitor](#internet-speed-monitor)
  - [Description](#description)
  - [Configuration](#configuration)
  - [Metrics](#metrics)
  - [Example Output](#example-output)

## Description

The `Internet Speed Monitor` collects data about the internet speed on the system.

## Configuration

```toml
# Monitors internet speed in the network
[[inputs.internet_speed]]
## Sets if runs file download test
## Default: false
enable_file_download = false
```

## Metrics

It collects latency, download speed and upload speed


| Name           | filed name | type    | Unit |
| -------------- | ---------- | ------- | ---- |
| Download Speed | download   | float64 | Mbps |
| Upload Speed   | upload     | float64 | Mbps |
| Latency        | latency    | float64 | ms   |

## Example Output

```sh
time                download host           latency  upload
----                -------- -------------  -------  ------
1628871027000000000 263.53   Sanyam-Ubuntu  5.995    257.313
1628871047000000000 269.59   Sanyam-Ubuntu  3.353    262.974
```