package jboss

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var jboss_exec_mode_out = `
{
    "outcome" : "success",
    "result" : {
        "launch-type" : "STANDALONE",
        "management-major-version" : 8,
        "management-micro-version" : 0,
        "management-minor-version" : 0,
        "name" : "czc6249qm8",
        "namespaces" : [],
        "organization" : null,
        "process-type" : "Server",
        "product-name" : "JBoss EAP",
        "product-version" : "6.4.21.GA",
        "profile-name" : null,
        "release-codename" : "",
        "release-version" : "6.0.27.Final-redhat-00001",
        "running-mode" : "NORMAL",
        "runtime-configuration-state" : "ok",
        "schema-locations" : [],
        "server-state" : "running",
        "suspend-state" : "RUNNING",
        "uuid" : "d337bad3-2550-4159-87fd-400e5f289033"
    }
}
`
var jboss_exec_mode_out_eap6 = `
{
    "outcome" : "success",
    "result" : {
        "launch-type" : "STANDALONE",
        "local-host-name" : "master",
        "management-major-version" : 1,
        "management-micro-version" : 0,
        "management-minor-version" : 8,
        "name" : "eaputv",
        "namespaces" : [],
        "process-type" : "Domain Controller",
        "product-name" : "EAP",
        "product-version" : "6.4.21.GA",
        "release-codename" : "Janus",
        "release-version" : "7.5.21.Final-redhat-1",
        "schema-locations" : []
    }
}
`

var jboss_host_list = `
{
    "outcome" : "success",
    "result" : [
        "eaputv6",
        "eaputv7",
        "master"
    ]
}
`

var jboss_database_out = `
{
    "outcome" : "success",
    "result" : {
        "installed-drivers" : [
            {
                "driver-name" : "h2",
                "deployment-name" : null,
                "driver-module-name" : "com.h2database.h2",
                "module-slot" : "main",
                "driver-datasource-class-name" : "",
                "driver-xa-datasource-class-name" : "org.h2.jdbcx.JdbcDataSource",
                "driver-class-name" : "org.h2.Driver",
                "driver-major-version" : 1,
                "driver-minor-version" : 3,
                "jdbc-compliant" : true
            },
            {
                "driver-name" : "oracle",
                "deployment-name" : null,
                "driver-module-name" : "com.liferay.portal",
                "module-slot" : "main",
                "driver-datasource-class-name" : "",
                "driver-xa-datasource-class-name" : "",
                "driver-class-name" : "oracle.jdbc.driver.OracleDriver",
                "driver-major-version" : 11,
                "driver-minor-version" : 2,
                "jdbc-compliant" : true
            }
        ],
        "data-source" : {
            "ExampleDS" : {
                "allocation-retry" : null,
                "allocation-retry-wait-millis" : null,
                "allow-multiple-users" : false,
                "background-validation" : null,
                "background-validation-millis" : null,
                "blocking-timeout-wait-millis" : null,
                "check-valid-connection-sql" : null,
                "connectable" : false,
                "connection-url" : "jdbc:h2:mem:test;DB_CLOSE_DELAY=-1;DB_CLOSE_ON_EXIT=FALSE",
                "datasource-class" : null,
                "driver-class" : null,
                "driver-name" : "h2",
                "enabled" : true,
                "exception-sorter-class-name" : null,
                "exception-sorter-properties" : null,
                "flush-strategy" : null,
                "idle-timeout-minutes" : null,
                "jndi-name" : "java:jboss/datasources/ExampleDS",
                "jta" : true,
                "max-pool-size" : null,
                "min-pool-size" : null,
                "new-connection-sql" : null,
                "password" : "PASSWORD",
                "pool-prefill" : null,
                "pool-use-strict-min" : null,
                "prepared-statements-cache-size" : null,
                "query-timeout" : null,
                "reauth-plugin-class-name" : null,
                "reauth-plugin-properties" : null,
                "security-domain" : null,
                "set-tx-query-timeout" : false,
                "share-prepared-statements" : false,
                "spy" : false,
                "stale-connection-checker-class-name" : null,
                "stale-connection-checker-properties" : null,
                "statistics-enabled" : false,
                "track-statements" : "NOWARN",
                "transaction-isolation" : null,
                "url-delimiter" : null,
                "url-selector-strategy-class-name" : null,
                "use-ccm" : true,
                "use-fast-fail" : false,
                "use-java-context" : true,
                "use-try-lock" : null,
                "user-name" : "sa",
                "valid-connection-checker-class-name" : null,
                "valid-connection-checker-properties" : null,
                "validate-on-match" : false,
                "connection-properties" : null,
                "statistics" : {
                    "jdbc" : {
                        "PreparedStatementCacheAccessCount" : "0",
                        "PreparedStatementCacheAddCount" : "0",
                        "PreparedStatementCacheCurrentSize" : "0",
                        "PreparedStatementCacheDeleteCount" : "0",
                        "PreparedStatementCacheHitCount" : "0",
                        "PreparedStatementCacheMissCount" : "0",
                        "statistics-enabled" : false
                    },
                    "pool" : {
                        "ActiveCount" : "0",
                        "AvailableCount" : "0",
                        "AverageBlockingTime" : "0",
                        "AverageCreationTime" : "0",
                        "CreatedCount" : "0",
                        "DestroyedCount" : "0",
                        "InUseCount" : "0",
                        "MaxCreationTime" : "0",
                        "MaxUsedCount" : "0",
                        "MaxWaitCount" : "0",
                        "MaxWaitTime" : "0",
                        "TimedOut" : "0",
                        "TotalBlockingTime" : "0",
                        "TotalCreationTime" : "0",
                        "statistics-enabled" : false
                    }
                }
            },
            "ExampleOracle" : {
                "allocation-retry" : null,
                "allocation-retry-wait-millis" : null,
                "allow-multiple-users" : false,
                "background-validation" : null,
                "background-validation-millis" : null,
                "blocking-timeout-wait-millis" : null,
                "check-valid-connection-sql" : null,
                "connectable" : false,
                "connection-url" : "jdbc:oracle:thin:@server:1524:ORASID",
                "datasource-class" : null,
                "driver-class" : null,
                "driver-name" : "oracle",
                "enabled" : true,
                "exception-sorter-class-name" : "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleExceptionSorter",
                "exception-sorter-properties" : null,
                "flush-strategy" : null,
                "idle-timeout-minutes" : null,
                "jndi-name" : "java:/jdbc/ExampleOracle",
                "jta" : true,
                "max-pool-size" : 30,
                "min-pool-size" : 10,
                "new-connection-sql" : null,
                "password" : "PASSWORD",
                "pool-prefill" : null,
                "pool-use-strict-min" : null,
                "prepared-statements-cache-size" : null,
                "query-timeout" : null,
                "reauth-plugin-class-name" : null,
                "reauth-plugin-properties" : null,
                "security-domain" : null,
                "set-tx-query-timeout" : false,
                "share-prepared-statements" : false,
                "spy" : false,
                "stale-connection-checker-class-name" : "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleStaleConnectionChecker",
                "stale-connection-checker-properties" : null,
                "statistics-enabled" : true,
                "track-statements" : "NOWARN",
                "transaction-isolation" : null,
                "url-delimiter" : null,
                "url-selector-strategy-class-name" : null,
                "use-ccm" : true,
                "use-fast-fail" : false,
                "use-java-context" : true,
                "use-try-lock" : null,
                "user-name" : "USEREXAMPLE",
                "valid-connection-checker-class-name" : "org.jboss.jca.adapters.jdbc.extensions.oracle.OracleValidConnectionChecker",
                "valid-connection-checker-properties" : null,
                "validate-on-match" : false,
                "connection-properties" : null,
                "statistics" : {
                    "jdbc" : {
                        "PreparedStatementCacheAccessCount" : "0",
                        "PreparedStatementCacheAddCount" : "0",
                        "PreparedStatementCacheCurrentSize" : "0",
                        "PreparedStatementCacheDeleteCount" : "0",
                        "PreparedStatementCacheHitCount" : "0",
                        "PreparedStatementCacheMissCount" : "0",
                        "statistics-enabled" : true
                    },
                    "pool" : {
                        "ActiveCount" : "3",
                        "AvailableCount" : "30",
                        "AverageBlockingTime" : "1",
                        "AverageCreationTime" : "2808",
                        "CreatedCount" : "3",
                        "DestroyedCount" : "0",
                        "InUseCount" : "0",
                        "MaxCreationTime" : "8340",
                        "MaxUsedCount" : "3",
                        "MaxWaitCount" : "0",
                        "MaxWaitTime" : "1",
                        "TimedOut" : "0",
                        "TotalBlockingTime" : "167",
                        "TotalCreationTime" : "8426",
                        "statistics-enabled" : true
                    }
                }
            }
        },
        "jdbc-driver" : {
            "h2" : {
                "deployment-name" : null,
                "driver-class-name" : null,
                "driver-datasource-class-name" : null,
                "driver-major-version" : null,
                "driver-minor-version" : null,
                "driver-module-name" : "com.h2database.h2",
                "driver-name" : "h2",
                "driver-xa-datasource-class-name" : "org.h2.jdbcx.JdbcDataSource",
                "jdbc-compliant" : null,
                "module-slot" : null,
                "xa-datasource-class" : null
            },
            "oracle" : {
                "deployment-name" : null,
                "driver-class-name" : "oracle.jdbc.driver.OracleDriver",
                "driver-datasource-class-name" : null,
                "driver-major-version" : null,
                "driver-minor-version" : null,
                "driver-module-name" : "com.liferay.portal",
                "driver-name" : "oracle",
                "driver-xa-datasource-class-name" : null,
                "jdbc-compliant" : null,
                "module-slot" : null,
                "xa-datasource-class" : null
            }
        },
        "xa-data-source" : null
    }
}
`

