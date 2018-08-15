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

- mem
  - fields:
  	- active (int)
  	- available (int)
  	- buffered (int)
  	- cached (int)
  	- free (int)
  	- inactive (int)
  	- slab (int)
  	- total (int)
  	- used (int)
  	- available_percent (float)
  	- used_percent (float)
  	- wired (int)

### Example Output:
```
mem cached=7809495040i,inactive=6348988416i,total=20855394304i,available=11378946048i,buffered=927199232i,active=11292905472i,slab=1351340032i,used_percent=45.43883523785713,available_percent=54.56116476214287,used=9476448256i,free=1715331072i 1511894782000000000
```
