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
    - sreclaimable (integer)
    - sunreclaim (integer)
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
mem active=9299595264i,available=16818249728i,available_percent=80.41654254645131,buffered=2383761408i,cached=13316689920i,commit_limit=14751920128i,committed_as=11781156864i,dirty=122880i,free=1877688320i,high_free=0i,high_total=0i,huge_page_size=2097152i,huge_pages_free=0i,huge_pages_total=0i,inactive=7549939712i,low_free=0i,low_total=0i,mapped=416763904i,page_tables=19787776i,shared=670679040i,slab=2081071104i,sreclaimable=1923395584i,sunreclaim=157675520i,swap_cached=1302528i,swap_free=4286128128i,swap_total=4294963200i,total=20913917952i,used=3335778304i,used_percent=15.95004011996231,vmalloc_chunk=0i,vmalloc_total=35184372087808i,vmalloc_used=0i,wired=0i,write_back=0i,write_back_tmp=0i 1574712869000000000
```
