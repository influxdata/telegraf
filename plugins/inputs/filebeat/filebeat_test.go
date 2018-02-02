package filebeat_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/influxdata/telegraf/plugins/inputs/filebeat"
    "github.com/influxdata/telegraf/testutil"
    "github.com/stretchr/testify/require"
)

func TestFilebeat(t *testing.T) {
    fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/endpoint" {
            _, _ = w.Write([]byte(filebeatReturn))
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer fakeInfluxServer.Close()

    plugin := &filebeat.Filebeat{
        URLs: []string{fakeInfluxServer.URL + "/endpoint"},
    }

    var acc testutil.Accumulator
    require.NoError(t, plugin.Gather(&acc))

    // there should be 3 metric types (filebeat, libbbeat, and filebeat_memstats)
    require.Len(t, acc.Metrics, 3)

    tags := map[string]string{
        "url": fakeInfluxServer.URL + "/endpoint",
    }


    filebeat_memstats_fields := map[string]interface{}{
        "alloc_bytes":         int64(1593856),
        "buck_hash_sys_bytes": int64(1473183),
        "frees":               int64(4420902),
        "gcc_pu_fraction":     float64(2.090470368793657e-05),
        "gc_sys_bytes":        int64(419840),
        "heap_alloc_bytes":    int64(1593856),
        "heap_idle_bytes":     int64(3153920),
        "heap_in_use_bytes":   int64(2482176),
        "heap_objects":        int64(7809),
        "heap_released_bytes": int64(2957312),
        "heap_sys_bytes":      int64(5636096),
        "last_gc_ns":          int64(1505123302765259120),
        "lookups":             int64(160),
        "mallocs":             int64(4428711),
        "mcache_in_use_bytes": int64(4800),
        "mcache_sys_bytes":    int64(16384),
        "mspan_in_use_bytes":  int64(35040),
        "mspan_sys_bytes":     int64(81920),
        "next_gc_ns":          int64(4194304),
        "num_gc":              int64(2091),
        "other_sys_bytes":     int64(1226329),
        "pause_total_ns":      int64(3910987717),
        "stack_in_use_bytes":  int64(655360),
        "stack_sys_bytes":     int64(655360),
        "sys_bytes":           int64(9509112),
        "total_alloc_bytes":   int64(1530571800),
    }
    acc.AssertContainsTaggedFields(t, "filebeat_memstats", filebeat_memstats_fields, tags)

    filebeat_fields := map[string]interface{}{
        "publish_events"                          : int64(13),
        "registrar_states_cleanup"                : int64(4),
        "registrar_states_current"                : int64(1),
        "registrar_states_update"                 : int64(13),
        "registrar_writes"                        : int64(11),

        "harvester_closed"                        : int64(4),
        "harvester_files_truncated"               : int64(0),
        "harvester_open_files"                    : int64(0),
        "harvester_running"                       : int64(0),
        "harvester_skipped"                       : int64(0),
        "harvester_started"                       : int64(4),
        "prospector_log_files_renamed"            : int64(0),
        "prospector_log_files_truncated"          : int64(0),
    }
    acc.AssertContainsTaggedFields(t, "filebeat", filebeat_fields, tags)

    libbeat_fields := map[string]interface{}{
        "config_module_running"                   : int64(0),
        "config_module_starts"                    : int64(0),
        "config_module_stops"                     : int64(0),
        "config_reloads"                          : int64(0),
        "es_call_count_publish_events"            : int64(0),
        "es_publish_read_bytes"                   : int64(0),
        "es_publish_read_errors"                  : int64(0),
        "es_publish_write_bytes"                  : int64(0),
        "es_publish_write_errors"                 : int64(0),
        "es_published_and_acked_events"           : int64(0),
        "es_published_but_not_acked_events"       : int64(0),
        "kafka_call_count_publishevents"          : int64(0),
        "kafka_published_and_acked_events"        : int64(0),
        "kafka_published_but_not_acked_events"    : int64(0),
        "logstash_call_count_publishevents"       : int64(0),
        "logstash_publish_read_bytes"             : int64(0),
        "logstash_publish_read_errors"            : int64(0),
        "logstash_publish_write_bytes"            : int64(0),
        "logstash_publish_write_errors"           : int64(0),
        "logstash_published_and_acked_events"     : int64(0),
        "logstash_published_but_not_acked_events" : int64(0),
        "outputs_messages_dropped"                : int64(0),
        "publisher_messages_in_worker_queues"     : int64(0),
        "publisher_published_events"              : int64(0),
        "redis_publish_read_bytes"                : int64(0),
        "redis_publish_read_errors"               : int64(0),
        "redis_publish_write_bytes"               : int64(0),
        "redis_publish_write_errors"              : int64(0),
    }
    acc.AssertContainsTaggedFields(t, "libbeat", libbeat_fields, tags)
}

