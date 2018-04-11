package icinga2

import (
    "encoding/json"
)

type CIB struct {
    Name     string        `json:"name"`
    Perfdata []interface{} `json:"perfdata"`
    Status   struct {
        ActiveHostChecks          float64 `json:"active_host_checks"`
        ActiveHostChecks15Min     float64 `json:"active_host_checks_15min"`
        ActiveHostChecks1Min      float64 `json:"active_host_checks_1min"`
        ActiveHostChecks5Min      float64 `json:"active_host_checks_5min"`
        ActiveServiceChecks       float64 `json:"active_service_checks"`
        ActiveServiceChecks15Min  float64 `json:"active_service_checks_15min"`
        ActiveServiceChecks1Min   float64 `json:"active_service_checks_1min"`
        ActiveServiceChecks5Min   float64 `json:"active_service_checks_5min"`
        AvgExecutionTime          float64 `json:"avg_execution_time"`
        AvgLatency                float64 `json:"avg_latency"`
        MaxExecutionTime          float64 `json:"max_execution_time"`
        MaxLatency                float64 `json:"max_latency"`
        MinExecutionTime          float64 `json:"min_execution_time"`
        MinLatency                float64 `json:"min_latency"`
        NumHostsAcknowledged      float64 `json:"num_hosts_acknowledged"`
        NumHostsDown              float64 `json:"num_hosts_down"`
        NumHostsFlapping          float64 `json:"num_hosts_flapping"`
        NumHostsInDowntime        float64 `json:"num_hosts_in_downtime"`
        NumHostsPending           float64 `json:"num_hosts_pending"`
        NumHostsUnreachable       float64 `json:"num_hosts_unreachable"`
        NumHostsUp                float64 `json:"num_hosts_up"`
        NumServicesAcknowledged   float64 `json:"num_services_acknowledged"`
        NumServicesCritical       float64 `json:"num_services_critical"`
        NumServicesFlapping       float64 `json:"num_services_flapping"`
        NumServicesInDowntime     float64 `json:"num_services_in_downtime"`
        NumServicesOk             float64 `json:"num_services_ok"`
        NumServicesPending        float64 `json:"num_services_pending"`
        NumServicesUnknown        float64 `json:"num_services_unknown"`
        NumServicesUnreachable    float64 `json:"num_services_unreachable"`
        NumServicesWarning        float64 `json:"num_services_warning"`
        PassiveHostChecks         float64 `json:"passive_host_checks"`
        PassiveHostChecks15Min    float64 `json:"passive_host_checks_15min"`
        PassiveHostChecks1Min     float64 `json:"passive_host_checks_1min"`
        PassiveHostChecks5Min     float64 `json:"passive_host_checks_5min"`
        PassiveServiceChecks      float64 `json:"passive_service_checks"`
        PassiveServiceChecks15Min float64 `json:"passive_service_checks_15min"`
        PassiveServiceChecks1Min  float64 `json:"passive_service_checks_1min"`
        PassiveServiceChecks5Min  float64 `json:"passive_service_checks_5min"`
        Uptime                    float64 `json:"uptime"`
    } `json:"status"`
}

type SummaryMetrics struct {
    RawResult []json.RawMessage`json:"results"`
    Cib *CIB
}

