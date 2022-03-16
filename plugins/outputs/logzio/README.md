# Logz.io Output Plugin

This plugin sends metrics to Logz.io over HTTPs.

## Configuration

```toml
# A plugin that can send metrics over HTTPs to Logz.io
[[outputs.logzio]]
  ## Set to true if Logz.io sender checks the disk space before adding metrics to the disk queue.
  # check_disk_space = true

  ## The percent of used file system space at which the sender will stop queueing.
  ## When we will reach that percentage, the file system in which the queue is stored will drop
  ## all new logs until the percentage of used space drops below that threshold.
  # disk_threshold = 98

  ## How often Logz.io sender should drain the queue.
  ## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  # drain_duration = "3s"

  ## Where Logz.io sender should store the queue
  ## queue_dir = Sprintf("%s%s%s%s%d", os.TempDir(), string(os.PathSeparator),
  ##                     "logzio-buffer", string(os.PathSeparator), time.Now().UnixNano())

  ## Logz.io account token
  token = "your Logz.io token" # required

  ## Use your listener URL for your Logz.io account region.
  # url = "https://listener.logz.io:8071"
```

### Required parameters

* `token`: Your Logz.io token, which can be found under "settings" in your account.

### Optional parameters

* `check_disk_space`: Set to true if Logz.io sender checks the disk space before adding metrics to the disk queue.
* `disk_threshold`: If the queue_dir space crosses this threshold (in % of disk usage), the plugin will start dropping logs.
* `drain_duration`: Time to sleep between sending attempts.
* `queue_dir`: Metrics disk path. All the unsent metrics are saved to the disk in this location.
* `url`: Logz.io listener URL.
