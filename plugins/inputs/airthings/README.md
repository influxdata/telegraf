# Airthings Input Plugin

The [Airthings](https://www.airthings.com) plugin collects data from Airthings devices via the Airthings API.
See the [Airthings API](https://developer.airthings.com/docs/api-getting-started/index.html) for details.

## Airthings for Consumers

A consumer can create one Airthings API-client, with Client Credentials Grant authorization.
The client can be used to fetch air quality data uploaded by the customer's own device.
The API is limited to 120 requests per hour

## Airthings for Business

**(Not implemented for now)**<BR/>
Airthings For Business Client is by default allowed 5000 requests per hour.
Users signed in through the same client all share that same quota.
The current rate limit status is indicated in the response headers:

#### API response headers (Airthings for Business)
```shell
X-RateLimit-Reset: 1607336100 // The time at which the current rate limit window resets (UTC epoch seconds).
X-RateLimit-Remaining: 1000 // The number of remaining requests in the current rate limit window.
X-RateLimit-Limit: 5000 // The maximum number of requests you're granted per hour.
X-RateLimit-Retry-After: 100 // A new request can be performed after this many seconds.
```

## Configuration

```toml
[[inputs.airthings]]
  ## URL is the address to get metrics from
  url = "https://ext-api.airthings.com/v1/"

  ## Show inactive devices true
  showInactive = true

  ## Timeout for HTTPS
  # timeout = "5s"

  ## Interval for the Consumers API (The API is limited to 120 requests per hour)
  ## One API call is made to get the list of devices, the two calls per device
  ## e.g. 3 devices will generate 1 + (3 * 2) = 7 calls per execution cycle.
  ## 120 / 7 = 17 max call / hour
  ## 60 min / 17 = 3,5 minutes pause between calls, 3,5 min = 210 sec interval
  ## 210 sec + safety margin = 225 sec
  interval = "225s"

  ## OAuth2 Client Credentials Grant
  client_id = "<INSERT CLIENT_ID HERE>"
  client_secret = "<INSERT CLIENT_SECRET HERE>"
  # token_url = "https://accounts-api.airthings.com/v1/token"
  # scopes = ["read:device:current_values"] 

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```
