# Telegraf plugin: zfs

Get ZFS stat from /proc/spl/kstat/zfs

# Measurements

Meta:

- tags: `pools=POOL1::POOL2`

Measurement names:

- arcstats_hits
- arcstats_misses
- arcstats_demand_data_hits
- arcstats_demand_data_misses
- arcstats_demand_metadata_hits
- arcstats_demand_metadata_misses
- arcstats_prefetch_data_hits
- arcstats_prefetch_data_misses
- arcstats_prefetch_metadata_hits
- arcstats_prefetch_metadata_misses
- arcstats_mru_hits
- arcstats_mru_ghost_hits
- arcstats_mfu_hits
- arcstats_mfu_ghost_hits
- arcstats_deleted
- arcstats_recycle_miss
- arcstats_mutex_miss
- arcstats_evict_skip
- arcstats_evict_l2_cached
- arcstats_evict_l2_eligible
- arcstats_evict_l2_ineligible
- arcstats_hash_elements
- arcstats_hash_elements_max
- arcstats_hash_collisions
- arcstats_hash_chains
- arcstats_hash_chain_max
- arcstats_p
- arcstats_c
- arcstats_c_min
- arcstats_c_max
- arcstats_size
- arcstats_hdr_size
- arcstats_data_size
- arcstats_meta_size
- arcstats_other_size
- arcstats_anon_size
- arcstats_anon_evict_data
- arcstats_anon_evict_metadata
- arcstats_mru_size
- arcstats_mru_evict_data
- arcstats_mru_evict_metadata
- arcstats_mru_ghost_size
- arcstats_mru_ghost_evict_data
- arcstats_mru_ghost_evict_metadata
- arcstats_mfu_size
- arcstats_mfu_evict_data
- arcstats_mfu_evict_metadata
- arcstats_mfu_ghost_size
- arcstats_mfu_ghost_evict_data
- arcstats_mfu_ghost_evict_metadata
- arcstats_l2_hits
- arcstats_l2_misses
- arcstats_l2_feeds
- arcstats_l2_rw_clash
- arcstats_l2_read_bytes
- arcstats_l2_write_bytes
- arcstats_l2_writes_sent
- arcstats_l2_writes_done
- arcstats_l2_writes_error
- arcstats_l2_writes_hdr_miss
- arcstats_l2_evict_lock_retry
- arcstats_l2_evict_reading
- arcstats_l2_free_on_write
- arcstats_l2_cdata_free_on_write
- arcstats_l2_abort_lowmem
- arcstats_l2_cksum_bad
- arcstats_l2_io_error
- arcstats_l2_size
- arcstats_l2_asize
- arcstats_l2_hdr_size
- arcstats_l2_compress_successes
- arcstats_l2_compress_zeros
- arcstats_l2_compress_failures
- arcstats_memory_throttle_count
- arcstats_duplicate_buffers
- arcstats_duplicate_buffers_size
- arcstats_duplicate_reads
- arcstats_memory_direct_count
- arcstats_memory_indirect_count
- arcstats_arc_no_grow
- arcstats_arc_tempreserve
- arcstats_arc_loaned_bytes
- arcstats_arc_prune
- arcstats_arc_meta_used
- arcstats_arc_meta_limit
- arcstats_arc_meta_max
- zfetchstats_hits
- zfetchstats_misses
- zfetchstats_colinear_hits
- zfetchstats_colinear_misses
- zfetchstats_stride_hits
- zfetchstats_stride_misses
- zfetchstats_reclaim_successes
- zfetchstats_reclaim_failures
- zfetchstats_streams_resets
- zfetchstats_streams_noresets
- zfetchstats_bogus_streams
- vdev_cache_stats_delegations
- vdev_cache_stats_hits
- vdev_cache_stats_misses

### Description

```
arcstats_hits
  Total amount of cache hits in the arc.

arcstats_misses
  Total amount of cache misses in the arc.

arcstats_demand_data_hits
  Amount of cache hits for demand data, this is what matters (is good) for your application/share.

arcstats_demand_data_misses
  Amount of cache misses for demand data, this is what matters (is bad) for your application/share.

arcstats_demand_metadata_hits
  Ammount of cache hits for demand metadata, this matters (is good) for getting filesystem data (ls,find,…)

arcstats_demand_metadata_misses
  Ammount of cache misses for demand metadata, this matters (is bad) for getting filesystem data (ls,find,…)

arcstats_prefetch_data_hits
  The zfs prefetcher tried to prefetch somethin, but it was allready cached (boring)

arcstats_prefetch_data_misses
  The zfs prefetcher prefetched something which was not in the cache (good job, could become a demand hit in the future)

arcstats_prefetch_metadata_hits
  Same as above, but for metadata

arcstats_prefetch_metadata_misses
  Same as above, but for metadata

arcstats_mru_hits
  Cache hit in the “most recently used cache”, we move this to the mfu cache.

arcstats_mru_ghost_hits
  Cache hit in the “most recently used ghost list” we had this item in the cache, but evicted it, maybe we should increase the mru cache size.

arcstats_mfu_hits
  Cache hit in the “most freqently used cache” we move this to the begining of the mfu cache.

arcstats_mfu_ghost_hits
  Cache hit in the “most frequently used ghost list” we had this item in the cache, but evicted it, maybe we should increase the mfu cache size.

arcstats_allocated
  New data is written to the cache.

arcstats_deleted
  Old data is evicted (deleted) from the cache.

arcstats_evict_l2_cached
  We evicted something from the arc, but its still cached in the l2 if we need it.

arcstats_evict_l2_eligible
  We evicted something from the arc, and it’s not in the l2 this is sad. (maybe we hadn’t had enough time to store it there)

arcstats_evict_l2_ineligible
  We evicted something which cannot be stored in the l2.
  Reasons could be:
  We have multiple pools, we evicted something from a pool whithot an l2 device.
  The zfs property secondarycache.

arcstats_c
  Arc target size, this is the size the system thinks the arc should have.

arcstats_size
  Total size of the arc.

arcstats_l2_hits
  Hits to the L2 cache. (It was not in the arc, but in the l2 cache)

arcstats_l2_misses
  Miss to the L2 cache. (It was not in the arc, and not in the l2 cache)

arcstats_l2_size
  Size of the l2 cache.

arcstats_l2_hdr_size
  Size of the metadata in the arc (ram) used to manage (lookup if someting is in the l2) the l2 cache.



zfetchstats_hits
  Counts the number of cache hits, to items wich are in the cache because of the prefetcher.

zfetchstats_colinear_hits
  Counts the number of cache hits, to items wich are in the cache because of the prefetcher (prefetched linear reads)

zfetchstats_stride_hits
  Counts the number of cache hits, to items wich are in the cache because of the prefetcher (prefetched stride reads)



vdev_cache_stats_hits
  Hits to the vdev (device level) cache.

vdev_cache_stats_misses
  Misses to the vdev (device level) cache.
```

# Default config

```
[zfs]
  # ZFS kstat path
  # If not specified, then default is:
  # kstatPath = "/proc/spl/kstat/zfs"
  #
  # By default, telegraf gather all zfs stats
  # If not specified, then default is:
  # kstatMetrics = ["arcstats", "zfetchstats", "vdev_cache_stats"]
```

