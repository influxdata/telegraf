# Librato Output Plugin

This plugin writes to the new [Librato Metrics API](https://www.librato.com/docs/api/#create-a-measurement) (the one that allows tags, exists since Jan. 2017).
As such it sends along all telegraf tags to librato (unlike the legacy librato plugin).

To find out if your account is ready for the new API read [Libratos instructions](https://www.librato.com/docs/kb/faq/account_questions/tags_or_sources/).

It requires an `api_user` and `api_token` which can be obtained [here](https://metrics.librato.com/account/api_tokens)
for the account.

Optionally you can also set `prefix` to prefix metrics in Librato, to give them their own
namespace to distinguish them from other agents of environemnts.

## Configuration Example

```
[[outputs.librato_with_tags]]
  ## Librato API user
  api_user = "librato@whatever.com"
  ## Librato API token
  api_token = "sdfsdf123123sdfsdf123123"
  ## Debug
  # debug = false
  ## Connection timeout.
  # timeout = "5s"
  ## Metrics prefix, used for the metric/measurement name prefix
  prefix = "telegraf.production"
```
