# Twitter Input Plugin

Gather account information from [Twitter](https://twitter.com/) accounts.

### Configuration

```toml
[[inputs.twitter]]
  ## List of accounts to monitor.
  accounts = [
    783214,
    1967601206
  ]

  ## Twitter API consumer key.
  # consumer_key = ""
  ## Twitter API consumer secret.
  # consumer_secret = ""
  ## Twitter API access token.
  # access_token = ""
  ## Twitter API access token secret.
  # access_token_secret = ""
```

### Metrics

- twitter_account
  - tags:
    - id - The ID of the account
    - screen_name - The screen name
  - fields:
    - favourites (int)
    - followers (int)
    - friends (int)
    - statuses (int64)

### Example Output

```
twitter_account,id=1967601206,screen_name=InfluxDB followers=16688i,friends=4235i,statuses=11209i,favourites=7542i 1575008500000000000
twitter_account,id=783214,screen_name=Twitter favourites=6353i,followers=56766065i,friends=102i,statuses=12423i 1575008500000000000
```
