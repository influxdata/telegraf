package elasticsearch

const clusterHealthResponse = `
{
   "cluster_name": "elasticsearch_telegraf",
   "status": "green",
   "timed_out": false,
   "number_of_nodes": 3,
   "number_of_data_nodes": 3,
   "active_primary_shards": 5,
   "active_shards": 15,
   "relocating_shards": 0,
   "initializing_shards": 0,
   "unassigned_shards": 0,
   "number_of_pending_tasks": 0,
   "task_max_waiting_in_queue_millis": 0,
   "active_shards_percent_as_number": 100.0
}
`

const clusterHealthResponseWithIndices = `
{
   "cluster_name": "elasticsearch_telegraf",
   "status": "green",
   "timed_out": false,
   "number_of_nodes": 3,
   "number_of_data_nodes": 3,
   "active_primary_shards": 5,
   "active_shards": 15,
   "relocating_shards": 0,
   "initializing_shards": 0,
   "unassigned_shards": 0,
   "number_of_pending_tasks": 0,
   "task_max_waiting_in_queue_millis": 0,
   "active_shards_percent_as_number": 100.0,
   "indices": {
      "v1": {
         "status": "green",
         "number_of_shards": 10,
         "number_of_replicas": 1,
         "active_primary_shards": 10,
         "active_shards": 20,
         "relocating_shards": 0,
         "initializing_shards": 0,
         "unassigned_shards": 0
      },
      "v2": {
         "status": "red",
         "number_of_shards": 10,
         "number_of_replicas": 1,
         "active_primary_shards": 0,
         "active_shards": 0,
         "relocating_shards": 0,
         "initializing_shards": 0,
         "unassigned_shards": 20
      }
   }
}
`

var clusterHealthExpected = map[string]interface{}{
	"status":                           "green",
	"status_code":                      1,
	"timed_out":                        false,
	"number_of_nodes":                  3,
	"number_of_data_nodes":             3,
	"active_primary_shards":            5,
	"active_shards":                    15,
	"relocating_shards":                0,
	"initializing_shards":              0,
	"unassigned_shards":                0,
	"number_of_pending_tasks":          0,
	"task_max_waiting_in_queue_millis": 0,
	"active_shards_percent_as_number":  100.0,
}

var v1IndexExpected = map[string]interface{}{
	"status":                "green",
	"status_code":           1,
	"number_of_shards":      10,
	"number_of_replicas":    1,
	"active_primary_shards": 10,
	"active_shards":         20,
	"relocating_shards":     0,
	"initializing_shards":   0,
	"unassigned_shards":     0,
}

var v2IndexExpected = map[string]interface{}{
	"status":                "red",
	"status_code":           3,
	"number_of_shards":      10,
	"number_of_replicas":    1,
	"active_primary_shards": 0,
	"active_shards":         0,
	"relocating_shards":     0,
	"initializing_shards":   0,
	"unassigned_shards":     20,
}

