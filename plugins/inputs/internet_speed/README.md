# Internet Speed Monitor

The `Internet Speed Monitor` collects data about the internet speed on the system.

## Configuration

```toml
# Monitors internet speed using speedtest.net service
[[inputs.internet_speed]]
  ## Sets if runs file download test
  # enable_file_download = false

  ## Caches the closest server location
  # cache = false
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
internet_speed,host=Sanyam-Ubuntu download=41.791,latency=28.518,upload=59.798 1631031183000000000
```
