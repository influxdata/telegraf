# Icecast Input Plugin

The icecast plugin gathers metrics from the Icecast listmount page enabling to see a detailed report of total listeners

### Configuration:

```toml
# Read listeners from an Icecast instance per mount
[[inputs.icecast]]
  ## Specify the IP adress to where the 'admin/listmounts' can be found. You can include port if needed.
  host = "localhost"

  ## The username/password combination needed to read the listmounts page.
  ## These must be equal to the admin login details specified in your Icecast configuration
  username = "admin"
  password = "hackme"

  ## If you wish your host name to be different then the one specified under host, you can change it here
  alias = ""

  ## Include the slash in mountpoint names or not
	slash = false
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
