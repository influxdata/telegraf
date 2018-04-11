package icinga2

import (
        "fmt"
        "net/http"
        "net/http/httptest"
        "testing"

        "github.com/influxdata/telegraf/testutil"
        "github.com/stretchr/testify/require"

)

var icingaStatus = `
{
    "results": [
        {
            "name": "ApiListener",
            "perfdata": [
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_conn_endpoints",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_endpoints",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_http_clients",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 1.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_clients",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_relay_queue_item_rate",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.6333333333333333,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_relay_queue_items",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_sync_queue_item_rate",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_sync_queue_items",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_work_queue_count",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_work_queue_item_rate",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_json_rpc_work_queue_items",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "api_num_not_conn_endpoints",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                }
            ],
            "status": {
                "api": {
                    "conn_endpoints": [],
                    "http": {
                        "clients": 1.0
                    },
                    "identity": "lenlnydtstcs05",
                    "json_rpc": {
                        "clients": 0.0,
                        "relay_queue_item_rate": 0.6333333333333333,
                        "relay_queue_items": 0.0,
                        "sync_queue_item_rate": 0.0,
                        "sync_queue_items": 0.0,
                        "work_queue_count": 0.0,
                        "work_queue_item_rate": 0.0,
                        "work_queue_items": 0.0
                    },
                    "not_conn_endpoints": [],
                    "num_conn_endpoints": 0.0,
                    "num_endpoints": 0.0,
                    "num_not_conn_endpoints": 0.0,
                    "zones": {
                        "lenlnydtstcs05": {
                            "client_log_lag": 0.0,
                            "connected": true,
                            "endpoints": [
                                "lenlnydtstcs05"
                            ],
                            "parent_zone": ""
                        }
                    }
                }
            }
        },
        {
            "name": "CIB",
            "perfdata": [],
            "status": {
                "active_host_checks": 0.016666666666666666,
                "active_host_checks_15min": 15.0,
                "active_host_checks_1min": 1.0,
                "active_host_checks_5min": 5.0,
                "active_service_checks": 0.18333333333333332,
                "active_service_checks_15min": 163.0,
                "active_service_checks_1min": 11.0,
                "active_service_checks_5min": 54.0,
                "avg_execution_time": 1.6000514897433193,
                "avg_latency": 0.0008189678192138672,
                "max_execution_time": 10.00667691230774,
                "max_latency": 0.0016567707061767578,
                "min_execution_time": 0.00045609474182128906,
                "min_latency": 0.000164031982421875,
                "num_hosts_acknowledged": 0.0,
                "num_hosts_down": 0.0,
                "num_hosts_flapping": 0.0,
                "num_hosts_in_downtime": 0.0,
                "num_hosts_pending": 0.0,
                "num_hosts_unreachable": 0.0,
                "num_hosts_up": 1.0,
                "num_services_acknowledged": 0.0,
                "num_services_critical": 2.0,
                "num_services_flapping": 0.0,
                "num_services_in_downtime": 0.0,
                "num_services_ok": 7.0,
                "num_services_pending": 0.0,
                "num_services_unknown": 0.0,
                "num_services_unreachable": 0.0,
                "num_services_warning": 2.0,
                "passive_host_checks": 0.0,
                "passive_host_checks_15min": 0.0,
                "passive_host_checks_1min": 0.0,
                "passive_host_checks_5min": 0.0,
                "passive_service_checks": 0.0,
                "passive_service_checks_15min": 0.0,
                "passive_service_checks_1min": 0.0,
                "passive_service_checks_5min": 0.0,
                "uptime": 1031466.8390378952
            }
        },
        {
            "name": "CheckerComponent",
            "perfdata": [
                {
                    "counter": false,
                    "crit": null,
                    "label": "checkercomponent_checker_idle",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 12.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "checkercomponent_checker_pending",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                }
            ],
            "status": {
                "checkercomponent": {
                    "checker": {
                        "idle": 12.0,
                        "pending": 0.0
                    }
                }
            }
        },
        {
            "name": "FileLogger",
            "perfdata": [],
            "status": {
                "filelogger": {
                    "main-log": 1.0
                }
            }
        },
        {
            "name": "IcingaApplication",
            "perfdata": [],
            "status": {
                "icingaapplication": {
                    "app": {
                        "enable_event_handlers": true,
                        "enable_flapping": true,
                        "enable_host_checks": true,
                        "enable_notifications": true,
                        "enable_perfdata": true,
                        "enable_service_checks": true,
                        "node_name": "lenlnydtstcs05",
                        "pid": 3299.0,
                        "program_start": 1518299112.86116,
                        "version": "r2.7.1-1"
                    }
                }
            }
        },
        {
            "name": "IdoMysqlConnection",
            "perfdata": [
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_queries_rate",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 3.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_queries_1min",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 180.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_queries_5mins",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 899.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_queries_15mins",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 2699.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_query_queue_items",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 0.0,
                    "warn": null
                },
                {
                    "counter": false,
                    "crit": null,
                    "label": "idomysqlconnection_ido-mysql_query_queue_item_rate",
                    "max": null,
                    "min": null,
                    "type": "PerfdataValue",
                    "unit": "",
                    "value": 3.1,
                    "warn": null
                }
            ],
            "status": {
                "idomysqlconnection": {
                    "ido-mysql": {
                        "connected": true,
                        "instance_name": "default",
                        "query_queue_item_rate": 3.1,
                        "query_queue_items": 0.0,
                        "version": "1.14.2"
                    }
                }
            }
        },
        {
            "name": "NotificationComponent",
            "perfdata": [],
            "status": {
                "notificationcomponent": {
                    "notification": 1.0
                }
            }
        },
        {
            "name": "SyslogLogger",
            "perfdata": [],
            "status": {
                "sysloglogger": {}
            }
        }
    ]
}
`