var jboss_jvm_out = `
{
    "outcome" : "success",
    "result" : {"type" : {
        "compilation" : {
            "name" : "HotSpot 64-Bit Tiered Compilers",
            "compilation-time-monitoring-supported" : true,
            "total-compilation-time" : 136234,
            "object-name" : "java.lang:type=Compilation"
        },
        "memory-manager" : {"name" : {
            "ParNew" : {
                "memory-pool-names" : [
                    "Par_Eden_Space",
                    "Par_Survivor_Space"
                ],
                "name" : "ParNew",
                "object-name" : "java.lang:type=MemoryManager,name=ParNew",
                "valid" : true
            },
            "CodeCacheManager" : {
                "memory-pool-names" : ["Code_Cache"],
                "name" : "CodeCacheManager",
                "object-name" : "java.lang:type=MemoryManager,name=CodeCacheManager",
                "valid" : true
            },
            "ConcurrentMarkSweep" : {
                "memory-pool-names" : [
                    "Par_Eden_Space",
                    "Par_Survivor_Space",
                    "CMS_Old_Gen",
                    "CMS_Perm_Gen"
                ],
                "name" : "ConcurrentMarkSweep",
                "object-name" : "java.lang:type=MemoryManager,name=ConcurrentMarkSweep",
                "valid" : true
            }
        }},
        "garbage-collector" : {"name" : {
            "ParNew" : {
                "collection-count" : 18,
                "collection-time" : 3259,
                "memory-pool-names" : [
                    "Par_Eden_Space",
                    "Par_Survivor_Space"
                ],
                "name" : "ParNew",
                "object-name" : "java.lang:type=GarbageCollector,name=ParNew",
                "valid" : true
            },
            "ConcurrentMarkSweep" : {
                "collection-count" : 1,
                "collection-time" : 703,
                "memory-pool-names" : [
                    "Par_Eden_Space",
                    "Par_Survivor_Space",
                    "CMS_Old_Gen",
                    "CMS_Perm_Gen"
                ],
                "name" : "ConcurrentMarkSweep",
                "object-name" : "java.lang:type=GarbageCollector,name=ConcurrentMarkSweep",
                "valid" : true
            }
        }},
        "memory" : {
            "heap-memory-usage" : {
                "init" : 8589934592,
                "used" : 5685415080,
                "committed" : 8589869056,
                "max" : 8589869056
            },
            "non-heap-memory-usage" : {
                "init" : 539426816,
                "used" : 390926664,
                "committed" : 572850176,
                "max" : 587202560
            },
            "object-name" : "java.lang:type=Memory",
            "object-pending-finalization-count" : 0,
            "verbose" : true
        },
        "threading" : {
            "all-thread-ids" : [
                520,
                519,
                517,
                516,
                515,
                514,
                513,
                509,
                511,
                510,
                508,
                507,
                506,
                505,
                504,
                503,
                502,
                501,
                500,
                499,
                498,
                497,
                496,
                495,
                494,
                493,
                492,
                491,
                481,
                480,
                479,
                478,
                477,
                474,
                473,
                472,
                470,
                469,
                468,
                467,
                466,
                465,
                464,
                463,
                462,
                461,
                460,
                459,
                458,
                457,
                456,
                455,
                454,
                453,
                452,
                451,
                450,
                449,
                448,
                447,
                446,
                445,
                444,
                443,
                442,
                441,
                440,
                439,
                438,
                437,
                436,
                435,
                434,
                433,
                432,
                431,
                430,
                428,
                427,
                425,
                423,
                422,
                421,
                420,
                419,
                418,
                417,
                416,
                415,
                414,
                413,
                412,
                411,
                410,
                409,
                408,
                407,
                405,
                404,
                403,
                402,
                401,
                400,
                399,
                398,
                397,
                396,
                395,
                393,
                392,
                391,
                389,
                388,
                387,
                386,
                385,
                384,
                383,
                382,
                381,
                380,
                379,
                377,
                376,
                375,
                374,
                373,
                372,
                371,
                370,
                369,
                368,
                367,
                366,
                365,
                364,
                359,
                358,
                357,
                356,
                355,
                354,
                353,
                352,
                351,
                350,
                347,
                346,
                345,
                344,
                343,
                342,
                341,
                338,
                337,
                336,
                329,
                328,
                327,
                324,
                318,
                317,
                314,
                313,
                312,
                311,
                310,
                309,
                308,
                307,
                306,
                305,
                304,
                303,
                302,
                301,
                300,
                299,
                298,
                297,
                296,
                295,
                294,
                293,
                292,
                291,
                290,
                289,
                288,
                287,
                286,
                285,
                284,
                283,
                282,
                281,
                280,
                279,
                278,
                277,
                276,
                275,
                274,
                273,
                272,
                271,
                270,
                269,
                268,
                267,
                266,
                264,
                262,
                261,
                260,
                258,
                257,
                256,
                255,
                254,
                253,
                251,
                249,
                246,
                245,
                244,
                243,
                242,
                241,
                240,
                239,
                238,
                237,
                236,
                235,
                234,
                233,
                232,
                231,
                230,
                229,
                228,
                227,
                226,
                225,
                224,
                223,
                158,
                157,
                155,
                154,
                152,
                151,
                148,
                147,
                146,
                145,
                143,
                142,
                141,
                140,
                139,
                138,
                137,
                136,
                135,
                134,
                133,
                132,
                131,
                130,
                129,
                128,
                127,
                126,
                125,
                124,
                122,
                121,
                119,
                118,
                116,
                115,
                114,
                113,
                112,
                111,
                110,
                109,
                106,
                108,
                107,
                103,
                102,
                101,
                100,
                99,
                96,
                93,
                92,
                91,
                90,
                59,
                58,
                57,
                56,
                55,
                23,
                22,
                20,
                19,
                18,
                17,
                16,
                15,
                14,
                13,
                9,
                5,
                3,
                2
            ],
            "thread-contention-monitoring-supported" : true,
            "thread-cpu-time-supported" : true,
            "current-thread-cpu-time-supported" : true,
            "object-monitor-usage-supported" : true,
            "synchronizer-usage-supported" : true,
            "thread-contention-monitoring-enabled" : false,
            "thread-cpu-time-enabled" : true,
            "thread-count" : 417,
            "peak-thread-count" : 428,
            "total-started-thread-count" : 8672,
            "daemon-thread-count" : 297,
            "current-thread-cpu-time" : 9260031,
            "current-thread-user-time" : 0,
            "object-name" : "java.lang:type=Threading"
        },
        "operating-system" : {
            "name" : "Linux",
            "arch" : "amd64",
            "version" : "2.6.32-642.3.1.el6.x86_64",
            "available-processors" : 8,
            "system-load-average" : 0.14,
            "object-name" : "java.lang:type=OperatingSystem"
        },
        "buffer-pool" : {"name" : {
            "direct" : {
                "count" : 1927,
                "memory-used" : 73406778,
                "name" : "direct",
                "object-name" : "java.nio:type=BufferPool,name=direct",
                "total-capacity" : 73406778
            },
            "mapped" : {
                "count" : 20,
                "memory-used" : 88102,
                "name" : "mapped",
                "object-name" : "java.nio:type=BufferPool,name=mapped",
                "total-capacity" : 88102
            }
        }},
        "memory-pool" : {"name" : {
            "Par_Eden_Space" : {
                "name" : "Par_Eden_Space",
                "type" : "HEAP",
                "valid" : true,
                "memory-manager-names" : [
                    "ConcurrentMarkSweep",
                    "ParNew"
                ],
                "usage-threshold-supported" : false,
                "collection-usage-threshold-supported" : true,
                "usage-threshold" : null,
                "collection-usage-threshold" : 0,
                "usage" : {
                    "init" : 4294836224,
                    "used" : 2629990272,
                    "committed" : 4294836224,
                    "max" : 4294836224
                },
                "peak-usage" : {
                    "init" : 4294836224,
                    "used" : 4294836224,
                    "committed" : 4294836224,
                    "max" : 4294836224
                },
                "usage-threshold-exceeded" : null,
                "usage-threshold-count" : null,
                "collection-usage-threshold-exceeded" : false,
                "collection-usage-threshold-count" : 0,
                "collection-usage" : {
                    "init" : 4294836224,
                    "used" : 0,
                    "committed" : 4294836224,
                    "max" : 4294836224
                },
                "object-name" : "java.lang:type=MemoryPool,name=\"Par Eden Space\""
            },
            "Code_Cache" : {
                "name" : "Code_Cache",
                "type" : "NON_HEAP",
                "valid" : true,
                "memory-manager-names" : ["CodeCacheManager"],
                "usage-threshold-supported" : true,
                "collection-usage-threshold-supported" : false,
                "usage-threshold" : 0,
                "collection-usage-threshold" : null,
                "usage" : {
                    "init" : 2555904,
                    "used" : 35578176,
                    "committed" : 35979264,
                    "max" : 50331648
                },
                "peak-usage" : {
                    "init" : 2555904,
                    "used" : 35583552,
                    "committed" : 35979264,
                    "max" : 50331648
                },
                "usage-threshold-exceeded" : false,
                "usage-threshold-count" : 0,
                "collection-usage-threshold-exceeded" : null,
                "collection-usage-threshold-count" : null,
                "collection-usage" : null,
                "object-name" : "java.lang:type=MemoryPool,name=\"Code Cache\""
            },
            "Par_Survivor_Space" : {
                "name" : "Par_Survivor_Space",
                "type" : "HEAP",
                "valid" : true,
                "memory-manager-names" : [
                    "ConcurrentMarkSweep",
                    "ParNew"
                ],
                "usage-threshold-supported" : false,
                "collection-usage-threshold-supported" : true,
                "usage-threshold" : null,
                "collection-usage-threshold" : 0,
                "usage" : {
                    "init" : 65536,
                    "used" : 0,
                    "committed" : 65536,
                    "max" : 65536
                },
                "peak-usage" : {
                    "init" : 65536,
                    "used" : 8208,
                    "committed" : 65536,
                    "max" : 65536
                },
                "usage-threshold-exceeded" : null,
                "usage-threshold-count" : null,
                "collection-usage-threshold-exceeded" : false,
                "collection-usage-threshold-count" : 0,
                "collection-usage" : {
                    "init" : 65536,
                    "used" : 0,
                    "committed" : 65536,
                    "max" : 65536
                },
                "object-name" : "java.lang:type=MemoryPool,name=\"Par Survivor Space\""
            },
            "CMS_Old_Gen" : {
                "name" : "CMS_Old_Gen",
                "type" : "HEAP",
                "valid" : true,
                "memory-manager-names" : ["ConcurrentMarkSweep"],
                "usage-threshold-supported" : true,
                "collection-usage-threshold-supported" : true,
                "usage-threshold" : 0,
                "collection-usage-threshold" : 0,
                "usage" : {
                    "init" : 4294967296,
                    "used" : 3054301888,
                    "committed" : 4294967296,
                    "max" : 4294967296
                },
                "peak-usage" : {
                    "init" : 4294967296,
                    "used" : 3054301888,
                    "committed" : 4294967296,
                    "max" : 4294967296
                },
                "usage-threshold-exceeded" : false,
                "usage-threshold-count" : 0,
                "collection-usage-threshold-exceeded" : false,
                "collection-usage-threshold-count" : 0,
                "collection-usage" : {
                    "init" : 4294967296,
                    "used" : 2123412320,
                    "committed" : 4294967296,
                    "max" : 4294967296
                },
                "object-name" : "java.lang:type=MemoryPool,name=\"CMS Old Gen\""
            },
            "CMS_Perm_Gen" : {
                "name" : "CMS_Perm_Gen",
                "type" : "NON_HEAP",
                "valid" : true,
                "memory-manager-names" : ["ConcurrentMarkSweep"],
                "usage-threshold-supported" : true,
                "collection-usage-threshold-supported" : true,
                "usage-threshold" : 0,
                "collection-usage-threshold" : 0,
                "usage" : {
                    "init" : 536870912,
                    "used" : 355346576,
                    "committed" : 536870912,
                    "max" : 536870912
                },
                "peak-usage" : {
                    "init" : 536870912,
                    "used" : 355346576,
                    "committed" : 536870912,
                    "max" : 536870912
                },
                "usage-threshold-exceeded" : false,
                "usage-threshold-count" : 0,
                "collection-usage-threshold-exceeded" : false,
                "collection-usage-threshold-count" : 0,
                "collection-usage" : {
                    "init" : 536870912,
                    "used" : 267982088,
                    "committed" : 536870912,
                    "max" : 536870912
                },
                "object-name" : "java.lang:type=MemoryPool,name=\"CMS Perm Gen\""
            }
        }},
        "runtime" : {
            "name" : "6026@server.domain.com",
            "vm-name" : "Java HotSpot(TM) 64-Bit Server VM",
            "vm-vendor" : "Oracle Corporation",
            "vm-version" : "24.75-b04",
            "spec-name" : "Java Virtual Machine Specification",
            "spec-vendor" : "Oracle Corporation",
            "spec-version" : "1.7",
            "management-spec-version" : "1.2",
            "class-path" : "/opt/jboss/jboss-modules.jar",
            "library-path" : "/usr/java/packages/lib/amd64:/usr/lib64:/lib64:/lib:/usr/lib",
            "boot-class-path-supported" : true,
            "boot-class-path" : "/opt/jboss/jdk1.7.0_75/jre/lib/resources.jar:/opt/jboss/jdk1.7.0_75/jre/lib/rt.jar:/opt/jboss/jdk1.7.0_75/jre/lib/jsse.jar:/opt/jboss/jdk1.7.0_75/jre/lib/jce.jar:/opt/jboss/jdk1.7.0_75/jre/lib/charsets.jar:/op/jdk1.7.0_75/jre/classes",
            "input-arguments" : [
                "-D[Standalone]",
                "-XX:+UseCompressedOops",
                "-verbose:gc",
                "-Xloggc:/opt/jboss/standalone/log/gc.log",
                "-XX:+PrintGCDetails",
                "-XX:+PrintGCDateStamps",
                "-XX:+UseGCLogFileRotation",
                "-XX:NumberOfGCLogFiles=5",
                "-XX:GCLogFileSize=3M",
                "-XX:-TraceClassUnloading",
                "-Djboss.modules.system.pkgs=org.jboss.byteman",
                "-Djava.awt.headless=true",
                "-Djboss.modules.policy-permissions=true",
                "-XX:NewSize=4096m",
                "-XX:MaxNewSize=4096m",
                "-Xms8g",
                "-Xmx8g",
                "-XX:PermSize=512m",
                "-XX:MaxPermSize=512m",
                "-XX:SurvivorRatio=65536",
                "-XX:TargetSurvivorRatio=0",
                "-XX:MaxTenuringThreshold=0",
                "-XX:+UseParNewGC",
                "-XX:ParallelGCThreads=8",
                "-XX:+UseConcMarkSweepGC",
                "-XX:+CMSParallelRemarkEnabled",
                "-XX:+CMSCompactWhenClearAllSoftRefs",
                "-XX:CMSInitiatingOccupancyFraction=85",
                "-XX:+CMSScavengeBeforeRemark",
                "-XX:+CMSConcurrentMTEnabled",
                "-XX:ParallelCMSThreads=1",
                "-XX:+UseLargePages",
                "-XX:LargePageSizeInBytes=2m",
                "-XX:+UseCompressedOops",
                "-XX:+DisableExplicitGC",
                "-XX:-UseBiasedLocking",
                "-XX:+BindGCTaskThreadsToCPUs",
                "-XX:+UseFastAccessorMethods",
                "-Dorg.jboss.boot.log.file=/opt/jboss/standalone/log/server.log",
                "-Dlogging.configuration=file:/opt/jboss/standalone/configuration/logging.properties"
            ],
            "start-time" : 1509961944445,
            "system-properties" : {
                "[Standalone]" : "",
                "awt.toolkit" : "sun.awt.X11.XToolkit",
                "base.path" : ".",
                "catalina.home" : "/opt/jboss/standalone/tmp",
                "com.sun.media.jai.disableMediaLib" : "true",
                "cookie.http.only.names.excludes" : "",
                "env.HOME" : "/home/jboss",
                "env.JAVA_HOME" : "/opt/jboss/jdk1.7.0_75",
                "env.JBOSS_HOME" : "/opt/jboss",
                "env.JBOSS_PIDFILE" : "/var/run/jboss-as/jboss-as-standalone.pid",
                "env.LANG" : "en_US.UTF-8",
                "env.LAUNCH_JBOSS_IN_BACKGROUND" : "1",
                "env.LOGNAME" : "jboss",
                "env.NLSPATH" : "/usr/dt/lib/nls/msg/%L/%N.cat",
                "env.PATH" : "/sbin:/usr/sbin:/bin:/usr/bin",
                "env.PWD" : "/opt/jboss",
                "env.SHELL" : "/bin/bash",
                "env.SHLVL" : "3",
                "env.TERM" : "xterm",
                "env.USER" : "jboss",
                "env.XFILESEARCHPATH" : "/usr/dt/app-defaults/%L/Dt",
                "env._" : "/opt/jboss/jdk1.7.0_75/bin/java",
                "file.encoding" : "UTF-8",
                "file.encoding.pkg" : "sun.io",
                "file.separator" : "/",
                "http.header.secure.x.content.type.options" : "true",
                "http.header.secure.x.frame.options" : "true",
                "http.header.secure.x.frame.options.255" : "/|SAMEORIGIN",
                "http.header.secure.x.xss.protection" : "1",
                "http.proxyHost" : "192.168.1.25",
                "http.proxyPort" : "3128",
                "https.proxyHost" : "192.168.1.25",
                "https.proxyPort" : "3128",
                "ical4j.compatibility.outlook" : "true",
                "ical4j.parsing.relaxed" : "true",
                "ical4j.unfolding.relaxed" : "true",
                "ical4j.validation.relaxed" : "true",
                "intraband.impl" : "",
                "intraband.timeout.default" : "10000",
                "intraband.welder.impl" : "",
                "java.awt.graphicsenv" : "sun.awt.X11GraphicsEnvironment",
                "java.awt.headless" : "true",
                "java.awt.printerjob" : "sun.print.PSPrinterJob",
                "java.class.path" : "/opt/jboss/jboss-modules.jar",
                "java.class.version" : "51.0",
                "java.endorsed.dirs" : "/opt/jboss/jdk1.7.0_75/jre/lib/endorsed",
                "java.ext.dirs" : "/opt/jboss/jdk1.7.0_75/jre/lib/ext:/usr/java/packages/lib/ext",
                "java.home" : "/opt/jboss/jdk1.7.0_75/jre",
                "java.io.tmpdir" : "/tmp",
                "java.library.path" : "/usr/java/packages/lib/amd64:/usr/lib64:/lib64:/lib:/usr/lib",
                "java.naming.factory.url.pkgs" : "org.jboss.as.naming.interfaces:org.jboss.ejb.client.naming",
                "java.protocol.handler.pkgs" : "org.jboss.net.protocol|org.jboss.vfs.protocol",
                "java.runtime.name" : "Java(TM) SE Runtime Environment",
                "java.runtime.version" : "1.7.0_75-b13",
                "java.specification.name" : "Java Platform API Specification",
                "java.specification.vendor" : "Oracle Corporation",
                "java.specification.version" : "1.7",
                "java.util.logging.manager" : "org.jboss.logmanager.LogManager",
                "java.vendor" : "Oracle Corporation",
                "java.vendor.url" : "http://java.oracle.com/",
                "java.vendor.url.bug" : "http://bugreport.sun.com/bugreport/",
                "java.version" : "1.7.0_75",
                "java.vm.info" : "mixed mode",
                "java.vm.name" : "Java HotSpot(TM) 64-Bit Server VM",
                "java.vm.specification.name" : "Java Virtual Machine Specification",
                "java.vm.specification.vendor" : "Oracle Corporation",
                "java.vm.specification.version" : "1.7",
                "java.vm.vendor" : "Oracle Corporation",
                "java.vm.version" : "24.75-b04",
                "javax.management.builder.initial" : "org.jboss.as.jmx.PluggableMBeanServerBuilder",
                "javax.xml.datatype.DatatypeFactory" : "__redirected.__DatatypeFactory",
                "javax.xml.parsers.DocumentBuilderFactory" : "__redirected.__DocumentBuilderFactory",
                "javax.xml.parsers.SAXParserFactory" : "__redirected.__SAXParserFactory",
                "javax.xml.stream.XMLEventFactory" : "__redirected.__XMLEventFactory",
                "javax.xml.stream.XMLInputFactory" : "__redirected.__XMLInputFactory",
                "javax.xml.stream.XMLOutputFactory" : "__redirected.__XMLOutputFactory",
                "javax.xml.transform.TransformerFactory" : "__redirected.__TransformerFactory",
                "javax.xml.validation.SchemaFactory:http://www.w3.org/2001/XMLSchema" : "__redirected.__SchemaFactory",
                "javax.xml.xpath.XPathFactory:http://java.sun.com/jaxp/xpath/dom" : "__redirected.__XPathFactory",
                "jboss.bind.address.management" : "172.25.55.110",
                "jboss.home.dir" : "/opt/jboss",
                "jboss.host.name" : "server",
                "jboss.modules.dir" : "/opt/jboss/modules",
                "jboss.modules.policy-permissions" : "true",
                "jboss.modules.system.pkgs" : "org.jboss.byteman",
                "jboss.node.name" : "server",
                "jboss.qualified.host.name" : "server.domain.com",
                "jboss.server.base.dir" : "/opt/jboss/standalone",
                "jboss.server.config.dir" : "/opt/jboss/standalone/configuration",
                "jboss.server.data.dir" : "/opt/jboss/standalone/data",
                "jboss.server.deploy.dir" : "/opt/jboss/standalone/data/content",
                "jboss.server.log.dir" : "/opt/jboss/standalone/log",
                "jboss.server.name" : "server",
                "jboss.server.persist.config" : "true",
                "jboss.server.temp.dir" : "/opt/jboss/standalone/tmp",
                "jgroups.bind_addr" : "172.25.55.110",
                "jgroups.bind_interface" : "eth3",
                "jgroups.logging.log_factory_class" : "org.jboss.as.clustering.jgroups.LogFactory",
                "jgroups.mping.ip_ttl" : "8",
                "jgroups.mping.mcast_addr" : "239.255.10.18",
                "jgroups.mping.mcast_port" : "23338",
                "jruby.native.enabled" : "false",
                "line.separator" : "\n",
                "log.sanitizer.enabled" : "true",
                "log.sanitizer.escape.html.enabled" : "false",
                "log.sanitizer.replacement.character" : "95",
                "log.sanitizer.whitelist.characters" : "9,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,8,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,103,104,105,10118,119,120,121,122,123,124,125,126",
                "log4j.configure.on.startup" : "true",
                "logging.configuration" : "file:/opt/jboss/standalone/configuration/logging.properties",
                "module.path" : "/opt/jboss/modules",
                "net.sf.ehcache.skipUpdateCheck" : "true",
                "org.apache.catalina.connector.URI_ENCODING" : "UTF-8",
                "org.apache.catalina.connector.USE_BODY_ENCODING_FOR_QUERY_STRING" : "true",
                "org.apache.xml.security.ignoreLineBreaks" : "true",
                "org.jboss.boot.log.file" : "/opt/jboss/standalone/log/server.log",
                "org.jboss.com.sun.CORBA.ORBUseDynamicStub" : "true",
                "org.jboss.resolver.warning" : "true",
                "org.jboss.security.context.ThreadLocal" : "true",
                "org.omg.CORBA.ORBClass" : "org.jacorb.orb.ORB",
                "org.omg.CORBA.ORBSingletonClass" : "org.jacorb.orb.ORBSingleton",
                "org.quartz.threadPool.makeThreadsDaemons" : "true",
                "org.terracotta.quartz.skipUpdateCheck" : "true",
                "org.xml.sax.driver" : "__redirected.__XMLReaderFactory",
                "os.arch" : "amd64",
                "os.name" : "Linux",
                "os.version" : "2.6.32-642.3.1.el6.x86_64",
                "path.separator" : ":",
                "sun.arch.data.model" : "64",
                "sun.boot.class.path" : "/opt/jboss/jdk1.7.0_75/jre/lib/resources.jar:/opt/jboss/jdk1.7.0_75/jre/lib/rt.jar:/opt/jboss/jdk1.7.0_75/jre/lib/jsse.jar:/opt/jboss/jdk1.7.0_75/jre/lib/jce.jar:/opt/jboss/jdk1.7.0_75/jre/lib/charsetspt/jboss/jdk1.7.0_75/jre/classes",
                "sun.boot.library.path" : "/opt/jboss/jdk1.7.0_75/jre/lib/amd64",
                "sun.cpu.endian" : "little",
                "sun.cpu.isalist" : "",
                "sun.font.fontmanager" : "sun.awt.X11FontManager",
                "sun.io.unicode.encoding" : "UnicodeLittle",
                "sun.java.command" : "/opt/jboss/jboss-modules.jar -mp /opt/jboss/modules -jaxpmodule javax.xml.jaxp-prov/opt/jboss -Djboss.server.base.dir=/opt/jboss/standalone -c standalone-full-ha.xml -bmanagement 172.25.55.110",
                "sun.java.launcher" : "SUN_STANDARD",
                "sun.jnu.encoding" : "UTF-8",
                "sun.management.compiler" : "HotSpot 64-Bit Tiered Compilers",
                "sun.nio.ch.bugLevel" : "",
                "sun.os.patch.level" : "unknown",
                "user.country" : "US",
                "user.dir" : "/opt/jboss",
                "user.home" : "/home/jboss",
                "user.language" : "en",
                "user.name" : "jboss",
                "user.timezone" : "UTC"
            },
            "uptime" : 110195561,
            "object-name" : "java.lang:type=Runtime"
        },
        "class-loading" : {
            "loaded-class-count" : 51550,
            "object-name" : "java.lang:type=ClassLoading",
            "total-loaded-class-count" : 51550,
            "unloaded-class-count" : 0,
            "verbose" : false
        }
    }}
}
`

