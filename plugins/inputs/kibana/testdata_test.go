package kibana

const kibanaStatusResponse = `
{
	"name": "my-kibana",
	"uuid": "00000000-0000-0000-0000-000000000000",
	"version": {
		"number": "6.3.2",
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
			"id": "plugin:kibana@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.567Z"
		},
		{
			"id": "plugin:elasticsearch@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:04.920Z"
		},
		{
			"id": "plugin:xpack_main@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.393Z"
		},
		{
			"id": "plugin:searchprofiler@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.395Z"
		},
		{
			"id": "plugin:tilemap@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.396Z"
		},
		{
			"id": "plugin:watcher@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.397Z"
		},
		{
			"id": "plugin:license_management@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.668Z"
		},
		{
			"id": "plugin:index_management@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.399Z"
		},
		{
			"id": "plugin:timelion@6.3.2",
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
			"id": "plugin:monitoring@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.922Z"
		},
		{
			"id": "plugin:grokdebugger@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.400Z"
		},
		{
			"id": "plugin:dashboard_mode@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.928Z"
		},
		{
			"id": "plugin:logstash@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.401Z"
		},
		{
			"id": "plugin:apm@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.950Z"
		},
		{
			"id": "plugin:console@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.958Z"
		},
		{
			"id": "plugin:console_extensions@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.961Z"
		},
		{
			"id": "plugin:metrics@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-27T07:37:42.965Z"
		},
		{
			"id": "plugin:reporting@6.3.2",
			"state": "green",
			"icon": "success",
			"message": "Ready",
			"since": "2018-07-28T10:07:02.402Z"
		}]
	},
	"metrics": {
		"last_updated": "2018-08-21T11:24:25.823Z",
		"collection_interval_in_millis": 5000,
		"uptime_in_millis": 2173595336,
		"process": {
			"mem": {
				"heap_max_in_bytes": 149954560,
				"heap_used_in_bytes": 126274392
			}
		},
		"os": {
			"cpu": {
				"load_average": {
					"1m": 0.1806640625,
					"5m": 0.49658203125,
					"15m": 0.458984375
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
				"200": 2
			}
		},
		"concurrent_connections": 10
	}
}
`

var kibanaStatusExpected = map[string]interface{}{
	"status_code":            1,
	"heap_max_bytes":         int64(149954560),
	"heap_used_bytes":        int64(126274392),
	"uptime_ms":              int64(2173595336),
	"response_time_avg_ms":   float64(12.5),
	"response_time_max_ms":   int64(123),
	"concurrent_connections": int64(10),
	"requests_per_sec":       float64(0.4),
}
