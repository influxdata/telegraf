# Librato Output Plugin

This plugin writes to the new [Librato Metrics API](https://www.librato.com/docs/api/#create-a-measurement) (the one that allows tags, exists since Jan. 2017).
As such it sends along all telegraf tags to librato (unlike the legacy librato plugin).

It requires an `api_user` and `api_token` which can be obtained [here](https://metrics.librato.com/account/api_tokens)
for the account.