var jboss_webcon_http = `
{
   "outcome" : "success",
   "result" : {
       "bytesReceived" : "0",
       "bytesSent" : "0",
       "enable-lookups" : false,
       "enabled" : true,
       "errorCount" : "0",
       "executor" : null,
       "max-connections" : null,
       "max-post-size" : 2097152,
       "max-save-post-size" : 4096,
       "maxTime" : "0",
       "name" : "http",
       "processingTime" : "0",
       "protocol" : "HTTP/1.1",
       "proxy-binding" : null,
       "proxy-name" : null,
       "proxy-port" : null,
       "redirect-binding" : null,
       "redirect-port" : 443,
       "requestCount" : "0",
       "scheme" : "http",
       "secure" : false,
       "socket-binding" : "http",
       "virtual-server" : null,
       "configuration" : null
   }
}
`

var jboss_deployment_list = `
{
    "outcome" : "success",
    "result" : [
        "sample.war",
        "HelloWorld.ear"
    ]
}
`

var jboss_web_app_war = `
{
    "outcome" : "success",
    "result" : {
        "content" : [{"hash" : {
            "BYTES_VALUE" : "gPUFOxZsadgWl7ohETxnP4NyrKA="
        }}],
        "enabled" : true,
        "name" : "sample.war",
        "persistent" : true,
        "runtime-name" : "sample.war",
        "status" : "OK",
        "subdeployment" : null,
        "subsystem" : {"web" : {
            "active-sessions" : 0,
            "context-root" : "/sample",
            "duplicated-session-ids" : 0,
            "expired-sessions" : 0,
            "max-active-sessions" : 0,
            "rejected-sessions" : 0,
            "session-avg-alive-time" : 0,
            "session-max-alive-time" : 0,
            "sessions-created" : 0,
            "virtual-host" : "default-host",
            "servlet" : {"HelloServlet" : {
                "load-time" : 0,
                "maxTime" : 0,
                "min-time" : 9223372036854775807,
                "processingTime" : 0,
                "requestCount" : 0,
                "servlet-class" : "mypackage.Hello",
                "servlet-name" : "HelloServlet"
            }}
        }}
    }
}
`

