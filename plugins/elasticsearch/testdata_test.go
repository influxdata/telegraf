package elasticsearch

const statsResponse = `
{
  "cluster_name": "es-testcluster",
  "nodes": {
    "SDFsfSDFsdfFSDSDfSFDSDF": {
      "timestamp": 1436365550135,
      "name": "test.host.com",
      "transport_address": "inet[/127.0.0.1:9300]",
      "host": "test",
      "ip": [
        "inet[/127.0.0.1:9300]",
        "NONE"
      ],
      "attributes": {
        "master": "true"
      },
      "indices": {
        "docs": {
          "count": 29652,
          "deleted": 5229
        },
        "store": {
          "size_in_bytes": 37715234,
          "throttle_time_in_millis": 215
        },
        "indexing": {
          "index_total": 84790,
          "index_time_in_millis": 29680,
          "index_current": 0,
          "delete_total": 13879,
          "delete_time_in_millis": 1139,
          "delete_current": 0,
          "noop_update_total": 0,
          "is_throttled": false,
          "throttle_time_in_millis": 0
        },
        "get": {
          "total": 1,
          "time_in_millis": 2,
          "exists_total": 0,
          "exists_time_in_millis": 0,
          "missing_total": 1,
          "missing_time_in_millis": 2,
          "current": 0
        },
        "search": {
          "open_contexts": 0,
          "query_total": 1452,
          "query_time_in_millis": 5695,
          "query_current": 0,
          "fetch_total": 414,
          "fetch_time_in_millis": 146,
          "fetch_current": 0
        },
        "merges": {
          "current": 0,
          "current_docs": 0,
          "current_size_in_bytes": 0,
          "total": 133,
          "total_time_in_millis": 21060,
          "total_docs": 203672,
          "total_size_in_bytes": 142900226
        },
        "refresh": {
          "total": 1076,
          "total_time_in_millis": 20078
        },
        "flush": {
          "total": 115,
          "total_time_in_millis": 2401
        },
        "warmer": {
          "current": 0,
          "total": 2319,
          "total_time_in_millis": 448
        },
        "filter_cache": {
          "memory_size_in_bytes": 7384,
          "evictions": 0
        },
        "id_cache": {
          "memory_size_in_bytes": 0
        },
        "fielddata": {
          "memory_size_in_bytes": 12996,
          "evictions": 0
        },
        "percolate": {
          "total": 0,
          "time_in_millis": 0,
          "current": 0,
          "memory_size_in_bytes": -1,
          "memory_size": "-1b",
          "queries": 0
        },
        "completion": {
          "size_in_bytes": 0
        },
        "segments": {
          "count": 134,
          "memory_in_bytes": 1285212,
          "index_writer_memory_in_bytes": 0,
          "index_writer_max_memory_in_bytes": 172368955,
          "version_map_memory_in_bytes": 611844,
          "fixed_bit_set_memory_in_bytes": 0
        },
        "translog": {
          "operations": 17702,
          "size_in_bytes": 17
        },
        "suggest": {
          "total": 0,
          "time_in_millis": 0,
          "current": 0
        },
        "query_cache": {
          "memory_size_in_bytes": 0,
          "evictions": 0,
          "hit_count": 0,
          "miss_count": 0
        },
        "recovery": {
          "current_as_source": 0,
          "current_as_target": 0,
          "throttle_time_in_millis": 0
        }
      },
      "os": {
        "timestamp": 1436460392944,
        "uptime_in_millis": 25092,
        "load_average": [
          0.01,
          0.04,
          0.05
        ],
        "cpu": {
          "sys": 0,
          "user": 0,
          "idle": 99,
          "usage": 0,
          "stolen": 0
        },
        "mem": {
          "free_in_bytes": 477761536,
          "used_in_bytes": 1621868544,
          "free_percent": 74,
          "used_percent": 25,
          "actual_free_in_bytes": 1565470720,
          "actual_used_in_bytes": 534159360
        },
        "swap": {
          "used_in_bytes": 0,
          "free_in_bytes": 487997440
        }
      },
      "process": {
        "timestamp": 1436460392945,
        "open_file_descriptors": 160,
        "cpu": {
          "percent": 2,
          "sys_in_millis": 1870,
          "user_in_millis": 13610,
          "total_in_millis": 15480
        },
        "mem": {
          "resident_in_bytes": 246382592,
          "share_in_bytes": 18747392,
          "total_virtual_in_bytes": 4747890688
        }
      },
      "jvm": {
        "timestamp": 1436460392945,
        "uptime_in_millis": 202245,
        "mem": {
          "heap_used_in_bytes": 52709568,
          "heap_used_percent": 5,
          "heap_committed_in_bytes": 259522560,
          "heap_max_in_bytes": 1038876672,
          "non_heap_used_in_bytes": 39634576,
          "non_heap_committed_in_bytes": 40841216,
          "pools": {
            "young": {
              "used_in_bytes": 32685760,
              "max_in_bytes": 279183360,
              "peak_used_in_bytes": 71630848,
              "peak_max_in_bytes": 279183360
            },
            "survivor": {
              "used_in_bytes": 8912880,
              "max_in_bytes": 34865152,
              "peak_used_in_bytes": 8912888,
              "peak_max_in_bytes": 34865152
            },
            "old": {
              "used_in_bytes": 11110928,
              "max_in_bytes": 724828160,
              "peak_used_in_bytes": 14354608,
              "peak_max_in_bytes": 724828160
            }
          }
        },
        "threads": {
          "count": 44,
          "peak_count": 45
        },
        "gc": {
          "collectors": {
            "young": {
              "collection_count": 2,
              "collection_time_in_millis": 98
            },
            "old": {
              "collection_count": 1,
              "collection_time_in_millis": 24
            }
          }
        },
        "buffer_pools": {
          "direct": {
            "count": 40,
            "used_in_bytes": 6304239,
            "total_capacity_in_bytes": 6304239
          },
          "mapped": {
            "count": 0,
            "used_in_bytes": 0,
            "total_capacity_in_bytes": 0
          }
        }
      }
    }
  }
}
`

