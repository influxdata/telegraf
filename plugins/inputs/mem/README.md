# Mem Input Plugin

The mem plugin collects system memory metrics.

For a more complete explanation of the difference between *used* and
*actual_used* RAM, see [Linux ate my ram](http://www.linuxatemyram.com/).

### Configuration:
```toml
# Read metrics about memory usage
[[inputs.mem]]
  # no configuration
```

### Metrics:

Available fields are dependent on platform.

- mem
  - fields:
    - active (integer)
    - available (integer)
    - buffered (integer)
    - cached (integer)
    - free (integer)
    - inactive (integer)
    - slab (integer)
    - total (integer)
    - used (integer)
    - available_percent (float)
    - used_percent (float)
    - wired (integer)
    - commit_limit (integer)
    - committed_as (integer)
    - dirty (integer)
    - high_free (integer)
    - high_total (integer)
    - huge_page_size (integer)
    - huge_pages_free (integer)
    - huge_pages_total (integer)
    - low_free (integer)
    - low_total (integer)
    - mapped (integer)
    - page_tables (integer)
    - shared (integer)
    - swap_cached (integer)
    - swap_free (integer)
    - swap_total (integer)
    - vmalloc_chunk (integer)
    - vmalloc_total (integer)
    - vmalloc_used (integer)
    - write_back (integer)
    - write_back_tmp (integer)

### Example Output:
```
mem active=11347566592i,available=18705133568i,available_percent=89.4288960571006,buffered=1976709120i,cached=13975572480i,commit_limit=14753067008i,committed_as=2872422400i,dirty=87461888i,free=1352400896i,high_free=0i,high_total=0i,huge_page_size=2097152i,huge_pages_free=0i,huge_pages_total=0i,inactive=6201593856i,low_free=0i,low_total=0i,mapped=310427648i,page_tables=14397440i,shared=200781824i,slab=1937526784i,swap_cached=0i,swap_free=4294963200i,swap_total=4294963200i,total=20916207616i,used=3611525120i,used_percent=17.26663449848977,vmalloc_chunk=0i,vmalloc_total=35184372087808i,vmalloc_used=0i,wired=0i,write_back=0i,write_back_tmp=0i 1536704085000000000
```
