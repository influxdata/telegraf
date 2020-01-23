package kibana

const kibanaStatusResponse6_5 = `
{
	"name": "my-kibana",
	"uuid": "00000000-0000-0000-0000-000000000000",
	"version": {
		"number": "6.5.4",
		"build_hash": "53d0c6758ac3fb38a3a1df198c1d4c87765e63f7",
		"build_number": 17307,
		"build_snapshot": false
	},
	"status": {
		"overall": {
			"state": "green",
			"title": "Green",
			"nickname": "Looking good",
			"icon": "success",
			"since": "2018-07-27T07:37:42.567Z"
		},
		"statuses": [{
			"id": "plugin:kibana@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.567Z"
		},
		{
			"id": "plugin:elasticsearch@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:04.920Z"
		},
		{
			"id": "plugin:xpack_main@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.393Z"
		},
		{
			"id": "plugin:searchprofiler@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.395Z"
		},
		{
			"id": "plugin:tilemap@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.396Z"
		},
		{
			"id": "plugin:watcher@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.397Z"
		},
		{
			"id": "plugin:license_management@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.668Z"
		},
		{
			"id": "plugin:index_management@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.399Z"
		},
		{
			"id": "plugin:timelion@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.912Z"
		},
		{
			"id": "plugin:logtrail@0.1.29",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.919Z"
		},
		{
			"id": "plugin:monitoring@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.922Z"
		},
		{
			"id": "plugin:grokdebugger@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.400Z"
		},
		{
			"id": "plugin:dashboard_mode@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.928Z"
		},
		{
			"id": "plugin:logstash@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.401Z"
		},
		{
			"id": "plugin:apm@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.950Z"
		},
		{
			"id": "plugin:console@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.958Z"
		},
		{
			"id": "plugin:console_extensions@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.961Z"
		},
		{
			"id": "plugin:metrics@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.965Z"
		},
		{
			"id": "plugin:reporting@6.5.4",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.402Z"
		}]
	},
	"metrics": {
		"last_updated": "2020-01-15T09:40:17.733Z",
		"collection_interval_in_millis": 5000,
		"process": {
			"memory": {
				"heap": {
          				"total_in_bytes": 149954560,
          				"used_in_bytes": 126274392,
          				"size_limit": 1501560832
        				},
        			"resident_set_size_in_bytes": 286650368
      				},
      			"event_loop_delay": 0.5314235687255859,
      			"pid": 6,
      			"uptime_in_millis": 2173595336
    		},
    		"os": {
      			"load": {
        			"1m": 2.66015625,
        			"5m": 2.8173828125,
        			"15m": 2.51025390625
      			},
      			"memory": {
        			"total_in_bytes": 404355756032,
        			"free_in_bytes": 294494244864,
        			"used_in_bytes": 109861511168
      			},
      			"uptime_in_millis": 8220745000,
      			"cgroup": {
        			"cpuacct": {
          				"control_group": "/",
          				"usage_nanos": 1086527218898
        			},
        			"cpu": {
          				"control_group": "/",
          				"cfs_period_micros": 100000,
          				"cfs_quota_micros": -1,
          				"stat": {
            					"number_of_elapsed_periods": 0,
            					"number_of_times_throttled": 0,
            					"time_throttled_nanos": 0
          					}
        				}
      				}
    			},
    		"response_times": {
      			"avg_in_millis": 12.5,
      			"max_in_millis": 123
    		},
    		"requests": {
      			"total": 2,
      			"disconnects": 0,
      			"status_codes": {
        			"200": 1,
        			"304": 1
      				}
    		},
    		"concurrent_connections": 10
  }
}
`

var kibanaStatusExpected6_5 = map[string]interface{}{
	"status_code":            1,
	"heap_total_bytes":       int64(149954560),
	"heap_max_bytes":         int64(149954560),
	"heap_used_bytes":        int64(126274392),
	"uptime_ms":              int64(2173595336),
	"response_time_avg_ms":   float64(12.5),
	"response_time_max_ms":   int64(123),
	"concurrent_connections": int64(10),
	"requests_per_sec":       float64(0.4),
}
