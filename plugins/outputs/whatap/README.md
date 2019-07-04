# WhaTap Output Plugin

This plugin writes to the WhaTap
and requires an `license` , 'whatap.server.host' (server address)

If the point value being sent cannot be converted to a float64, the metric is skipped.

Metrics are grouped by converting any `_` characters to `.` in the Point Name.
