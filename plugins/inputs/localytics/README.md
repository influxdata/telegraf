# Localytics Input Plugin

Gather information about the configured apps from [Localytics](localytics).

> there are [limits](localytics_limits) to the API usage 

**Note:** Telegraf also contains the [webhook][] input which could be used as an
alternative method for collecting [Localytics](localytics) information.

### Configuration

```toml
  ## Localytics API access key.
  # access_token = ""

  ## Localytics API secret key.
  # secret_key = ""

  ## Timeout for HTTP requests.
  # http_timeout = "5s"
```

### Metrics

> There is the [Localytics API](localytics_api) which allows to fetch a variety of information about an app. This plugin specifically supports gathering the configured apps from [Localytics](localytics).

- app
  - tags:
    - name
    - id
  - fields:
    - sessions
    - closes
    - users
    - events

### Example Output

```
 localytics,host=my_machine,id=00000000000000000-000000-0000-0000-000-000000000,name=MyAwesomeApp closes=9i,sessions=26i,users=16i 1565942342000000000

```

[localytics]: https://localytics.com
[localytics_api]: https://docs.localytics.com/dev/query-api.html
[localytics_limits]: https://docs.localytics.com/dev/query-api.html#query-api-overview-limits
[webhook]: /plugins/inputs/webhooks/github