func TestMissingStats(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{}`))
    }))
    defer server.Close()

    plugin := &filebeat.Filebeat{
        URLs: []string{server.URL},
    }

    var acc testutil.Accumulator
    plugin.Gather(&acc)

    require.False(t, acc.HasField("filebeat_memstats", "alloc_bytes"))
    require.True(t, acc.HasField("filebeat", "publish_events"))
    require.True(t, acc.HasField("libbeat", "config_module_running"))
}

func TestErrorHandling(t *testing.T) {
    badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/endpoint" {
            _, _ = w.Write([]byte("not json"))
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer badServer.Close()

    plugin := &filebeat.Filebeat{
        URLs: []string{badServer.URL + "/endpoint"},
    }

    var acc testutil.Accumulator
    plugin.Gather(&acc)
    acc.WaitError(1)
    require.Equal(t, uint64(0), acc.NMetrics())
}

func TestErrorHandling404(t *testing.T) {
    badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer badServer.Close()

    plugin := &filebeat.Filebeat{
        URLs: []string{badServer.URL},
    }

    var acc testutil.Accumulator
    plugin.Gather(&acc)
    acc.WaitError(1)
    require.Equal(t, uint64(0), acc.NMetrics())
}

const filebeatReturn = `{
"cmdline": ["/opt/OSAGbeat/bin/filebeat","-path.home","/opt/OSAGbeat","-path.data","/shared/data/OSAGbeat","-path.config","/opt/OSAGbeat/etc","-httpprof","localhost:9602"],
"filebeat.harvester.closed": 4,
"filebeat.harvester.files.truncated": 0,
"filebeat.harvester.open_files": 0,
"filebeat.harvester.running": 0,
"filebeat.harvester.skipped": 0,
"filebeat.harvester.started": 4,
"filebeat.prospector.log.files.renamed": 0,
"filebeat.prospector.log.files.truncated": 0,
"libbeat.config.module.running": 0,
"libbeat.config.module.starts": 0,
"libbeat.config.module.stops": 0,
"libbeat.config.reloads": 0,
"libbeat.es.call_count.PublishEvents": 0,
"libbeat.es.publish.read_bytes": 0,
"libbeat.es.publish.read_errors": 0,
"libbeat.es.publish.write_bytes": 0,
"libbeat.es.publish.write_errors": 0,
"libbeat.es.published_and_acked_events": 0,
"libbeat.es.published_but_not_acked_events": 0,
"libbeat.kafka.call_count.PublishEvents": 0,
"libbeat.kafka.published_and_acked_events": 0,
"libbeat.kafka.published_but_not_acked_events": 0,
"libbeat.logstash.call_count.PublishEvents": 0,
"libbeat.logstash.publish.read_bytes": 0,
"libbeat.logstash.publish.read_errors": 0,
"libbeat.logstash.publish.write_bytes": 0,
"libbeat.logstash.publish.write_errors": 0,
"libbeat.logstash.published_and_acked_events": 0,
"libbeat.logstash.published_but_not_acked_events": 0,
"libbeat.outputs.messages_dropped": 0,
"libbeat.publisher.messages_in_worker_queues": 0,
"libbeat.publisher.published_events": 0,
"libbeat.redis.publish.read_bytes": 0,
"libbeat.redis.publish.read_errors": 0,
"libbeat.redis.publish.write_bytes": 0,
"libbeat.redis.publish.write_errors": 0,
"memstats": {"Alloc":1593856,"TotalAlloc":1530571800,"Sys":9509112,"Lookups":160,"Mallocs":4428711,"Frees":4420902,"HeapAlloc":1593856,"HeapSys":5636096,"HeapIdle":3153920,"HeapInuse":2482176,"HeapReleased":2957312,"HeapObjects":7809,"StackInuse":655360,"StackSys":655360,"MSpanInuse":35040,"MSpanSys":81920,"MCacheInuse":4800,"MCacheSys":16384,"BuckHashSys":1473183,"GCSys":419840,"OtherSys":1226329,"NextGC":4194304,"LastGC":1505123302765259120,"PauseTotalNs":3910987717,"PauseNs":[452560,2904248,6489285,474258,336898,4351492,570041,493726,208432,502915,480520,446599,389969,1928560,442359,479284,538046,481046,1188760,7760382,592605,442524,4901294,5937587,6248907,3774914,5715664,445746,419601,1752162,487722,2172108,515872,2239079,282372,5090386,5697276,358877,21535602,500555,546029,193172,348039,3329688,4778114,1599740,1198320,519223,501066,5947017,538719,386455,1884998,1154332,479002,1779702,429338,412454,503249,431967,385074,349869,497753,455742,5972071,392049,5458662,359374,4718031,3853326,455978,4365420,531150,518323,246479,3815736,2890667,1190818,4338222,631184,522890,6292335,3167848,669604,398166,2507638,464265,484831,4075702,2829186,1197969,5163332,3903629,349242,430308,8133501,508672,6486490,4791813,4405987,5762092,391571,166345,304736,5263455,411965,8762222,446260,441136,1548804,348662,450885,452150,626231,5023273,2703214,6845268,3466666,400497,391494,1802378,1094017,445867,4988261,420411,220418,2075724,2876825,691683,3402210,3734368,2159163,492843,447170,380190,6094068,7729292,4140273,1127844,194784,4741197,1253894,471279,580162,381437,5096494,534345,569931,261521,421754,348061,528074,5345159,3466401,7873055,3402813,422606,1698020,3808176,450013,471752,564590,402549,497345,469633,525449,2603160,3150215,1467892,10580244,2956776,3999065,1221977,3313463,2780649,529425,503400,3397409,400282,389815,4411550,373062,515467,2094592,4445937,2444448,437138,6318914,3771642,503842,417851,364990,8138047,235232,337501,423333,398745,3650864,1822359,986874,540499,952006,459553,441679,3299373,445546,428062,979273,290505,415152,471324,2518263,3198212,1040357,305228,1234512,1404740,419924,641650,410788,2516233,339644,383357,340171,1182948,494845,340765,265845,550072,3810410,318845,2210350,376388,507209,2538652,394902,445022,5298119,3257383,1707633,4971696,466608,457214,1236620,505651,454734,3288603,329985,418371,1154116,2871559,519622,3307376,410067,8806543,477909],"PauseEnd":[1505118252986807907,1505118372993734946,1505118493005068759,1505118613012035994,1505118733016458955,1505118853024045715,1505118973029051855,1505119093039480628,1505119213043816711,1505119333049802465,1505119453056018142,1505119573060766555,1505119693065716225,1505119813075782267,1505119933081083705,1505120053090364129,1505120173099920073,1505120293105563550,1505120413110733794,1505120533123089203,1505120653766653555,1505120774766295737,1505120895769072489,1505121016769524258,1505121137493081622,1505121257759003036,1505121377768373508,1505121498176723269,1505121618186505948,1505121738192245891,1505121858198722596,1505121978208628715,1505122098212409108,1505122218220807806,1505122338223699656,1505122458235626630,1505122578768078720,1505122699764759545,1505122820031111508,1505122940766375114,1505123061766639849,1505123182757279843,1505123302765259120,1505092632765974852,1505092753767528284,1505092874765478488,1505092995766074062,1505093116357241548,1505093236362703401,1505093356370739857,1505093476427924520,1505093596433941827,1505093716439443633,1505093836444436933,1505093956448868738,1505094076456670311,1505094196460920552,1505094316472740350,1505094436766146502,1505094557758422290,1505094677765950742,1505094798766080897,1505094919767030978,1505095040105821427,1505095160768971360,1505095281470727528,1505095401479217708,1505095521766141835,1505095642762268466,1505095763766699788,1505095884766069375,1505096005768438961,1505096126519645168,1505096246529617490,1505096366538193631,1505096486548041338,1505096606556491311,1505096726560461819,1505096846568284731,1505096966574761913,1505097086584131778,1505097206595208000,1505097326602544799,1505097446607047872,1505097566765069080,1505097687757395583,1505097807764765989,1505097928765753479,1505098049766703885,1505098170093303380,1505098290103780857,1505098410767908408,1505098531673787131,1505098651683234911,1505098771764576039,1505098892763065313,1505099013766368217,1505099134770574031,1505099255768967814,1505099376684156663,1505099496739125471,1505099616743044626,1505099736749580715,1505099856754999772,1505099976764024932,1505100097757293807,1505100217771493407,1505100338766499669,1505100459766268531,1505100580107699475,1505100700768632578,1505100821765615904,1505100941803164543,1505101061808775255,1505101181818335589,1505101302757539639,1505101422769614884,1505101543766229071,1505101664766389755,1505101785766045959,1505101906765735347,1505102026774791389,1505102146779038958,1505102266870433996,1505102387758669945,1505102507765379090,1505102628766599507,1505102749767003827,1505102870099844069,1505102990766692660,1505103111766505096,1505103231929111571,1505103351937242347,1505103471945415873,1505103591950711799,1505103711960269506,1505103832762455924,1505103953766786243,1505104074764819579,1505104195764266610,1505104316768388262,1505104436864840519,1505104556869317208,1505104676873039466,1505104796880524569,1505104916889719304,1505105037060427090,1505105157072164843,1505105277081771272,1505105397092029251,1505105517097857834,1505105637107795853,1505105757121662170,1505105877134934831,1505105997148608530,1505106117158470455,1505106237164786932,1505106357174504828,1505106477183509814,1505106597191726574,1505106717203102753,1505106837214384412,1505106957224207849,1505107077232279020,1505107197242315760,1505107317252182071,1505107437263018215,1505107557272457179,1505107677281074937,1505107797297890526,1505107917306559327,1505108037317913362,1505108157323729607,1505108277333444454,1505108397339530409,1505108517350062431,1505108637359443152,1505108757366634035,1505108877374003964,1505108997383679007,1505109117397715303,1505109237403307000,1505109357410434611,1505109477418940048,1505109597472860168,1505109717484376624,1505109837492357317,1505109957507127137,1505110077513363301,1505110197522218243,1505110317533350322,1505110437544034655,1505110557554401509,1505110677758271964,1505110797772932877,1505110918765280663,1505111039765602475,1505111160097348218,1505111280765447655,1505111401769113613,1505111522599123005,1505111642603993116,1505111762612083042,1505111882616943152,1505112002626354927,1505112122632544314,1505112242641406410,1505112362650422060,1505112482658226936,1505112602667699008,1505112722676283914,1505112842686894866,1505112962696038317,1505113082708238331,1505113202714322198,1505113322724494264,1505113442733010767,1505113562740362171,1505113682750892798,1505113802757643846,1505113922765586402,1505114042772480017,1505114162777149006,1505114282782456603,1505114402788737839,1505114522792485511,1505114642798577246,1505114762803509807,1505114882809573510,1505115002819706207,1505115122823452309,1505115242833033129,1505115362837274961,1505115482841780129,1505115602849940608,1505115722856478906,1505115842862464026,1505115962871785630,1505116083766596642,1505116204765791830,1505116325769650313,1505116446766086979,1505116567316742033,1505116687323755767,1505116807328627215,1505116927336982925,1505117047758183404,1505117167771233802,1505117287943405496,1505117408766585755,1505117529766935891,1505117650054667085,1505117770766171420,1505117891764945233,1505118012763675191,1505118132982347483],"NumGC":2091,"GCCPUFraction":2.090470368793657e-05,"EnableGC":true,"DebugGC":false,"BySize":[{"Size":0,"Mallocs":0,"Frees":0},{"Size":8,"Mallocs":52077,"Frees":52015},{"Size":16,"Mallocs":1591146,"Frees":1589435},{"Size":32,"Mallocs":213183,"Frees":208913},{"Size":48,"Mallocs":275569,"Frees":275101},{"Size":64,"Mallocs":98722,"Frees":98504},{"Size":80,"Mallocs":368,"Frees":180},{"Size":96,"Mallocs":69287,"Frees":69191},{"Size":112,"Mallocs":224,"Frees":160},{"Size":128,"Mallocs":25037,"Frees":25005},{"Size":144,"Mallocs":49746,"Frees":49715},{"Size":160,"Mallocs":25029,"Frees":24969},{"Size":176,"Mallocs":56811,"Frees":56757},{"Size":192,"Mallocs":11743,"Frees":11738},{"Size":208,"Mallocs":143912,"Frees":143820},{"Size":224,"Mallocs":12,"Frees":11},{"Size":240,"Mallocs":24911,"Frees":24894},{"Size":256,"Mallocs":61,"Frees":49},{"Size":288,"Mallocs":72294,"Frees":72223},{"Size":320,"Mallocs":52,"Frees":46},{"Size":352,"Mallocs":74612,"Frees":74564},{"Size":384,"Mallocs":7,"Frees":5},{"Size":416,"Mallocs":8669,"Frees":8603},{"Size":448,"Mallocs":5,"Frees":5},{"Size":480,"Mallocs":3,"Frees":0},{"Size":512,"Mallocs":50,"Frees":47},{"Size":576,"Mallocs":24880,"Frees":24866},{"Size":640,"Mallocs":16,"Frees":11},{"Size":704,"Mallocs":8340,"Frees":8331},{"Size":768,"Mallocs":3,"Frees":3},{"Size":896,"Mallocs":8603,"Frees":8592},{"Size":1024,"Mallocs":56,"Frees":41},{"Size":1152,"Mallocs":24878,"Frees":24861},{"Size":1280,"Mallocs":10,"Frees":4},{"Size":1408,"Mallocs":4,"Frees":3},{"Size":1536,"Mallocs":4,"Frees":4},{"Size":1664,"Mallocs":8595,"Frees":8582},{"Size":2048,"Mallocs":5472,"Frees":5454},{"Size":2304,"Mallocs":24868,"Frees":24849},{"Size":2560,"Mallocs":6,"Frees":3},{"Size":2816,"Mallocs":3,"Frees":1},{"Size":3072,"Mallocs":0,"Frees":0},{"Size":3328,"Mallocs":9,"Frees":3},{"Size":4096,"Mallocs":60,"Frees":54},{"Size":4608,"Mallocs":24964,"Frees":24952},{"Size":5376,"Mallocs":496,"Frees":488},{"Size":6144,"Mallocs":25370,"Frees":25355},{"Size":6400,"Mallocs":158,"Frees":158},{"Size":6656,"Mallocs":162,"Frees":161},{"Size":6912,"Mallocs":165,"Frees":165},{"Size":8192,"Mallocs":841,"Frees":839},{"Size":8448,"Mallocs":156,"Frees":156},{"Size":8704,"Mallocs":12670,"Frees":12670},{"Size":9472,"Mallocs":12861,"Frees":12849},{"Size":10496,"Mallocs":28539,"Frees":28522},{"Size":12288,"Mallocs":1,"Frees":1},{"Size":13568,"Mallocs":3,"Frees":3},{"Size":14080,"Mallocs":0,"Frees":0},{"Size":16384,"Mallocs":5,"Frees":4},{"Size":16640,"Mallocs":0,"Frees":0},{"Size":17664,"Mallocs":22720,"Frees":22708}]},
"publish.events": 13,
"registrar.states.cleanup": 4,
"registrar.states.current": 1,
"registrar.states.update": 13,
"registrar.writes": 11
}
`