var jboss_web_app_ear = `
{
    "outcome" : "success",
    "result" : {
        "content" : [{"hash" : {
            "BYTES_VALUE" : "L00a/u5Z7U2/rsIT5BsSijC8Usg="
        }}],
        "enabled" : true,
        "name" : "HelloWorld.ear",
        "persistent" : true,
        "runtime-name" : "HelloWorld.ear",
        "status" : "OK",
        "subdeployment" : {
            "web.war" : {"subsystem" : {"web" : {
                "active-sessions" : 0,
                "context-root" : "/HelloWorld",
                "duplicated-session-ids" : 0,
                "expired-sessions" : 0,
                "max-active-sessions" : 0,
                "rejected-sessions" : 0,
                "session-avg-alive-time" : 0,
                "session-max-alive-time" : 0,
                "sessions-created" : 0,
                "virtual-host" : "default-host",
                "servlet" : {"HelloWorldServlet" : {
                    "load-time" : 0,
                    "maxTime" : 0,
                    "min-time" : 9223372036854775807,
                    "processingTime" : 0,
                    "requestCount" : 0,
                    "servlet-class" : "eu.glotzich.j2ee.common.HelloWorldServlet",
                    "servlet-name" : "HelloWorldServlet"
                }}
            }}},
            "common.jar" : {"subsystem" : null},
            "ejb.jar" : {"subsystem" : {"ejb3" : {
                "entity-bean" : null,
                "message-driven-bean" : null,
                "singleton-bean" : null,
                "stateful-session-bean" : null,
                "stateless-session-bean" : {"MyEJB" : {
                    "component-class-name" : "MyEJB",
                    "declared-roles" : [],
                    "execution-time" : 0,
                    "invocations" : 0,
                    "methods" : {},
                    "peak-concurrent-invocations" : 0,
                    "pool-available-count" : 20,
                    "pool-create-count" : 0,
                    "pool-current-size" : 0,
                    "pool-max-size" : 20,
                    "pool-name" : "slsb-strict-max-pool",
                    "pool-remove-count" : 0,
                    "run-as-role" : null,
                    "security-domain" : "other",
                    "timers" : [],
                    "wait-time" : 0,
                    "service" : null
                }}
            }}}
        },
        "subsystem" : null
    }
}
`
var jboss_jms_out = `
{
    "outcome" : "success",
    "result" : {
        "DLQ" : {
            "consumer-count" : 0,
            "dead-letter-address" : "jms.queue.DLQ",
            "delivering-count" : 0,
            "durable" : true,
            "entries" : ["java:/jms/queue/DLQ"],
            "expiry-address" : "jms.queue.ExpiryQueue",
            "message-count" : 0,
            "messages-added" : 0,
            "paused" : false,
            "queue-address" : "jms.queue.DLQ",
            "scheduled-count" : 0,
            "selector" : null,
            "temporary" : false
        },
        "ExpiryQueue" : {
            "consumer-count" : 0,
            "dead-letter-address" : "jms.queue.DLQ",
            "delivering-count" : 0,
            "durable" : true,
            "entries" : ["java:/jms/queue/ExpiryQueue"],
            "expiry-address" : "jms.queue.ExpiryQueue",
            "message-count" : 0,
            "messages-added" : 0,
            "paused" : false,
            "queue-address" : "jms.queue.ExpiryQueue",
            "scheduled-count" : 0,
            "selector" : null,
            "temporary" : false
        }
    }
}

`

