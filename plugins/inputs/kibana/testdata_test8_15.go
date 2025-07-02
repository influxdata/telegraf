package kibana

// Kibana 8.x test data with new "level" field
const kibanastatusresponse815 = `
{
	"name": "my-kibana",
	"uuid": "00000000-0000-0000-0000-000000000000",
	"version": {
		"number": "8.15.2",
		"build_hash": "bd7e2f0e02aa2ff6de58cfc9c70ac1aa6f84b2e7",
		"build_number": 76543,
		"build_snapshot": false
	},
	"status": {
		"overall": {
			"level": "available",
			"summary": "Kibana is operating normally"
		},
		"core": {
			"elasticsearch": {
				"level": "available",
				"summary": "Elasticsearch is available"
			},
			"savedObjects": {
				"level": "available", 
				"summary": "SavedObjects service has completed migrations and is available"
			}
		},
		"plugins": {
			"monitoring": {
				"level": "available",
				"summary": "Ready"
			},
			"security": {
				"level": "available",
				"summary": "Ready"
			}
		}
	},
	"metrics": {
		"last_updated": "2025-01-15T09:40:17.733Z",
		"collection_interval_in_millis": 5000,
		"process": {
			"memory": {
				"heap": {
					"total_in_bytes": 505769984,
					"used_in_bytes": 411445984,
					"size_limit": 2197815296
				},
				"resident_set_size_in_bytes": 625000000
			},
			"event_loop_delay": 1.2,
			"pid": 1,
			"uptime_in_millis": 1039853908
		},
		"os": {
			"load": {
				"1m": 1.5,
				"5m": 1.8,
				"15m": 2.1
			},
			"memory": {
				"total_in_bytes": 8589934592,
				"free_in_bytes": 4294967296,
				"used_in_bytes": 4294967296
			},
			"uptime_in_millis": 86400000
		},
		"response_times": {
			"avg_in_millis": 6,
			"max_in_millis": 11
		},
		"requests": {
			"total": 2,
			"disconnects": 0,
			"status_codes": {
				"200": 2
			}
		},
		"concurrent_connections": 1
	}
}
`

// Test data for Kibana 8.x with degraded status
const kibanastatusresponse815Degraded = `
{
	"name": "my-kibana",
	"uuid": "00000000-0000-0000-0000-000000000000",
	"version": {
		"number": "8.15.2",
		"build_hash": "bd7e2f0e02aa2ff6de58cfc9c70ac1aa6f84b2e7",
		"build_number": 76543,
		"build_snapshot": false
	},
	"status": {
		"overall": {
			"level": "degraded",
			"summary": "Kibana is degraded"
		}
	},
	"metrics": {
		"last_updated": "2025-01-15T09:40:17.733Z",
		"collection_interval_in_millis": 5000,
		"process": {
			"memory": {
				"heap": {
					"total_in_bytes": 505769984,
					"used_in_bytes": 411445984,
					"size_limit": 2197815296
				}
			},
			"uptime_in_millis": 1039853908
		},
		"response_times": {
			"avg_in_millis": 6,
			"max_in_millis": 11
		},
		"requests": {
			"total": 2
		},
		"concurrent_connections": 1
	}
}
`

// Test data for Kibana 8.x with unavailable status
const kibanastatusresponse815Unavailable = `
{
	"name": "my-kibana",
	"uuid": "00000000-0000-0000-0000-000000000000",
	"version": {
		"number": "8.15.2",
		"build_hash": "bd7e2f0e02aa2ff6de58cfc9c70ac1aa6f84b2e7",
		"build_number": 76543,
		"build_snapshot": false
	},
	"status": {
		"overall": {
			"level": "unavailable",
			"summary": "Kibana is unavailable"
		}
	},
	"metrics": {
		"last_updated": "2025-01-15T09:40:17.733Z",
		"collection_interval_in_millis": 5000,
		"process": {
			"memory": {
				"heap": {
					"total_in_bytes": 505769984,
					"used_in_bytes": 411445984,
					"size_limit": 2197815296
				}
			},
			"uptime_in_millis": 1039853908
		},
		"response_times": {
			"avg_in_millis": 6,
			"max_in_millis": 11
		},
		"requests": {
			"total": 2
		},
		"concurrent_connections": 1
	}
}
`

var kibanastatusexpected815 = map[string]interface{}{
	"status_code":            1, // available -> green -> 1
	"heap_total_bytes":       int64(505769984),
	"heap_max_bytes":         int64(505769984),
	"heap_used_bytes":        int64(411445984),
	"heap_size_limit":        int64(2197815296),
	"uptime_ms":              int64(1039853908),
	"response_time_avg_ms":   float64(6),
	"response_time_max_ms":   int64(11),
	"concurrent_connections": int64(1),
	"requests_per_sec":       float64(0.4),
}

var kibanastatusexpected815Degraded = map[string]interface{}{
	"status_code":            2, // degraded -> yellow -> 2
	"heap_total_bytes":       int64(505769984),
	"heap_max_bytes":         int64(505769984),
	"heap_used_bytes":        int64(411445984),
	"heap_size_limit":        int64(2197815296),
	"uptime_ms":              int64(1039853908),
	"response_time_avg_ms":   float64(6),
	"response_time_max_ms":   int64(11),
	"concurrent_connections": int64(1),
	"requests_per_sec":       float64(0.4),
}

var kibanastatusexpected815Unavailable = map[string]interface{}{
	"status_code":            3, // unavailable -> red -> 3
	"heap_total_bytes":       int64(505769984),
	"heap_max_bytes":         int64(505769984),
	"heap_used_bytes":        int64(411445984),
	"heap_size_limit":        int64(2197815296),
	"uptime_ms":              int64(1039853908),
	"response_time_avg_ms":   float64(6),
	"response_time_max_ms":   int64(11),
	"concurrent_connections": int64(1),
	"requests_per_sec":       float64(0.4),
}
