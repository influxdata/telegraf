package filebeat

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/internal"
    "github.com/influxdata/telegraf/plugins/inputs"
)

const (
    defaultURL = "http://localhost:9602/debug/vars"
)

type Filebeat struct {
    URLs []string `toml:"urls"`

    Timeout internal.Duration

    client *http.Client
}

func (*Filebeat) Description() string {
    return "Read Filebeat-formatted JSON metrics from one or more HTTP endpoints"
}

func (*Filebeat) SampleConfig() string {
    return `
  ## Multiple URLs from which to read Filebeat-formatted JSON
  ## Default is "http://localhost:9602/debug/vars".
  urls = [
    "http://localhost:9602/debug/vars"
  ]

  ## Time limit for http requests
  timeout = "5s"
`
}

func (f *Filebeat) Gather(acc telegraf.Accumulator) error {
    if f.client == nil {
        f.client = &http.Client{Timeout: f.Timeout.Duration}
    }

    var wg sync.WaitGroup
    for _, u := range f.URLs {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            err := f.gatherURL(acc, url)
            if err != nil {
                acc.AddError(fmt.Errorf("[url=%s]: %s", url, err))
            }
        }(u)
    }

    wg.Wait()
    return nil
}

type memstats struct {
    Alloc         int64   `json:"Alloc"`
    TotalAlloc    int64   `json:"TotalAlloc"`
    Sys           int64   `json:"Sys"`
    Lookups       int64   `json:"Lookups"`
    Mallocs       int64   `json:"Mallocs"`
    Frees         int64   `json:"Frees"`
    HeapAlloc     int64   `json:"HeapAlloc"`
    HeapSys       int64   `json:"HeapSys"`
    HeapIdle      int64   `json:"HeapIdle"`
    HeapInuse     int64   `json:"HeapInuse"`
    HeapReleased  int64   `json:"HeapReleased"`
    HeapObjects   int64   `json:"HeapObjects"`
    StackInuse    int64   `json:"StackInuse"`
    StackSys      int64   `json:"StackSys"`
    MSpanInuse    int64   `json:"MSpanInuse"`
    MSpanSys      int64   `json:"MSpanSys"`
    MCacheInuse   int64   `json:"MCacheInuse"`
    MCacheSys     int64   `json:"MCacheSys"`
    BuckHashSys   int64   `json:"BuckHashSys"`
    GCSys         int64   `json:"GCSys"`
    OtherSys      int64   `json:"OtherSys"`
    NextGC        int64   `json:"NextGC"`
    LastGC        int64   `json:"LastGC"`
    PauseTotalNs  int64   `json:"PauseTotalNs"`
    NumGC         int64   `json:"NumGC"`
    GCCPUFraction float64 `json:"GCCPUFraction"`
}

