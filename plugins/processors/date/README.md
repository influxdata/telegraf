# Date Processor Plugin

Use the `date` processor to add the metric timestamp as a human readable tag.

A common use is to add a tag that can be used to group by month or year.

A few example usecases include:
1) consumption data for utilities on per month basis
2) bandwith capacity per month
3) compare energy production or sales on a yearly or monthly basis

### Configuration

```toml
[[processors.date]]
  ## New tag to create
  tag_key = "month"

  ## Date format string, must be a representation of the Go "reference time"
  ## which is "Mon Jan 2 15:04:05 -0700 MST 2006".
  date_format = "Jan"

  ## Offset duration added to the date string when writing the new tag.
  # date_offset = "0s"
```

### Example

```diff
- throughput lower=10i,upper=1000i,mean=500i 1560540094000000000
+ throughput,month=Jun lower=10i,upper=1000i,mean=500i 1560540094000000000
```