var indicesExpected = map[string]float64{
	"indices_id_cache_memory_size_in_bytes":             0,
	"indices_completion_size_in_bytes":                  0,
	"indices_suggest_total":                             0,
	"indices_suggest_time_in_millis":                    0,
	"indices_suggest_current":                           0,
	"indices_query_cache_memory_size_in_bytes":          0,
	"indices_query_cache_evictions":                     0,
	"indices_query_cache_hit_count":                     0,
	"indices_query_cache_miss_count":                    0,
	"indices_store_size_in_bytes":                       37715234,
	"indices_store_throttle_time_in_millis":             215,
	"indices_merges_current_docs":                       0,
	"indices_merges_current_size_in_bytes":              0,
	"indices_merges_total":                              133,
	"indices_merges_total_time_in_millis":               21060,
	"indices_merges_total_docs":                         203672,
	"indices_merges_total_size_in_bytes":                142900226,
	"indices_merges_current":                            0,
	"indices_filter_cache_memory_size_in_bytes":         7384,
	"indices_filter_cache_evictions":                    0,
	"indices_indexing_index_total":                      84790,
	"indices_indexing_index_time_in_millis":             29680,
	"indices_indexing_index_current":                    0,
	"indices_indexing_noop_update_total":                0,
	"indices_indexing_throttle_time_in_millis":          0,
	"indices_indexing_delete_total":                     13879,
	"indices_indexing_delete_time_in_millis":            1139,
	"indices_indexing_delete_current":                   0,
	"indices_get_exists_time_in_millis":                 0,
	"indices_get_missing_total":                         1,
	"indices_get_missing_time_in_millis":                2,
	"indices_get_current":                               0,
	"indices_get_total":                                 1,
	"indices_get_time_in_millis":                        2,
	"indices_get_exists_total":                          0,
	"indices_refresh_total":                             1076,
	"indices_refresh_total_time_in_millis":              20078,
	"indices_percolate_current":                         0,
	"indices_percolate_memory_size_in_bytes":            -1,
	"indices_percolate_queries":                         0,
	"indices_percolate_total":                           0,
	"indices_percolate_time_in_millis":                  0,
	"indices_translog_operations":                       17702,
	"indices_translog_size_in_bytes":                    17,
	"indices_recovery_current_as_source":                0,
	"indices_recovery_current_as_target":                0,
	"indices_recovery_throttle_time_in_millis":          0,
	"indices_docs_count":                                29652,
	"indices_docs_deleted":                              5229,
	"indices_flush_total_time_in_millis":                2401,
	"indices_flush_total":                               115,
	"indices_fielddata_memory_size_in_bytes":            12996,
	"indices_fielddata_evictions":                       0,
	"indices_search_fetch_current":                      0,
	"indices_search_open_contexts":                      0,
	"indices_search_query_total":                        1452,
	"indices_search_query_time_in_millis":               5695,
	"indices_search_query_current":                      0,
	"indices_search_fetch_total":                        414,
	"indices_search_fetch_time_in_millis":               146,
	"indices_warmer_current":                            0,
	"indices_warmer_total":                              2319,
	"indices_warmer_total_time_in_millis":               448,
	"indices_segments_count":                            134,
	"indices_segments_memory_in_bytes":                  1285212,
	"indices_segments_index_writer_memory_in_bytes":     0,
	"indices_segments_index_writer_max_memory_in_bytes": 172368955,
	"indices_segments_version_map_memory_in_bytes":      611844,
	"indices_segments_fixed_bit_set_memory_in_bytes":    0,
}

