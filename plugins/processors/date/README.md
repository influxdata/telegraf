# Date Processor Plugin

Use the `date` processor to add the metric timestamp as a human readable tag.

A common use is to add a tag that can be used to group by month or year.

A few example usecases include:

1) consumption data for utilities on per month basis
2) bandwidth capacity per month
3) compare energy production or sales on a yearly or monthly basis

## Configuration

```toml
# Dates measurements, tags, and fields that pass through this filter.
[[processors.date]]
  ## New tag to create
  tag_key = "month"

  ## New field to create (cannot set both field_key and tag_key)
  # field_key = "month"

  ## Date format string, must be a representation of the Go "reference time"
  ## which is "Mon Jan 2 15:04:05 -0700 MST 2006".
  date_format = "Jan"

  ## If destination is a field, date format can also be one of
  ## "unix", "unix_ms", "unix_us", or "unix_ns", which will insert an integer field.
  # date_format = "unix"

  ## Offset duration added to the date string when writing the new tag.
  # date_offset = "0s"

  ## Timezone to use when creating the tag or field using a reference time
  ## string.  This can be set to one of "UTC", "Local", or to a location name
  ## in the IANA Time Zone database.
  ##   example: timezone = "America/Los_Angeles"
  # timezone = "UTC"
```

### timezone

On Windows, only the `Local` and `UTC` zones are available by default.  To use
other timezones, set the `ZONEINFO` environment variable to the location of
[`zoneinfo.zip`][zoneinfo]:

```text
set ZONEINFO=C:\zoneinfo.zip
```

## Example

```diff
- throughput lower=10i,upper=1000i,mean=500i 1560540094000000000
+ throughput,month=Jun lower=10i,upper=1000i,mean=500i 1560540094000000000
```

[zoneinfo]: https://github.com/golang/go/raw/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/lib/time/zoneinfo.zip
