# Amon Output Plugin

This plugin writes to [Amon](https://www.amon.cx)
and requires an `serverkey` and `amoninstance` URL which can be obtained [here](https://www.amon.cx/docs/monitoring/)
for the account.

If the point value being sent cannot be converted to a float64, the metric is skipped.

Metrics are grouped by converting any `_` characters to `.` in the Point Name.