type stats struct {
    CmdLine                                        []string  `json:"cmdline"`
    MemStats                                       *memstats `json:"memstats"`

    PublishEvents                                  int64       `json:"publish.events"`
    RegistrarStatesCleanup                         int64       `json:"registrar.states.cleanup"`
    RegistrarStatesCurrent                         int64       `json:"registrar.states.current"`
    RegistrarStatesUpdate                          int64       `json:"registrar.states.update"`
    RegistrarWrites                                int64       `json:"registrar.writes"`

    FilebeatHarvesterClosed                        int64       `json:"filebeat.harvester.closed"`
    FilebeatHarvesterFilesTruncated                int64       `json:"filebeat.harvester.files.truncated"`
    FilebeatHarvesterOpen_files                    int64       `json:"filebeat.harvester.open_files"`
    FilebeatHarvesterRunning                       int64       `json:"filebeat.harvester.running"`
    FilebeatHarvesterSkipped                       int64       `json:"filebeat.harvester.skipped"`
    FilebeatHarvesterStarted                       int64       `json:"filebeat.harvester.started"`
    FilebeatProspectorLogFilesRenamed              int64       `json:"filebeat.prospector.log.files.renamed"`
    FilebeatProspectorLogFilesTruncated            int64       `json:"filebeat.prospector.log.files.truncated"`

    LibbeatConfigModuleRunning                     int64       `json:"libbeat.config.module.running"`
    LibbeatConfigModuleStarts                      int64       `json:"libbeat.config.module.starts"`
    LibbeatConfigModuleStops                       int64       `json:"libbeat.config.module.stops"`
    LibbeatConfigReloads                           int64       `json:"libbeat.config.reloads"`
    LibbeatEsCall_countPublishEvents               int64       `json:"libbeat.es.call_count.PublishEvents"`
    LibbeatEsPublishRead_bytes                     int64       `json:"libbeat.es.publish.read_bytes"`
    LibbeatEsPublishRead_errors                    int64       `json:"libbeat.es.publish.read_errors"`
    LibbeatEsPublishWrite_bytes                    int64       `json:"libbeat.es.publish.write_bytes"`
    LibbeatEsPublishWrite_errors                   int64       `json:"libbeat.es.publish.write_errors"`
    LibbeatEsPublished_and_acked_events            int64       `json:"libbeat.es.published_and_acked_events"`
    LibbeatEsPublished_but_not_acked_events        int64       `json:"libbeat.es.published_but_not_acked_events"`
    LibbeatKafkaCall_countPublishEvents            int64       `json:"libbeat.kafka.call_count.PublishEvents"`
    LibbeatKafkaPublished_and_acked_events         int64       `json:"libbeat.kafka.published_and_acked_events"`
    LibbeatKafkaPublished_but_not_acked_events     int64       `json:"libbeat.kafka.published_but_not_acked_events"`
    LibbeatLogstashCall_countPublishEvents         int64       `json:"libbeat.logstash.call_count.PublishEvents"`
    LibbeatLogstashPublishRead_bytes               int64       `json:"libbeat.logstash.publish.read_bytes"`
    LibbeatLogstashPublishRead_errors              int64       `json:"libbeat.logstash.publish.read_errors"`
    LibbeatLogstashPublishWrite_bytes              int64       `json:"libbeat.logstash.publish.write_bytes"`
    LibbeatLogstashPublishWrite_errors             int64       `json:"libbeat.logstash.publish.write_errors"`
    LibbeatLogstashPublished_and_acked_events      int64       `json:"libbeat.logstash.published_and_acked_events"`
    LibbeatLogstashPublished_but_not_acked_events  int64       `json:"libbeat.logstash.published_but_not_acked_events"`
    LibbeatOutputsMessages_dropped                 int64       `json:"libbeat.outputs.messages_dropped"`
    LibbeatPublisherMessages_in_worker_queues      int64       `json:"libbeat.publisher.messages_in_worker_queues"`
    LibbeatPublisherPublished_events               int64       `json:"libbeat.publisher.published_events"`
    LibbeatRedisPublishRead_bytes                  int64       `json:"libbeat.redis.publish.read_bytes"`
    LibbeatRedisPublishRead_errors                 int64       `json:"libbeat.redis.publish.read_errors"`
    LibbeatRedisPublishWrite_bytes                 int64       `json:"libbeat.redis.publish.write_bytes"`
    LibbeatRedisPublishWrite_errors                int64       `json:"libbeat.redis.publish.write_errors"`
}


// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (f *Filebeat) gatherURL(
    acc telegraf.Accumulator,
    url string,
) error {
    now := time.Now()

    resp, err := f.client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    dec := json.NewDecoder(resp.Body)

    var s stats
    err = dec.Decode(&s)
    if err != nil {
        return err
    }

    if s.MemStats != nil {
        acc.AddFields("filebeat_memstats",
            map[string]interface{}{
                "alloc_bytes":         s.MemStats.Alloc,
                "buck_hash_sys_bytes": s.MemStats.BuckHashSys,
                "frees":               s.MemStats.Frees,
                "gcc_pu_fraction":     s.MemStats.GCCPUFraction,
                "gc_sys_bytes":        s.MemStats.GCSys,
                "heap_alloc_bytes":    s.MemStats.HeapAlloc,
                "heap_idle_bytes":     s.MemStats.HeapIdle,
                "heap_in_use_bytes":   s.MemStats.HeapInuse,
                "heap_objects":        s.MemStats.HeapObjects,
                "heap_released_bytes": s.MemStats.HeapReleased,
                "heap_sys_bytes":      s.MemStats.HeapSys,
                "last_gc_ns":          s.MemStats.LastGC,
                "lookups":             s.MemStats.Lookups,
                "mallocs":             s.MemStats.Mallocs,
                "mcache_in_use_bytes": s.MemStats.MCacheInuse,
                "mcache_sys_bytes":    s.MemStats.MCacheSys,
                "mspan_in_use_bytes":  s.MemStats.MSpanInuse,
                "mspan_sys_bytes":     s.MemStats.MSpanSys,
                "next_gc_ns":          s.MemStats.NextGC,
                "num_gc":              s.MemStats.NumGC,
                "other_sys_bytes":     s.MemStats.OtherSys,
                "pause_total_ns":      s.MemStats.PauseTotalNs,
                "stack_in_use_bytes":  s.MemStats.StackInuse,
                "stack_sys_bytes":     s.MemStats.StackSys,
                "sys_bytes":           s.MemStats.Sys,
                "total_alloc_bytes":   s.MemStats.TotalAlloc,
            },
            map[string]string{
                "url":        url,
            },
            now)
    }

    acc.AddFields("filebeat",
        map[string]interface{} {
            "publish_events"                          : s.PublishEvents,
            "registrar_states_cleanup"                : s.RegistrarStatesCleanup,
            "registrar_states_current"                : s.RegistrarStatesCurrent,
            "registrar_states_update"                 : s.RegistrarStatesUpdate,
            "registrar_writes"                        : s.RegistrarWrites,

            "harvester_closed"                        : s.FilebeatHarvesterClosed,
            "harvester_files_truncated"               : s.FilebeatHarvesterFilesTruncated,
            "harvester_open_files"                    : s.FilebeatHarvesterOpen_files,
            "harvester_running"                       : s.FilebeatHarvesterRunning,
            "harvester_skipped"                       : s.FilebeatHarvesterSkipped,
            "harvester_started"                       : s.FilebeatHarvesterStarted,
            "prospector_log_files_renamed"            : s.FilebeatProspectorLogFilesRenamed,
            "prospector_log_files_truncated"          : s.FilebeatProspectorLogFilesTruncated,
        },
        map[string]string{
            "url":        url,
        },
        now)

    acc.AddFields("libbeat",
        map[string]interface{}{
            "config_module_running"                   : s.LibbeatConfigModuleRunning,
            "config_module_starts"                    : s.LibbeatConfigModuleStarts,
            "config_module_stops"                     : s.LibbeatConfigModuleStops,
            "config_reloads"                          : s.LibbeatConfigReloads,
            "es_call_count_publish_events"            : s.LibbeatEsCall_countPublishEvents,
            "es_publish_read_bytes"                   : s.LibbeatEsPublishRead_bytes,
            "es_publish_read_errors"                  : s.LibbeatEsPublishRead_errors,
            "es_publish_write_bytes"                  : s.LibbeatEsPublishWrite_bytes,
            "es_publish_write_errors"                 : s.LibbeatEsPublishWrite_errors,
            "es_published_and_acked_events"           : s.LibbeatEsPublished_and_acked_events,
            "es_published_but_not_acked_events"       : s.LibbeatEsPublished_but_not_acked_events,
            "kafka_call_count_publishevents"          : s.LibbeatKafkaCall_countPublishEvents,
            "kafka_published_and_acked_events"        : s.LibbeatKafkaPublished_and_acked_events,
            "kafka_published_but_not_acked_events"    : s.LibbeatKafkaPublished_but_not_acked_events,
            "logstash_call_count_publishevents"       : s.LibbeatLogstashCall_countPublishEvents,
            "logstash_publish_read_bytes"             : s.LibbeatLogstashPublishRead_bytes,
            "logstash_publish_read_errors"            : s.LibbeatLogstashPublishRead_errors,
            "logstash_publish_write_bytes"            : s.LibbeatLogstashPublishWrite_bytes,
            "logstash_publish_write_errors"           : s.LibbeatLogstashPublishWrite_errors,
            "logstash_published_and_acked_events"     : s.LibbeatLogstashPublished_and_acked_events,
            "logstash_published_but_not_acked_events" : s.LibbeatLogstashPublished_but_not_acked_events,
            "outputs_messages_dropped"                : s.LibbeatOutputsMessages_dropped,
            "publisher_messages_in_worker_queues"     : s.LibbeatPublisherMessages_in_worker_queues,
            "publisher_published_events"              : s.LibbeatPublisherPublished_events,
            "redis_publish_read_bytes"                : s.LibbeatRedisPublishRead_bytes,
            "redis_publish_read_errors"               : s.LibbeatRedisPublishRead_errors,
            "redis_publish_write_bytes"               : s.LibbeatRedisPublishWrite_bytes,
            "redis_publish_write_errors"              : s.LibbeatRedisPublishWrite_errors,
        },
        map[string]string{
            "url":        url,
        },
        now)

    return nil
}

func init() {
    inputs.Add("filebeat", func() telegraf.Input {
        return &Filebeat{
            URLs:    []string{defaultURL},
            Timeout: internal.Duration{Duration: time.Second * 5},
        }
    })
}