type BodyContent struct {
	Operation      string                 `json:"operation"`
	Name           string                 `json:"name"`
	IncludeRuntime string                 `json:"include-runtime"`
	AttributesOnly string                 `json:"attributes-only"`
	ChildType      string                 `json:"child-type"`
	RecursiveDepth int                    `json:"recursive-depth"`
	Recursive      string                 `json:"recursive"`
	Address        []map[string]interface{} `json:"address"`
	JsonPretty     int                    `json:"json.pretty"`
}

func testJBossServer(t *testing.T, eap7 bool) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Println("--------------------INIT REQUEST------------------------------------------")
		w.WriteHeader(http.StatusOK)
		decoder := json.NewDecoder(r.Body)
		var b BodyContent
		err := decoder.Decode(&b)
		if err != nil {
			fmt.Printf("ERROR DECODE: %s\n", err)
		}

//		fmt.Printf("REQUEST:%+v\n", r.Body)
		fmt.Printf("BODYCONTENT:%+v\n", b)
		if b.Operation == "read-resource" {
			if b.AttributesOnly == "true" {
				if eap7 {
					fmt.Fprintln(w, jboss_exec_mode_out)
				} else {
					fmt.Fprintln(w, jboss_exec_mode_out_eap6)
				}
				return
			}
			if v, ok := b.Address[0]["core-service"]; ok {
				if v == "platform-mbean" {
					fmt.Fprintln(w, jboss_jvm_out)
					return
				}
			}
			if v, ok := b.Address[0]["subsystem"]; ok {
				if v == "web" && b.Address[1]["connector"] == "http" {
					fmt.Fprintln(w, jboss_webcon_http)
					return
				}
				if v == "datasources" {
					fmt.Fprintln(w, jboss_database_out)
					return
				}
			}
			if v, ok := b.Address[0]["deployment"]; ok {
				switch v {
				case "sample.war":
					fmt.Fprintln(w, jboss_web_app_war)
					return
				case "HelloWorld.ear":
					fmt.Fprintln(w, jboss_web_app_ear)
					return
				}
			}
		}
		if b.Operation == "read-children-names" {
			switch b.ChildType {
			case "host":
				fmt.Fprintln(w, jboss_host_list)
				return
			case "server":
			case "deployment":
				fmt.Fprintln(w, jboss_deployment_list)
				return

			}
		}
		if b.Operation == "read-children-resources" {
			switch b.ChildType {
			case "jms-queue":
				fmt.Fprintln(w, jboss_jms_out)
				return
			case "jms-topic":
				fmt.Fprintln(w, jboss_jms_out)
				return
			}
		}
	}))

	return ts
}