const nodeStatsResponse = `
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
        "load_average": [
          0.01,
          0.04,
          0.05
        ],
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
      },
      "thread_pool": {
        "percolate": {
          "threads": 123,
          "queue": 23,
          "active": 13,
          "rejected": 235,
          "largest": 23,
          "completed": 33
        },
        "fetch_shard_started": {
          "threads": 3,
          "queue": 1,
          "active": 5,
          "rejected": 6,
          "largest": 4,
          "completed": 54
        },
        "listener": {
          "threads": 1,
          "queue": 2,
          "active": 4,
          "rejected": 8,
          "largest": 1,
          "completed": 1
        },
        "index": {
          "threads": 6,
          "queue": 8,
          "active": 4,
          "rejected": 2,
          "largest": 3,
          "completed": 6
        },
        "refresh": {
          "threads": 23,
          "queue": 7,
          "active": 3,
          "rejected": 4,
          "largest": 8,
          "completed": 3
        },
        "suggest": {
          "threads": 2,
          "queue": 7,
          "active": 2,
          "rejected": 1,
          "largest": 8,
          "completed": 3
        },
        "generic": {
          "threads": 1,
          "queue": 4,
          "active": 6,
          "rejected": 3,
          "largest": 2,
          "completed": 27
        },
        "warmer": {
          "threads": 2,
          "queue": 7,
          "active": 3,
          "rejected": 2,
          "largest": 3,
          "completed": 1
        },
        "search": {
          "threads": 5,
          "queue": 7,
          "active": 2,
          "rejected": 7,
          "largest": 2,
          "completed": 4
        },
        "flush": {
          "threads": 3,
          "queue": 8,
          "active": 0,
          "rejected": 1,
          "largest": 5,
          "completed": 3
        },
        "optimize": {
          "threads": 3,
          "queue": 4,
          "active": 1,
          "rejected": 2,
          "largest": 7,
          "completed": 3
        },
        "fetch_shard_store": {
          "threads": 1,
          "queue": 7,
          "active": 4,
          "rejected": 2,
          "largest": 4,
          "completed": 1
        },
        "management": {
          "threads": 2,
          "queue": 3,
          "active": 1,
          "rejected": 6,
          "largest": 2,
          "completed": 22
        },
        "get": {
          "threads": 1,
          "queue": 8,
          "active": 4,
          "rejected": 3,
          "largest": 2,
          "completed": 1
        },
        "merge": {
          "threads": 6,
          "queue": 4,
          "active": 5,
          "rejected": 2,
          "largest": 5,
          "completed": 1
        },
        "bulk": {
          "threads": 4,
          "queue": 5,
          "active": 7,
          "rejected": 3,
          "largest": 1,
          "completed": 4
        },
        "snapshot": {
          "threads": 8,
          "queue": 5,
          "active": 6,
          "rejected": 2,
          "largest": 1,
          "completed": 0
        }
      },
      "fs": {
        "timestamp": 1436460392946,
        "total": {
          "total_in_bytes": 19507089408,
          "free_in_bytes": 16909316096,
          "available_in_bytes": 15894814720
        },
        "data": [
          {
            "path": "/usr/share/elasticsearch/data/elasticsearch/nodes/0",
            "mount": "/usr/share/elasticsearch/data",
            "type": "ext4",
            "total_in_bytes": 19507089408,
            "free_in_bytes": 16909316096,
            "available_in_bytes": 15894814720
          }
        ]
      },
      "transport": {
        "server_open": 13,
        "rx_count": 6,
        "rx_size_in_bytes": 1380,
        "tx_count": 6,
        "tx_size_in_bytes": 1380
      },
      "http": {
        "current_open": 3,
        "total_opened": 3
      },
      "breakers": {
        "fielddata": {
          "limit_size_in_bytes": 623326003,
          "limit_size": "594.4mb",
          "estimated_size_in_bytes": 0,
          "estimated_size": "0b",
          "overhead": 1.03,
          "tripped": 0
        },
        "request": {
          "limit_size_in_bytes": 415550668,
          "limit_size": "396.2mb",
          "estimated_size_in_bytes": 0,
          "estimated_size": "0b",
          "overhead": 1.0,
          "tripped": 0
        },
        "parent": {
          "limit_size_in_bytes": 727213670,
          "limit_size": "693.5mb",
          "estimated_size_in_bytes": 0,
          "estimated_size": "0b",
          "overhead": 1.0,
          "tripped": 0
        }
      }
    }
  }
}
`