func TestIcinga2Status(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, icingaStatus)
    }))
    defer ts.Close()

     i := &Icinga2{
                URL: ts.URL,
        }

    var acc testutil.Accumulator
    err := acc.GatherError(i.Gather)
    require.NoError(t, err)


    fields := map[string]interface{}{
"ActiveHostChecks":      float64(0.016666666666666666),
"ActiveHostChecks15Min":         float64(15.0),
"ActiveHostChecks1Min":  float64(1.0),
"ActiveHostChecks5Min":  float64(5.0),
"ActiveServiceChecks":   float64(0.18333333333333332),
"ActiveServiceChecks15Min":      float64(163.0),
"ActiveServiceChecks1Min":       float64(11.0),
"ActiveServiceChecks5Min":       float64(54.0),
"AvgExecutionTime":      float64(1.6000514897433193),
"AvgLatency":    float64(0.0008189678192138672),
"MaxExecutionTime":      float64(10.00667691230774),
"MaxLatency":    float64(0.0016567707061767578),
"MinExecutionTime":      float64(0.00045609474182128906),
"MinLatency":    float64(0.000164031982421875),
"NumHostsAcknowledged":  float64(0.0),
"NumHostsDown":  float64(0.0),
"NumHostsFlapping":      float64(0.0),
"NumHostsInDowntime":    float64(0.0),
"NumHostsPending":       float64(0.0),
"NumHostsUnreachable":   float64(0.0),
"NumHostsUp":    float64(1.0),
"NumServicesAcknowledged":       float64(0.0),
"NumServicesCritical":   float64(2.0),
"NumServicesFlapping":   float64(0.0),
"NumServicesInDowntime":         float64(0.0),
"NumServicesOk":         float64(7.0),
"NumServicesPending":    float64(0.0),
"NumServicesUnknown":    float64(0.0),
"NumServicesUnreachable":        float64(0.0),
"NumServicesWarning":    float64(2.0),
"PassiveHostChecks":     float64(0.0),
"PassiveHostChecks15Min":        float64(0.0),
"PassiveHostChecks1Min":         float64(0.0),
"PassiveHostChecks5Min":         float64(0.0),
"PassiveServiceChecks":  float64(0.0),
"PassiveServiceChecks15Min":     float64(0.0),
"PassiveServiceChecks1Min":      float64(0.0),
"PassiveServiceChecks5Min":      float64(0.0),
"Uptime":        float64(1031466.8390378952),
    }
    acc.AssertContainsFields(t, "icinga2", fields)
}