func TestHTTPJboss(t *testing.T) {

	ts := testJBossServer(t, false)
	defer ts.Close()
	j := JBoss{
		Servers:       []string{ts.URL},
		Username:      "",
		Password:      "",
		Authorization: "digest",
		Metrics: []string{
			"jvm",
			"web",
			"deployment",
			"database",
			"jms",
		},
		client: &RealHTTPClient{},
	}
	//var acc testutil.Accumulator
	acc := new(testutil.Accumulator)
	err := acc.GatherError(j.Gather)
	require.NoError(t, err)
	//TEST JVM
	fields_jvm := map[string]interface{}{
		"thread-count":              float64(417),
		"peak-thread-count":         float64(428),
		"daemon-thread-count":       float64(297),
		"ConcurrentMarkSweep_count": float64(1),
		"ConcurrentMarkSweep_time":  float64(703),
		"ParNew_count":              float64(18),
		"ParNew_time":               float64(3259),
		"heap_committed":            float64(8.589869056e+09),
		"heap_init":                 float64(8.589934592e+09),
		"heap_max":                  float64(8.589869056e+09),
		"heap_used":                 float64(5.68541508e+09),
		"nonheap_committed":         float64(5.72850176e+08),
		"nonheap_init":              float64(5.39426816e+08),
		"nonheap_max":               float64(5.8720256e+08),
		"nonheap_used":              float64(3.90926664e+08),
	}
	acc.AssertContainsFields(t, "jboss_jvm", fields_jvm)

	//TEST WEB CONNETOR
	fields_web := map[string]interface{}{
		"bytesReceived":  float64(0),
		"bytesSent":      float64(0),
		"errorCount":     float64(0),
		"maxTime":        float64(0),
		"processingTime": float64(0),
		"requestCount":   float64(0),
	}
	acc.AssertContainsFields(t, "jboss_web", fields_web)

	//TEST WEBAPP WAR
	fields_web_app_sample_war := map[string]interface{}{
		"active-sessions":     float64(0),
		"expired-sessions":    float64(0),
		"max-active-sessions": float64(0),
		"sessions-created":    float64(0),
	}
	tags_web_app_sample_war := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "sample.war",
		"context-root": "/sample",
		"runtime_name": "sample.war",
	}

	acc.AssertContainsTaggedFields(t, "jboss_web_app", fields_web_app_sample_war, tags_web_app_sample_war)

	//TEST WEBAPP EAR
	fields_web_app_sample_ear := map[string]interface{}{
		"active-sessions":     float64(0),
		"expired-sessions":    float64(0),
		"max-active-sessions": float64(0),
		"sessions-created":    float64(0),
	}
	tags_web_app_sample_ear := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "web.war",
		"context-root": "/HelloWorld",
		"runtime_name": "HelloWorld.ear",
	}

	acc.AssertContainsTaggedFields(t, "jboss_web_app", fields_web_app_sample_ear, tags_web_app_sample_ear)

	//TEST DATASOURCES
	fields_datasource_exampleDS := map[string]interface{}{
		"in-use-count":    int64(0),
		"active-count":    int64(0),
		"available-count": int64(0),
	}
	fields_datasource_exampleOracle := map[string]interface{}{
		"in-use-count":    int64(0),
		"active-count":    int64(3),
		"available-count": int64(30),
	}

	tags_datasource_exampleDS := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExampleDS",
	}
	tags_datasource_exampleOracle := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExampleOracle",
	}

	acc.AssertContainsTaggedFields(t, "jboss_database", fields_datasource_exampleDS, tags_datasource_exampleDS)
	acc.AssertContainsTaggedFields(t, "jboss_database", fields_datasource_exampleOracle, tags_datasource_exampleOracle)

	// TEST JMS
	fields_jms_DLQ := map[string]interface{}{
		"message-count":   float64(0),
		"messages-added":  float64(0),
		"consumer-count":  float64(0),
		"scheduled-count": float64(0),
	}
	fields_jms_ExpiryQueue := map[string]interface{}{
		"message-count":   float64(0),
		"messages-added":  float64(0),
		"consumer-count":  float64(0),
		"scheduled-count": float64(0),
	}

	tags_jms_DLQ := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "DLQ",
	}
	tags_jms_ExpiryQueue := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExpiryQueue",
	}

	acc.AssertContainsTaggedFields(t, "jboss_jms", fields_jms_DLQ, tags_jms_DLQ)
	acc.AssertContainsTaggedFields(t, "jboss_jms", fields_jms_ExpiryQueue, tags_jms_ExpiryQueue)
}

func TestHTTPJbossEAP6Domain(t *testing.T) {

	ts := testJBossServer(t, false)
	defer ts.Close()
	j := JBoss{
		Servers:       []string{ts.URL},
		Username:      "",
		Password:      "",
		Metrics: []string{
			"jvm",
			"web",
			"deployment",
			"database",
			"jms",
		},
		client: &RealHTTPClient{},
	}
	//var acc testutil.Accumulator
	acc := new(testutil.Accumulator)
	err := acc.GatherError(j.Gather)
	require.NoError(t, err)
}

