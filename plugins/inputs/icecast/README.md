# Icecast Input Plugin

The icecast plugin gathers metrics from the Icecast listmount page enabling to see a detailed report of total listeners

### Configuration:

```toml
# Read listeners from an Icecast instance per mount
[[inputs.icecast]]
  ## Specify the IP adress to where the '/admin/listmounts' can be found. You can include port if needed.
  ## If you'd like to report under an alias, use ; (e.g. https://localhost;Server 1)
  ## You can use multiple hosts who use the same login credentials by dividing with , (e.g. "http://localhost","https://127.0.0.1")
  urls = ["http://localhost"]

  ## Timeout to the complete conection and reponse time in seconds. Default (5 seconds)
  # response_timeout = "25s"

  ## The username/password combination needed to read the listmounts page.
  ## These must be equal to the admin login details specified in your Icecast configuration
  username = "admin"
  password = "hackme"

  ## Include the slash in mountpoint names or not
  slash = false

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

### Measurements & Fields:

- listeners

### Tags:

- All measurements have the following tags:
    - host (can be either hostname/ip or an alias)
    - mount



### Sample Queries:

These are some useful queries (to generate dashboards or other) to run against data from this plugin:

```
SELECT last("listeners") FROM "icecast" WHERE "host" = "host" AND $timeFilter GROUP BY time($interval), "host" fill(null)
```

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter icecast -test
> icecast,host=Server1 E2,mount=Stream1.mp3 listeners=220i 1493979352000000000

```