var osExpected = map[string]float64{
	"os_swap_used_in_bytes":       0,
	"os_swap_free_in_bytes":       487997440,
	"os_timestamp":                1436460392944,
	"os_uptime_in_millis":         25092,
	"os_cpu_sys":                  0,
	"os_cpu_user":                 0,
	"os_cpu_idle":                 99,
	"os_cpu_usage":                0,
	"os_cpu_stolen":               0,
	"os_mem_free_percent":         74,
	"os_mem_used_percent":         25,
	"os_mem_actual_free_in_bytes": 1565470720,
	"os_mem_actual_used_in_bytes": 534159360,
	"os_mem_free_in_bytes":        477761536,
	"os_mem_used_in_bytes":        1621868544,
}

var processExpected = map[string]float64{
	"process_mem_resident_in_bytes":      246382592,
	"process_mem_share_in_bytes":         18747392,
	"process_mem_total_virtual_in_bytes": 4747890688,
	"process_timestamp":                  1436460392945,
	"process_open_file_descriptors":      160,
	"process_cpu_total_in_millis":        15480,
	"process_cpu_percent":                2,
	"process_cpu_sys_in_millis":          1870,
	"process_cpu_user_in_millis":         13610,
}

var jvmExpected = map[string]float64{
	"jvm_timestamp":                                     1436460392945,
	"jvm_uptime_in_millis":                              202245,
	"jvm_mem_non_heap_used_in_bytes":                    39634576,
	"jvm_mem_non_heap_committed_in_bytes":               40841216,
	"jvm_mem_pools_young_max_in_bytes":                  279183360,
	"jvm_mem_pools_young_peak_used_in_bytes":            71630848,
	"jvm_mem_pools_young_peak_max_in_bytes":             279183360,
	"jvm_mem_pools_young_used_in_bytes":                 32685760,
	"jvm_mem_pools_survivor_peak_used_in_bytes":         8912888,
	"jvm_mem_pools_survivor_peak_max_in_bytes":          34865152,
	"jvm_mem_pools_survivor_used_in_bytes":              8912880,
	"jvm_mem_pools_survivor_max_in_bytes":               34865152,
	"jvm_mem_pools_old_peak_max_in_bytes":               724828160,
	"jvm_mem_pools_old_used_in_bytes":                   11110928,
	"jvm_mem_pools_old_max_in_bytes":                    724828160,
	"jvm_mem_pools_old_peak_used_in_bytes":              14354608,
	"jvm_mem_heap_used_in_bytes":                        52709568,
	"jvm_mem_heap_used_percent":                         5,
	"jvm_mem_heap_committed_in_bytes":                   259522560,
	"jvm_mem_heap_max_in_bytes":                         1038876672,
	"jvm_threads_peak_count":                            45,
	"jvm_threads_count":                                 44,
	"jvm_gc_collectors_young_collection_count":          2,
	"jvm_gc_collectors_young_collection_time_in_millis": 98,
	"jvm_gc_collectors_old_collection_count":            1,
	"jvm_gc_collectors_old_collection_time_in_millis":   24,
	"jvm_buffer_pools_direct_count":                     40,
	"jvm_buffer_pools_direct_used_in_bytes":             6304239,
	"jvm_buffer_pools_direct_total_capacity_in_bytes":   6304239,
	"jvm_buffer_pools_mapped_count":                     0,
	"jvm_buffer_pools_mapped_used_in_bytes":             0,
	"jvm_buffer_pools_mapped_total_capacity_in_bytes":   0,
}
