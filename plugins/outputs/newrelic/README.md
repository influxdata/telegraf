# New Relic Output Plugin

This plugin writes to the [New Relic Insights API](https://docs.newrelic.com/docs/insights/inserting-events)
and requires an `apikey` which can be obtained [here](https://docs.newrelic.com/docs/insights/insights-data-sources/custom-data/send-custom-events-event-api#register)
for the account.

Metrics are published in a flat JSON notation, with "eventType" used as the primary key.