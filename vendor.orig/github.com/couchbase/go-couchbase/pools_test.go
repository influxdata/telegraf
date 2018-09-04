package couchbase

import (
	"encoding/json"
	"sync"
	"testing"
	"unsafe"
)

var samplePools = `{
    "componentsVersion": {
        "ale": "eff9516",
        "couch": "1.2.0a-7dd003e-git",
        "couch_set_view": "1.2.0a-7dd003e-git",
        "crypto": "2.0.3",
        "ibrowse": "2.2.0",
        "inets": "5.6",
        "kernel": "2.14.4",
        "mnesia": "4.4.19",
        "mochiweb": "1.4.1",
        "ns_server": "2.0.0r-388-gf35126e-community",
        "oauth": "7d85d3ef",
        "os_mon": "2.2.6",
        "public_key": "0.12",
        "sasl": "2.1.9.4",
        "ssl": "4.1.5",
        "stdlib": "1.17.4"
    },
    "implementationVersion": "2.0.0r-388-gf35126e-community",
    "isAdminCreds": false,
    "pools": [
        {
            "name": "default",
            "streamingUri": "/poolsStreaming/default",
            "uri": "/pools/default"
        }
    ],
    "uuid": "debbb353c26a6ee1c9eceb748a8c6907"
}`

var samplePool = `{
    "alerts": [],
    "autoCompactionSettings": {
        "databaseFragmentationThreshold": 80,
        "parallelDBAndViewCompaction": false,
        "viewFragmentationThreshold": 80
    },
    "balanced": true,
    "buckets": {
        "uri": "/pools/default/buckets?v=118084983"
    },
    "controllers": {
        "addNode": {
            "uri": "/controller/addNode"
        },
        "ejectNode": {
            "uri": "/controller/ejectNode"
        },
        "failOver": {
            "uri": "/controller/failOver"
        },
        "reAddNode": {
            "uri": "/controller/reAddNode"
        },
        "rebalance": {
            "uri": "/controller/rebalance"
        },
        "replication": {
            "createURI": "/controller/createReplication",
            "infosURI": "/couchBase/_replicator/_design/_replicator_info/_view/infos?group_level=1",
            "replicatorDBURI": "/couchBase/_replicator"
        },
        "setAutoCompaction": {
            "uri": "/controller/setAutoCompaction",
            "validateURI": "/controller/setAutoCompaction?just_validate=1"
        }
    },
    "counters": {
        "rebalance_start": 1,
        "rebalance_success": 1
    },
    "failoverWarnings": [],
    "name": "default",
    "nodeStatusesUri": "/nodeStatuses",
    "nodes": [
        {
            "clusterCompatibility": 1,
            "clusterMembership": "active",
            "couchApiBase": "http://10.203.6.236:8092/",
            "hostname": "10.203.6.236:8091",
            "interestingStats": {
                "curr_items": 0,
                "curr_items_tot": 0,
                "vb_replica_curr_items": 0
            },
            "mcdMemoryAllocated": 5978,
            "mcdMemoryReserved": 5978,
            "memoryFree": 6891118592,
            "memoryTotal": 7836254208,
            "os": "x86_64-unknown-linux-gnu",
            "ports": {
                "direct": 11210,
                "proxy": 11211
            },
            "status": "healthy",
            "systemStats": {
                "cpu_utilization_rate": 0.5025125628140703,
                "swap_total": 4294963200,
                "swap_used": 0
            },
            "thisNode": true,
            "uptime": "20516",
            "version": "2.0.0r-388-gf35126e-community"
        },
        {
            "clusterCompatibility": 1,
            "clusterMembership": "active",
            "couchApiBase": "http://10.32.21.163:8092/",
            "hostname": "10.32.21.163:8091",
            "interestingStats": {
                "curr_items": 0,
                "curr_items_tot": 0,
                "vb_replica_curr_items": 0
            },
            "mcdMemoryAllocated": 5978,
            "mcdMemoryReserved": 5978,
            "memoryFree": 6959566848,
            "memoryTotal": 7836254208,
            "os": "x86_64-unknown-linux-gnu",
            "ports": {
                "direct": 11210,
                "proxy": 11211
            },
            "status": "healthy",
            "systemStats": {
                "cpu_utilization_rate": 0.7575757575757576,
                "swap_total": 4294963200,
                "swap_used": 0
            },
            "uptime": "20523",
            "version": "2.0.0r-388-gf35126e-community"
        },
        {
            "clusterCompatibility": 1,
            "clusterMembership": "active",
            "couchApiBase": "http://10.98.83.17:8092/",
            "hostname": "10.98.83.17:8091",
            "interestingStats": {
                "curr_items": 0,
                "curr_items_tot": 0,
                "vb_replica_curr_items": 0
            },
            "mcdMemoryAllocated": 5978,
            "mcdMemoryReserved": 5978,
            "memoryFree": 6960541696,
            "memoryTotal": 7836254208,
            "os": "x86_64-unknown-linux-gnu",
            "ports": {
                "direct": 11210,
                "proxy": 11211
            },
            "status": "healthy",
            "systemStats": {
                "cpu_utilization_rate": 0.24213075060532688,
                "swap_total": 4294963200,
                "swap_used": 0
            },
            "uptime": "20505",
            "version": "2.0.0r-388-gf35126e-community"
        },
        {
            "clusterCompatibility": 1,
            "clusterMembership": "active",
            "couchApiBase": "http://10.34.21.232:8092/",
            "hostname": "10.34.21.232:8091",
            "interestingStats": {
                "curr_items": 0,
                "curr_items_tot": 0,
                "vb_replica_curr_items": 0
            },
            "mcdMemoryAllocated": 5978,
            "mcdMemoryReserved": 5978,
            "memoryFree": 6961504256,
            "memoryTotal": 7836254208,
            "os": "x86_64-unknown-linux-gnu",
            "ports": {
                "direct": 11210,
                "proxy": 11211
            },
            "status": "healthy",
            "systemStats": {
                "cpu_utilization_rate": 0.7334963325183375,
                "swap_total": 4294963200,
                "swap_used": 0
            },
            "uptime": "20528",
            "version": "2.0.0r-388-gf35126e-community"
        },
        {
            "clusterCompatibility": 1,
            "clusterMembership": "active",
            "couchApiBase": "http://10.203.33.4:8092/",
            "hostname": "10.203.33.4:8091",
            "interestingStats": {
                "curr_items": 0,
                "curr_items_tot": 0,
                "vb_replica_curr_items": 0
            },
            "mcdMemoryAllocated": 5978,
            "mcdMemoryReserved": 5978,
            "memoryFree": 6960599040,
            "memoryTotal": 7836254208,
            "os": "x86_64-unknown-linux-gnu",
            "ports": {
                "direct": 11210,
                "proxy": 11211
            },
            "status": "healthy",
            "systemStats": {
                "cpu_utilization_rate": 0.7575757575757576,
                "swap_total": 4294963200,
                "swap_used": 0
            },
            "uptime": "20537",
            "version": "2.0.0r-388-gf35126e-community"
        }
    ],
    "rebalanceProgressUri": "/pools/default/rebalanceProgress",
    "rebalanceStatus": "none",
    "remoteClusters": {
        "uri": "/pools/default/remoteClusters",
        "validateURI": "/pools/default/remoteClusters?just_validate=1"
    },
    "stats": {
        "uri": "/pools/default/stats"
    },
    "stopRebalanceUri": "/controller/stopRebalance",
    "storageTotals": {
        "hdd": {
            "free": 1046325215240,
            "quotaTotal": 1056894156800,
            "total": 1056894156800,
            "used": 10568941560,
            "usedByData": 12543880
        },
        "ram": {
            "quotaTotal": 31341936640,
            "quotaUsed": 31341936640,
            "total": 39181271040,
            "used": 4447940608,
            "usedByData": 13557744
        }
    },
    "tasksProgressUri": "/pools/default/tasksProgress",
    "tasksStatus": "none"
}`