const nodeStatsResponseJVMProcess = `
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

var nodestatsIndicesExpected = map[string]interface{}{
	"id_cache_memory_size_in_bytes":             float64(0),
	"completion_size_in_bytes":                  float64(0),
	"suggest_total":                             float64(0),
	"suggest_time_in_millis":                    float64(0),
	"suggest_current":                           float64(0),
	"query_cache_memory_size_in_bytes":          float64(0),
	"query_cache_evictions":                     float64(0),
	"query_cache_hit_count":                     float64(0),
	"query_cache_miss_count":                    float64(0),
	"store_size_in_bytes":                       float64(37715234),
	"store_throttle_time_in_millis":             float64(215),
	"merges_current_docs":                       float64(0),
	"merges_current_size_in_bytes":              float64(0),
	"merges_total":                              float64(133),
	"merges_total_time_in_millis":               float64(21060),
	"merges_total_docs":                         float64(203672),
	"merges_total_size_in_bytes":                float64(142900226),
	"merges_current":                            float64(0),
	"filter_cache_memory_size_in_bytes":         float64(7384),
	"filter_cache_evictions":                    float64(0),
	"indexing_index_total":                      float64(84790),
	"indexing_index_time_in_millis":             float64(29680),
	"indexing_index_current":                    float64(0),
	"indexing_noop_update_total":                float64(0),
	"indexing_throttle_time_in_millis":          float64(0),
	"indexing_delete_total":                     float64(13879),
	"indexing_delete_time_in_millis":            float64(1139),
	"indexing_delete_current":                   float64(0),
	"get_exists_time_in_millis":                 float64(0),
	"get_missing_total":                         float64(1),
	"get_missing_time_in_millis":                float64(2),
	"get_current":                               float64(0),
	"get_total":                                 float64(1),
	"get_time_in_millis":                        float64(2),
	"get_exists_total":                          float64(0),
	"refresh_total":                             float64(1076),
	"refresh_total_time_in_millis":              float64(20078),
	"percolate_current":                         float64(0),
	"percolate_memory_size_in_bytes":            float64(-1),
	"percolate_queries":                         float64(0),
	"percolate_total":                           float64(0),
	"percolate_time_in_millis":                  float64(0),
	"translog_operations":                       float64(17702),
	"translog_size_in_bytes":                    float64(17),
	"recovery_current_as_source":                float64(0),
	"recovery_current_as_target":                float64(0),
	"recovery_throttle_time_in_millis":          float64(0),
	"docs_count":                                float64(29652),
	"docs_deleted":                              float64(5229),
	"flush_total_time_in_millis":                float64(2401),
	"flush_total":                               float64(115),
	"fielddata_memory_size_in_bytes":            float64(12996),
	"fielddata_evictions":                       float64(0),
	"search_fetch_current":                      float64(0),
	"search_open_contexts":                      float64(0),
	"search_query_total":                        float64(1452),
	"search_query_time_in_millis":               float64(5695),
	"search_query_current":                      float64(0),
	"search_fetch_total":                        float64(414),
	"search_fetch_time_in_millis":               float64(146),
	"warmer_current":                            float64(0),
	"warmer_total":                              float64(2319),
	"warmer_total_time_in_millis":               float64(448),
	"segments_count":                            float64(134),
	"segments_memory_in_bytes":                  float64(1285212),
	"segments_index_writer_memory_in_bytes":     float64(0),
	"segments_index_writer_max_memory_in_bytes": float64(172368955),
	"segments_version_map_memory_in_bytes":      float64(611844),
	"segments_fixed_bit_set_memory_in_bytes":    float64(0),
}

var nodestatsOsExpected = map[string]interface{}{
	"load_average_0":           float64(0.01),
	"load_average_1":           float64(0.04),
	"load_average_2":           float64(0.05),
	"swap_used_in_bytes":       float64(0),
	"swap_free_in_bytes":       float64(487997440),
	"timestamp":                float64(1436460392944),
	"mem_free_percent":         float64(74),
	"mem_used_percent":         float64(25),
	"mem_actual_free_in_bytes": float64(1565470720),
	"mem_actual_used_in_bytes": float64(534159360),
	"mem_free_in_bytes":        float64(477761536),
	"mem_used_in_bytes":        float64(1621868544),
}

var nodestatsProcessExpected = map[string]interface{}{
	"mem_total_virtual_in_bytes": float64(4747890688),
	"timestamp":                  float64(1436460392945),
	"open_file_descriptors":      float64(160),
	"cpu_total_in_millis":        float64(15480),
	"cpu_percent":                float64(2),
	"cpu_sys_in_millis":          float64(1870),
	"cpu_user_in_millis":         float64(13610),
}

var nodestatsJvmExpected = map[string]interface{}{
	"timestamp":                                     float64(1436460392945),
	"uptime_in_millis":                              float64(202245),
	"mem_non_heap_used_in_bytes":                    float64(39634576),
	"mem_non_heap_committed_in_bytes":               float64(40841216),
	"mem_pools_young_max_in_bytes":                  float64(279183360),
	"mem_pools_young_peak_used_in_bytes":            float64(71630848),
	"mem_pools_young_peak_max_in_bytes":             float64(279183360),
	"mem_pools_young_used_in_bytes":                 float64(32685760),
	"mem_pools_survivor_peak_used_in_bytes":         float64(8912888),
	"mem_pools_survivor_peak_max_in_bytes":          float64(34865152),
	"mem_pools_survivor_used_in_bytes":              float64(8912880),
	"mem_pools_survivor_max_in_bytes":               float64(34865152),
	"mem_pools_old_peak_max_in_bytes":               float64(724828160),
	"mem_pools_old_used_in_bytes":                   float64(11110928),
	"mem_pools_old_max_in_bytes":                    float64(724828160),
	"mem_pools_old_peak_used_in_bytes":              float64(14354608),
	"mem_heap_used_in_bytes":                        float64(52709568),
	"mem_heap_used_percent":                         float64(5),
	"mem_heap_committed_in_bytes":                   float64(259522560),
	"mem_heap_max_in_bytes":                         float64(1038876672),
	"threads_peak_count":                            float64(45),
	"threads_count":                                 float64(44),
	"gc_collectors_young_collection_count":          float64(2),
	"gc_collectors_young_collection_time_in_millis": float64(98),
	"gc_collectors_old_collection_count":            float64(1),
	"gc_collectors_old_collection_time_in_millis":   float64(24),
	"buffer_pools_direct_count":                     float64(40),
	"buffer_pools_direct_used_in_bytes":             float64(6304239),
	"buffer_pools_direct_total_capacity_in_bytes":   float64(6304239),
	"buffer_pools_mapped_count":                     float64(0),
	"buffer_pools_mapped_used_in_bytes":             float64(0),
	"buffer_pools_mapped_total_capacity_in_bytes":   float64(0),
}

var nodestatsThreadPoolExpected = map[string]interface{}{
	"merge_threads":                 float64(6),
	"merge_queue":                   float64(4),
	"merge_active":                  float64(5),
	"merge_rejected":                float64(2),
	"merge_largest":                 float64(5),
	"merge_completed":               float64(1),
	"bulk_threads":                  float64(4),
	"bulk_queue":                    float64(5),
	"bulk_active":                   float64(7),
	"bulk_rejected":                 float64(3),
	"bulk_largest":                  float64(1),
	"bulk_completed":                float64(4),
	"warmer_threads":                float64(2),
	"warmer_queue":                  float64(7),
	"warmer_active":                 float64(3),
	"warmer_rejected":               float64(2),
	"warmer_largest":                float64(3),
	"warmer_completed":              float64(1),
	"get_largest":                   float64(2),
	"get_completed":                 float64(1),
	"get_threads":                   float64(1),
	"get_queue":                     float64(8),
	"get_active":                    float64(4),
	"get_rejected":                  float64(3),
	"index_threads":                 float64(6),
	"index_queue":                   float64(8),
	"index_active":                  float64(4),
	"index_rejected":                float64(2),
	"index_largest":                 float64(3),
	"index_completed":               float64(6),
	"suggest_threads":               float64(2),
	"suggest_queue":                 float64(7),
	"suggest_active":                float64(2),
	"suggest_rejected":              float64(1),
	"suggest_largest":               float64(8),
	"suggest_completed":             float64(3),
	"fetch_shard_store_queue":       float64(7),
	"fetch_shard_store_active":      float64(4),
	"fetch_shard_store_rejected":    float64(2),
	"fetch_shard_store_largest":     float64(4),
	"fetch_shard_store_completed":   float64(1),
	"fetch_shard_store_threads":     float64(1),
	"management_threads":            float64(2),
	"management_queue":              float64(3),
	"management_active":             float64(1),
	"management_rejected":           float64(6),
	"management_largest":            float64(2),
	"management_completed":          float64(22),
	"percolate_queue":               float64(23),
	"percolate_active":              float64(13),
	"percolate_rejected":            float64(235),
	"percolate_largest":             float64(23),
	"percolate_completed":           float64(33),
	"percolate_threads":             float64(123),
	"listener_active":               float64(4),
	"listener_rejected":             float64(8),
	"listener_largest":              float64(1),
	"listener_completed":            float64(1),
	"listener_threads":              float64(1),
	"listener_queue":                float64(2),
	"search_rejected":               float64(7),
	"search_largest":                float64(2),
	"search_completed":              float64(4),
	"search_threads":                float64(5),
	"search_queue":                  float64(7),
	"search_active":                 float64(2),
	"fetch_shard_started_threads":   float64(3),
	"fetch_shard_started_queue":     float64(1),
	"fetch_shard_started_active":    float64(5),
	"fetch_shard_started_rejected":  float64(6),
	"fetch_shard_started_largest":   float64(4),
	"fetch_shard_started_completed": float64(54),
	"refresh_rejected":              float64(4),
	"refresh_largest":               float64(8),
	"refresh_completed":             float64(3),
	"refresh_threads":               float64(23),
	"refresh_queue":                 float64(7),
	"refresh_active":                float64(3),
	"optimize_threads":              float64(3),
	"optimize_queue":                float64(4),
	"optimize_active":               float64(1),
	"optimize_rejected":             float64(2),
	"optimize_largest":              float64(7),
	"optimize_completed":            float64(3),
	"snapshot_largest":              float64(1),
	"snapshot_completed":            float64(0),
	"snapshot_threads":              float64(8),
	"snapshot_queue":                float64(5),
	"snapshot_active":               float64(6),
	"snapshot_rejected":             float64(2),
	"generic_threads":               float64(1),
	"generic_queue":                 float64(4),
	"generic_active":                float64(6),
	"generic_rejected":              float64(3),
	"generic_largest":               float64(2),
	"generic_completed":             float64(27),
	"flush_threads":                 float64(3),
	"flush_queue":                   float64(8),
	"flush_active":                  float64(0),
	"flush_rejected":                float64(1),
	"flush_largest":                 float64(5),
	"flush_completed":               float64(3),
}

var nodestatsFsExpected = map[string]interface{}{
	"data_0_total_in_bytes":     float64(19507089408),
	"data_0_free_in_bytes":      float64(16909316096),
	"data_0_available_in_bytes": float64(15894814720),
	"timestamp":                 float64(1436460392946),
	"total_free_in_bytes":       float64(16909316096),
	"total_available_in_bytes":  float64(15894814720),
	"total_total_in_bytes":      float64(19507089408),
}

var nodestatsTransportExpected = map[string]interface{}{
	"server_open":      float64(13),
	"rx_count":         float64(6),
	"rx_size_in_bytes": float64(1380),
	"tx_count":         float64(6),
	"tx_size_in_bytes": float64(1380),
}

var nodestatsHttpExpected = map[string]interface{}{
	"current_open": float64(3),
	"total_opened": float64(3),
}

var nodestatsBreakersExpected = map[string]interface{}{
	"fielddata_estimated_size_in_bytes": float64(0),
	"fielddata_overhead":                float64(1.03),
	"fielddata_tripped":                 float64(0),
	"fielddata_limit_size_in_bytes":     float64(623326003),
	"request_estimated_size_in_bytes":   float64(0),
	"request_overhead":                  float64(1.0),
	"request_tripped":                   float64(0),
	"request_limit_size_in_bytes":       float64(415550668),
	"parent_overhead":                   float64(1.0),
	"parent_tripped":                    float64(0),
	"parent_limit_size_in_bytes":        float64(727213670),
	"parent_estimated_size_in_bytes":    float64(0),
}

const clusterStatsResponse = `
{
   "host":"ip-10-0-1-214",
   "log_type":"metrics",
   "timestamp":1475767451229,
   "log_level":"INFO",
   "node_name":"test.host.com",
   "cluster_name":"es-testcluster",
   "status":"red",
   "indices":{
      "count":1,
      "shards":{
         "total":4,
         "primaries":4,
         "replication":0.0,
         "index":{
            "shards":{
               "min":4,
               "max":4,
               "avg":4.0
            },
            "primaries":{
               "min":4,
               "max":4,
               "avg":4.0
            },
            "replication":{
               "min":0.0,
               "max":0.0,
               "avg":0.0
            }
         }
      },
      "docs":{
         "count":4,
         "deleted":0
      },
      "store":{
         "size_in_bytes":17084,
         "throttle_time_in_millis":0
      },
      "fielddata":{
         "memory_size_in_bytes":0,
         "evictions":0
      },
      "query_cache":{
         "memory_size_in_bytes":0,
         "total_count":0,
         "hit_count":0,
         "miss_count":0,
         "cache_size":0,
         "cache_count":0,
         "evictions":0
      },
      "completion":{
         "size_in_bytes":0
      },
      "segments":{
         "count":4,
         "memory_in_bytes":11828,
         "terms_memory_in_bytes":8932,
         "stored_fields_memory_in_bytes":1248,
         "term_vectors_memory_in_bytes":0,
         "norms_memory_in_bytes":1280,
         "doc_values_memory_in_bytes":368,
         "index_writer_memory_in_bytes":0,
         "index_writer_max_memory_in_bytes":2048000,
         "version_map_memory_in_bytes":0,
         "fixed_bit_set_memory_in_bytes":0
      },
      "percolate":{
         "total":0,
         "time_in_millis":0,
         "current":0,
         "memory_size_in_bytes":-1,
         "memory_size":"-1b",
         "queries":0
      }
   },
   "nodes":{
      "count":{
         "total":1,
         "master_only":0,
         "data_only":0,
         "master_data":1,
         "client":0
      },
      "versions":[
         {
         "version": "2.3.3"
         }
      ],
      "os":{
         "available_processors":1,
         "allocated_processors":1,
         "mem":{
            "total_in_bytes":593301504
         },
         "names":[
            {
               "name":"Linux",
               "count":1
            }
         ]
      },
      "process":{
         "cpu":{
            "percent":0
         },
         "open_file_descriptors":{
            "min":145,
            "max":145,
            "avg":145
         }
      },
      "jvm":{
         "max_uptime_in_millis":11580527,
         "versions":[
            {
               "version":"1.8.0_101",
               "vm_name":"OpenJDK 64-Bit Server VM",
               "vm_version":"25.101-b13",
               "vm_vendor":"Oracle Corporation",
               "count":1
            }
         ],
         "mem":{
            "heap_used_in_bytes":70550288,
            "heap_max_in_bytes":1065025536
         },
         "threads":30
      },
      "fs":{
         "total_in_bytes":8318783488,
         "free_in_bytes":6447439872,
         "available_in_bytes":6344785920
      },
      "plugins":[
         {
            "name":"cloud-aws",
            "version":"2.3.3",
            "description":"The Amazon Web Service (AWS) Cloud plugin allows to use AWS API for the unicast discovery mechanism and add S3 repositories.",
            "jvm":true,
            "classname":"org.elasticsearch.plugin.cloud.aws.CloudAwsPlugin",
            "isolated":true,
            "site":false
         },
         {
            "name":"kopf",
            "version":"2.0.1",
            "description":"kopf - simple web administration tool for Elasticsearch",
            "url":"/_plugin/kopf/",
            "jvm":false,
            "site":true
         },
         {
            "name":"tr-metrics",
            "version":"7bd5b4b",
            "description":"Logs cluster and node stats for performance monitoring.",
            "jvm":true,
            "classname":"com.trgr.elasticsearch.plugin.metrics.MetricsPlugin",
            "isolated":true,
            "site":false
         }
      ]
   }
}
`

var clusterstatsIndicesExpected = map[string]interface{}{
	"completion_size_in_bytes":                  float64(0),
	"count":                                     float64(1),
	"docs_count":                                float64(4),
	"docs_deleted":                              float64(0),
	"fielddata_evictions":                       float64(0),
	"fielddata_memory_size_in_bytes":            float64(0),
	"percolate_current":                         float64(0),
	"percolate_memory_size_in_bytes":            float64(-1),
	"percolate_queries":                         float64(0),
	"percolate_time_in_millis":                  float64(0),
	"percolate_total":                           float64(0),
	"percolate_memory_size":                     "-1b",
	"query_cache_cache_count":                   float64(0),
	"query_cache_cache_size":                    float64(0),
	"query_cache_evictions":                     float64(0),
	"query_cache_hit_count":                     float64(0),
	"query_cache_memory_size_in_bytes":          float64(0),
	"query_cache_miss_count":                    float64(0),
	"query_cache_total_count":                   float64(0),
	"segments_count":                            float64(4),
	"segments_doc_values_memory_in_bytes":       float64(368),
	"segments_fixed_bit_set_memory_in_bytes":    float64(0),
	"segments_index_writer_max_memory_in_bytes": float64(2.048e+06),
	"segments_index_writer_memory_in_bytes":     float64(0),
	"segments_memory_in_bytes":                  float64(11828),
	"segments_norms_memory_in_bytes":            float64(1280),
	"segments_stored_fields_memory_in_bytes":    float64(1248),
	"segments_term_vectors_memory_in_bytes":     float64(0),
	"segments_terms_memory_in_bytes":            float64(8932),
	"segments_version_map_memory_in_bytes":      float64(0),
	"shards_index_primaries_avg":                float64(4),
	"shards_index_primaries_max":                float64(4),
	"shards_index_primaries_min":                float64(4),
	"shards_index_replication_avg":              float64(0),
	"shards_index_replication_max":              float64(0),
	"shards_index_replication_min":              float64(0),
	"shards_index_shards_avg":                   float64(4),
	"shards_index_shards_max":                   float64(4),
	"shards_index_shards_min":                   float64(4),
	"shards_primaries":                          float64(4),
	"shards_replication":                        float64(0),
	"shards_total":                              float64(4),
	"store_size_in_bytes":                       float64(17084),
	"store_throttle_time_in_millis":             float64(0),
}

var clusterstatsNodesExpected = map[string]interface{}{
	"count_client":                      float64(0),
	"count_data_only":                   float64(0),
	"count_master_data":                 float64(1),
	"count_master_only":                 float64(0),
	"count_total":                       float64(1),
	"fs_available_in_bytes":             float64(6.34478592e+09),
	"fs_free_in_bytes":                  float64(6.447439872e+09),
	"fs_total_in_bytes":                 float64(8.318783488e+09),
	"jvm_max_uptime_in_millis":          float64(1.1580527e+07),
	"jvm_mem_heap_max_in_bytes":         float64(1.065025536e+09),
	"jvm_mem_heap_used_in_bytes":        float64(7.0550288e+07),
	"jvm_threads":                       float64(30),
	"jvm_versions_0_count":              float64(1),
	"jvm_versions_0_version":            "1.8.0_101",
	"jvm_versions_0_vm_name":            "OpenJDK 64-Bit Server VM",
	"jvm_versions_0_vm_vendor":          "Oracle Corporation",
	"jvm_versions_0_vm_version":         "25.101-b13",
	"os_allocated_processors":           float64(1),
	"os_available_processors":           float64(1),
	"os_mem_total_in_bytes":             float64(5.93301504e+08),
	"os_names_0_count":                  float64(1),
	"os_names_0_name":                   "Linux",
	"process_cpu_percent":               float64(0),
	"process_open_file_descriptors_avg": float64(145),
	"process_open_file_descriptors_max": float64(145),
	"process_open_file_descriptors_min": float64(145),
	"versions_0_version":                "2.3.3",
	"plugins_0_classname":               "org.elasticsearch.plugin.cloud.aws.CloudAwsPlugin",
	"plugins_0_description":             "The Amazon Web Service (AWS) Cloud plugin allows to use AWS API for the unicast discovery mechanism and add S3 repositories.",
	"plugins_0_isolated":                true,
	"plugins_0_jvm":                     true,
	"plugins_0_name":                    "cloud-aws",
	"plugins_0_site":                    false,
	"plugins_0_version":                 "2.3.3",
	"plugins_1_description":             "kopf - simple web administration tool for Elasticsearch",
	"plugins_1_jvm":                     false,
	"plugins_1_name":                    "kopf",
	"plugins_1_site":                    true,
	"plugins_1_url":                     "/_plugin/kopf/",
	"plugins_1_version":                 "2.0.1",
	"plugins_2_classname":               "com.trgr.elasticsearch.plugin.metrics.MetricsPlugin",
	"plugins_2_description":             "Logs cluster and node stats for performance monitoring.",
	"plugins_2_isolated":                true,
	"plugins_2_jvm":                     true,
	"plugins_2_name":                    "tr-metrics",
	"plugins_2_site":                    false,
	"plugins_2_version":                 "7bd5b4b",
}

const IsMasterResult = "SDFsfSDFsdfFSDSDfSFDSDF 10.206.124.66 10.206.124.66 test.host.com "

const IsNotMasterResult = "junk 10.206.124.66 10.206.124.66 test.junk.com "