func assert(t *testing.T, name string, got interface{}, expected interface{}) {
	if got != expected {
		t.Fatalf("Expected %v for %s, got %v", expected, name, got)
	}
}

func testParse(t *testing.T, s string, rv interface{}) {
	if err := json.Unmarshal([]byte(s), rv); err != nil {
		t.Fatalf("Error decoding:  %v", err)
	}
}

func TestPoolsResponse(t *testing.T) {
	res := Pools{}
	testParse(t, samplePools, &res)

	assert(t, "couch", res.ComponentsVersion["couch"],
		"1.2.0a-7dd003e-git")
	assert(t, "implementationVersion", res.ImplementationVersion,
		"2.0.0r-388-gf35126e-community")
	assert(t, "uuid", res.UUID, "debbb353c26a6ee1c9eceb748a8c6907")
	assert(t, "IsAdmin", res.IsAdmin, false)
	assert(t, "pool name", res.Pools[0].Name, "default")
	assert(t, "pool streamingUri", res.Pools[0].StreamingURI,
		"/poolsStreaming/default")
	assert(t, "pool URI", res.Pools[0].URI, "/pools/default")
}

func TestPool(t *testing.T) {
	res := Pool{}
	testParse(t, samplePool, &res)
	assert(t, "len(pools)", 5, len(res.Nodes))
}

func TestCommonAddressSuffixEmpty(t *testing.T) {
	b := Bucket{nodeList: mkNL([]Node{})}
	assert(t, "empty", "", b.CommonAddressSuffix())
}

func TestCommonAddressSuffixUncommon(t *testing.T) {
	b := Bucket{vBucketServerMap: unsafe.Pointer(&VBucketServerMap{
		ServerList: []string{"somestring", "unrelated"}}),
		nodeList: mkNL([]Node{}),
	}
	assert(t, "shouldn't match", "", b.CommonAddressSuffix())
}

func TestCommonAddressSuffixCommon(t *testing.T) {
	b := Bucket{nodeList: unsafe.Pointer(&[]Node{
		{Hostname: "server1.example.com:11210"},
		{Hostname: "server2.example.com:11210"},
		{Hostname: "server3.example.com:11210"},
		{Hostname: "server4.example.com:11210"},
	})}
	assert(t, "useful suffix", ".example.com:11210",
		b.CommonAddressSuffix())
}

func TestBucketConnPool(t *testing.T) {
	b := Bucket{}
	b.replaceConnPools([]*connectionPool{})
	p := b.getConnPool(3)
	if p != nil {
		t.Fatalf("Successfully got a pool where there was none: %v", p)
	}
	// TODO: I have a few more cases to cover here.
}

// No assertions, but this is meant to be tested with the race
// detector to verify the connection pool stuff is clean.
func TestBucketConnPoolConcurrent(t *testing.T) {
	b := Bucket{}

	wg := sync.WaitGroup{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				b.replaceConnPools([]*connectionPool{})
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
