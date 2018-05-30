package solr

const statusResponse = `
{
  "status": {
    "core1": {
      "index": {
        "size": "1.66 GB",
        "sizeInBytes": 1784635686,
        "lastModified": "2017-01-14T10:30:07.419Z",
        "userData": {
          "commitTimeMSec": "1484389807419"
        },
        "numDocs": 7517488,
        "maxDoc": 7620303,
        "deletedDocs": 102815,
        "version": 267485,
        "segmentCount": 21,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core1/index.20160607000000124 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@646d42ce"
      },
      "name": "core1",
      "isDefaultCore": false,
      "instanceDir": "solr/core1/",
      "dataDir": "/srv/solr-core1/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-12-20T18:41:10.449Z",
      "uptime": 2314746645
    },
    "main": {
      "index": {
        "size": "230.5 GB",
        "sizeInBytes": 247497521642,
        "lastModified": "2017-01-16T11:59:18.189Z",
        "userData": {
          "commitTimeMSec": "1484567958189"
        },
        "numDocs": 168943425,
        "maxDoc": 169562700,
        "deletedDocs": 619275,
        "version": 70688464,
        "segmentCount": 33,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr/index.20161110090000012 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@15088f05"
      },
      "name": "main",
      "isDefaultCore": true,
      "instanceDir": "solr/main/",
      "dataDir": "/srv/solr/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-12-20T18:41:10.796Z",
      "uptime": 2314746294
    }
  },
  "initFailures": {},
  "defaultCoreName": "main",
  "responseHeader": {
    "QTime": 13,
    "status": 0
  }
}
`

const mBeansMainResponse = `{
  "solr-mbeans": [
    "CORE",
    {
      "core": {
        "stats": {
          "aliases": [
            "main"
          ],
          "indexDir": "/srv/solr/index.20161110090000012",
          "refCount": 2,
          "startTime": "2016-12-20T18:41:10.796Z",
          "coreName": "main"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/SolrCore.java $",
        "description": "SolrCore",
        "version": "1.0",
        "class": "main"
      },
      "searcher": {
        "stats": {
          "warmupTime": 0,
          "registeredAt": "2017-01-17T09:00:03.303Z",
          "openedAt": "2017-01-17T09:00:03.301Z",
          "searcherName": "Searcher@62d3fac7 main",
          "caching": true,
          "numDocs": 168962621,
          "maxDoc": 169647870,
          "deletedDocs": 685249,
          "reader": "StandardDirectoryReader(segments_jwq89:70709031:nrt _dp3n5(4.3.1):C168268689/592191 _dph0g(4.3.1):C311982/51776 _dpz3u(4.3.1):C589116/12754 _dpsbv(4.3.1):C262008/22358 _dq1e0(4.3.1):C104991/772 _dpy04(4.3.1):C24856/1389 _dq029(4.3.1):C42680/1406 _dq0rr(4.3.1):C5064/581 _dq13q(4.3.1):C4322/574 _dq165(4.3.1):C4679/364 _dq1kt(4.3.1):C8124/196 _dq1ta(4.3.1):C8138/152 _dq1x7(4.3.1):C3842/76 _dq212(4.3.1):C4934/111 _dq1wi(4.3.1):C778/145 _dq20q(4.3.1):C805/92 _dq20g(4.3.1):C1183/96 _dq21g(4.3.1):C257/58 _dq20y(4.3.1):C159/19 _dq213(4.3.1):C108/17 _dq218(4.3.1):C89/9 _dq21a(4.3.1):C213/20 _dq21d(4.3.1):C100/10 _dq21f(4.3.1):C214/16 _dq21j(4.3.1):C198/17 _dq21m(4.3.1):C112/2 _dq21n(4.3.1):C105/46 _dq21o(4.3.1):C124/2)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr/index.20161110090000012 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@15088f05",
          "indexVersion": 70709031
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "description": "index searcher",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrIndexSearcher"
      },
      "Searcher@62d3fac7 main": {
        "stats": {
          "warmupTime": 0,
          "registeredAt": "2017-01-17T09:00:03.303Z",
          "openedAt": "2017-01-17T09:00:03.301Z",
          "searcherName": "Searcher@62d3fac7 main",
          "caching": true,
          "numDocs": 168962621,
          "maxDoc": 169647870,
          "deletedDocs": 685249,
          "reader": "StandardDirectoryReader(segments_jwq89:70709031:nrt _dp3n5(4.3.1):C168268689/592191 _dph0g(4.3.1):C311982/51776 _dpz3u(4.3.1):C589116/12754 _dpsbv(4.3.1):C262008/22358 _dq1e0(4.3.1):C104991/772 _dpy04(4.3.1):C24856/1389 _dq029(4.3.1):C42680/1406 _dq0rr(4.3.1):C5064/581 _dq13q(4.3.1):C4322/574 _dq165(4.3.1):C4679/364 _dq1kt(4.3.1):C8124/196 _dq1ta(4.3.1):C8138/152 _dq1x7(4.3.1):C3842/76 _dq212(4.3.1):C4934/111 _dq1wi(4.3.1):C778/145 _dq20q(4.3.1):C805/92 _dq20g(4.3.1):C1183/96 _dq21g(4.3.1):C257/58 _dq20y(4.3.1):C159/19 _dq213(4.3.1):C108/17 _dq218(4.3.1):C89/9 _dq21a(4.3.1):C213/20 _dq21d(4.3.1):C100/10 _dq21f(4.3.1):C214/16 _dq21j(4.3.1):C198/17 _dq21m(4.3.1):C112/2 _dq21n(4.3.1):C105/46 _dq21o(4.3.1):C124/2)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr/index.20161110090000012 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@15088f05",
          "indexVersion": 70709031
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "description": "index searcher",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrIndexSearcher"
      }
    },
    "QUERYHANDLER",
    {
      "org.apache.solr.handler.CSVRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270814,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.CSVRequestHandler"
      },
      "/admin/": {
        "stats": null,
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "description": "Register Standard Admin Handlers",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.AdminHandlers"
      },
      "/admin/mbeans": {
        "stats": {
          "999thPcRequestTime": 129.19901009200038,
          "99thPcRequestTime": 11.944256130000017,
          "95thPcRequestTime": 9.10313265,
          "75thPcRequestTime": 7.423904,
          "medianRequestTime": 0.046796000000000004,
          "avgTimePerRequest": 2.0964317122172575,
          "handlerStart": 1482259271568,
          "requests": 230953,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 484175.0968,
          "avgRequestsPerSecond": 0.0967113175627352,
          "5minRateReqsPerSecond": 0.5543011916891444,
          "15minRateReqsPerSecond": 0.5409225999558686
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SolrInfoMBeanHandler.java $",
        "description": "Get Info (and statistics) for registered SolrInfoMBeans",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.SolrInfoMBeanHandler"
      },
      "/debug/dump": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270816,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "description": "Dump handler (debug)",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.DumpRequestHandler"
      },
      "/admin/logging": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259271569,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LoggingHandler.java $",
        "description": "Logging Handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.LoggingHandler"
      },
      "/admin/plugins": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259271568,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PluginInfoHandler.java $",
        "description": "Registry",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.PluginInfoHandler"
      },
      "/admin/system": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259271568,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SystemInfoHandler.java $",
        "description": "Get System Info",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.SystemInfoHandler"
      },
      "/select": {
        "stats": {
          "999thPcRequestTime": 145.3845197990004,
          "99thPcRequestTime": 1.404113640000005,
          "95thPcRequestTime": 0.21192269999999988,
          "75thPcRequestTime": 0.12097,
          "medianRequestTime": 0.116272,
          "avgTimePerRequest": 3.0322834013981987,
          "handlerStart": 1482259270810,
          "requests": 729510,
          "errors": 0,
          "timeouts": 9,
          "totalTime": 2212081.064154,
          "avgRequestsPerSecond": 0.3054827447483145,
          "5minRateReqsPerSecond": 0.32614693588972216,
          "15minRateReqsPerSecond": 0.3320899738059959
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,spellcheck,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/tvrh": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "org.apache.solr.handler.component.SearchHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270810,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/admin/luke": {
        "stats": {
          "999thPcRequestTime": 0.4085,
          "99thPcRequestTime": 0.4085,
          "95thPcRequestTime": 0.4085,
          "75thPcRequestTime": 0.4085,
          "medianRequestTime": 0.31491,
          "avgTimePerRequest": 0.3105803333333333,
          "handlerStart": 1482259271568,
          "requests": 3,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0.931741,
          "avgRequestsPerSecond": 1.256252178648736e-06,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 1.387107477978473e-152
        },
        "docs": [
          "http://wiki.apache.org/solr/LukeRequestHandler"
        ],
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LukeRequestHandler.java $",
        "description": "Lucene Index Browser.  Inspired and modeled after Luke: http://www.getopt.org/luke/",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.LukeRequestHandler"
      },
      "/update/json": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270814,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "description": "Add documents with JSON",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler"
      },
      "org.apache.solr.handler.admin.AdminHandlers": {
        "stats": null,
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "description": "Register Standard Admin Handlers",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.AdminHandlers"
      },
      "org.apache.solr.handler.ReplicationHandler": {
        "stats": {
          "lastCycleBytesDownloaded": "240149545",
          "timesIndexReplicated": "2697",
          "timesFailed": "1",
          "replicationFailedAt": "Mon Jun 06 11:55:11 UTC 2016",
          "indexReplicatedAt": "Tue Jan 17 09:00:03 UTC 2017",
          "previousCycleTimeInSeconds": "3",
          "isReplicating": "false",
          "isPollingDisabled": "false",
          "pollInterval": "03:00:00",
          "masterUrl": "http://solr-s1:8983/solr/main",
          "isSlave": "true",
          "isMaster": "false",
          "indexPath": "/srv/solr/index.20161110090000012",
          "generation": 33439689,
          "15minRateReqsPerSecond": 4.3340312709959365e-152,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "avgRequestsPerSecond": 2.0937529718964145e-06,
          "totalTime": 49.630943,
          "timeouts": 0,
          "errors": 0,
          "requests": 5,
          "handlerStart": 1482259270817,
          "avgTimePerRequest": 9.9261886,
          "medianRequestTime": 8.547115,
          "75thPcRequestTime": 12.1924675,
          "95thPcRequestTime": 15.377019,
          "99thPcRequestTime": 15.377019,
          "999thPcRequestTime": 15.377019,
          "indexSize": "229.77 GB",
          "indexVersion": 1484643564822
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.ReplicationHandler"
      },
      "org.apache.solr.handler.JsonUpdateRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270814,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "description": "Add documents with JSON",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler"
      },
      "org.apache.solr.handler.DumpRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270816,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "description": "Dump handler (debug)",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.DumpRequestHandler"
      },
      "org.apache.solr.handler.RealTimeGetHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270810,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "description": "The realtime get handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.RealTimeGetHandler"
      },
      "/get": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270810,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "description": "The realtime get handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.RealTimeGetHandler"
      },
      "/admin/properties": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259271568,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PropertiesRequestHandler.java $",
        "description": "Get System Properties",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.PropertiesRequestHandler"
      },
      "/query": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270810,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/admin/threads": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259271568,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ThreadDumpHandler.java $",
        "description": "Thread Dump",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.ThreadDumpHandler"
      },
      "/analysis/field": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.FieldAnalysisRequestHandler]",
        "version": null,
        "class": "Lazy[solr.FieldAnalysisRequestHandler]"
      },
      "org.apache.solr.handler.PingRequestHandler": {
        "stats": {
          "999thPcRequestTime": 41.331967987,
          "99thPcRequestTime": 5.392157590000151,
          "95thPcRequestTime": 0.4901222999999999,
          "75thPcRequestTime": 0.357574,
          "medianRequestTime": 0.3474125,
          "avgTimePerRequest": 0.7749319095595372,
          "handlerStart": 1482259270816,
          "requests": 477021,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 369658.79443,
          "avgRequestsPerSecond": 0.19975282688134827,
          "5minRateReqsPerSecond": 0.2000000000000008,
          "15minRateReqsPerSecond": 0.20000000000000234
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "description": "Reports application health to a load-balancer",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.PingRequestHandler"
      },
      "/analysis/document": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "version": null,
        "class": "Lazy[solr.DocumentAnalysisRequestHandler]"
      },
      "/spell": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "/update/csv": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270814,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.CSVRequestHandler"
      },
      "org.apache.solr.handler.UpdateRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270811,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.UpdateRequestHandler"
      },
      "/replication": {
        "stats": {
          "lastCycleBytesDownloaded": "240149545",
          "timesIndexReplicated": "2697",
          "timesFailed": "1",
          "replicationFailedAt": "Mon Jun 06 11:55:11 UTC 2016",
          "indexReplicatedAt": "Tue Jan 17 09:00:03 UTC 2017",
          "previousCycleTimeInSeconds": "3",
          "isReplicating": "false",
          "isPollingDisabled": "false",
          "pollInterval": "03:00:00",
          "masterUrl": "http://solr-s1:8983/solr/main",
          "isSlave": "true",
          "isMaster": "false",
          "indexPath": "/srv/solr/index.20161110090000012",
          "generation": 33439689,
          "15minRateReqsPerSecond": 4.3340312709959365e-152,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "avgRequestsPerSecond": 2.0937529681743243e-06,
          "totalTime": 49.630943,
          "timeouts": 0,
          "errors": 0,
          "requests": 5,
          "handlerStart": 1482259270817,
          "avgTimePerRequest": 9.9261886,
          "medianRequestTime": 8.547115,
          "75thPcRequestTime": 12.1924675,
          "95thPcRequestTime": 15.377019,
          "99thPcRequestTime": 15.377019,
          "999thPcRequestTime": 15.377019,
          "indexSize": "229.77 GB",
          "indexVersion": 1484643564822
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.ReplicationHandler"
      },
      "/admin/ping": {
        "stats": {
          "999thPcRequestTime": 41.331967987,
          "99thPcRequestTime": 5.392157590000151,
          "95thPcRequestTime": 0.4901222999999999,
          "75thPcRequestTime": 0.357574,
          "medianRequestTime": 0.3474125,
          "avgTimePerRequest": 0.7749319095595372,
          "handlerStart": 1482259270816,
          "requests": 477021,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 369658.79443,
          "avgRequestsPerSecond": 0.19975282659471533,
          "5minRateReqsPerSecond": 0.2000000000000008,
          "15minRateReqsPerSecond": 0.20000000000000234
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "description": "Reports application health to a load-balancer",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.PingRequestHandler"
      },
      "/update": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270811,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.UpdateRequestHandler"
      },
      "/terms": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "/admin/file": {
        "stats": {
          "999thPcRequestTime": 0.509739,
          "99thPcRequestTime": 0.509739,
          "95thPcRequestTime": 0.509739,
          "75thPcRequestTime": 0.38605350000000005,
          "medianRequestTime": 0.184437,
          "avgTimePerRequest": 0.2358606,
          "handlerStart": 1482259271569,
          "requests": 5,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 1.179303,
          "avgRequestsPerSecond": 2.0937536245042723e-06,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 3.0856020161426622e-152
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ShowFileRequestHandler.java $",
        "description": "Admin Get File -- view config files directly",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.ShowFileRequestHandler"
      },
      "/update/extract": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "version": null,
        "class": "Lazy[solr.extraction.ExtractingRequestHandler]"
      }
    },
    "UPDATEHANDLER",
    {
      "updateHandler": {
        "stats": {
          "cumulative_errors": 0,
          "expungeDeletes": 0,
          "rollbacks": 0,
          "optimizes": 0,
          "soft autocommits": 0,
          "autocommits": 0,
          "autocommit maxTime": "900ms",
          "autocommit maxDocs": 500,
          "commits": 0,
          "docsPending": 0,
          "adds": 0,
          "deletesById": 0,
          "deletesByQuery": 0,
          "errors": 0,
          "cumulative_adds": 0,
          "cumulative_deletesById": 0,
          "cumulative_deletesByQuery": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/update/DirectUpdateHandler2.java $",
        "description": "Update handler that efficiently directly updates the on-disk main lucene index",
        "version": "1.0",
        "class": "org.apache.solr.update.DirectUpdateHandler2"
      }
    },
    "CACHE",
    {
      "fieldCache": {
        "stats": {
          "insanity_count": 0,
          "entries_count": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrFieldCacheMBean.java $",
        "description": "Provides introspection of the Lucene FieldCache, this is **NOT** a cache that is managed by Solr.",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrFieldCacheMBean"
      },
      "fieldValueCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_hits": 0,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=5120, initialSize=5120, minSize=4608, acceptableSize=4864, cleanupThread=false, autowarmCount=1024, regenerator=org.apache.solr.search.SolrIndexSearcher$1@40acdd5)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "documentCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 5746,
          "cumulative_hitratio": "0.99",
          "cumulative_hits": 2405834,
          "lookups": 3750,
          "hits": 3733,
          "hitratio": "0.99",
          "inserts": 17,
          "evictions": 0,
          "size": 17,
          "warmupTime": 0,
          "cumulative_lookups": 2411580
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=10240, initialSize=10240, minSize=9216, acceptableSize=9728, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "queryResultCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 2660,
          "cumulative_hitratio": "0.99",
          "cumulative_hits": 726607,
          "lookups": 1206,
          "hits": 1179,
          "hitratio": "0.97",
          "inserts": 27,
          "evictions": 0,
          "size": 27,
          "warmupTime": 0,
          "cumulative_lookups": 729510
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=5120, initialSize=5120, minSize=4608, acceptableSize=4864, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "filterCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 14,
          "cumulative_hitratio": 0,
          "cumulative_hits": 55,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.01",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 69
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=2560, initialSize=2560, minSize=2304, acceptableSize=2432, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      }
    }
  ],
  "responseHeader": {
    "QTime": 8,
    "status": 0
  }
}
`

const mBeansCore1Response = `{
  "solr-mbeans": [
    "CORE",
    {
      "core": {
        "stats": {
          "aliases": [
            "corename"
          ],
          "indexDir": "/srv/solr-corename/index.20160607000000124",
          "refCount": 2,
          "startTime": "2016-12-20T18:41:10.449Z",
          "coreName": "core1"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/SolrCore.java $",
        "description": "SolrCore",
        "version": "1.0",
        "class": "core1"
      },
      "Searcher@6f0833b2 main": {
        "stats": {
          "warmupTime": 0,
          "registeredAt": "2017-01-14T12:00:00.209Z",
          "openedAt": "2017-01-14T12:00:00.208Z",
          "searcherName": "Searcher@6f0833b2 main",
          "caching": true,
          "numDocs": 7517488,
          "maxDoc": 7620303,
          "deletedDocs": 102815,
          "reader": "StandardDirectoryReader(segments_20iv:267485:nrt _2849(4.3.1):C7517434/102330 _28e8(4.3.1):C7363/115 _28h0(4.3.1):C5430/77 _28kw(4.3.1):C5984 _28k2(4.3.1):C6510/12 _28g6(4.3.1):C4537/25 _28ha(4.3.1):C5529/25 _28i4(4.3.1):C5087/42 _28js(4.3.1):C5823/10 _28ix(4.3.1):C5627/18 _28kc(4.3.1):C6710/14 _28kl(4.3.1):C7179/10 _28hk(4.3.1):C5149/65 _28j7(4.3.1):C5643/28 _28ht(4.3.1):C5428/9 _28ji(4.3.1):C5150/15 _28gq(4.3.1):C4989/9 _28ie(4.3.1):C5460/8 _28io(4.3.1):C5165/3 _28kv(4.3.1):C51 _28kx(4.3.1):C55)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core1/index.20160607000000124 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@646d42ce",
          "indexVersion": 267485
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "description": "index searcher",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrIndexSearcher"
      },
      "searcher": {
        "stats": {
          "warmupTime": 0,
          "registeredAt": "2017-01-14T12:00:00.209Z",
          "openedAt": "2017-01-14T12:00:00.208Z",
          "searcherName": "Searcher@6f0833b2 main",
          "caching": true,
          "numDocs": 7517488,
          "maxDoc": 7620303,
          "deletedDocs": 102815,
          "reader": "StandardDirectoryReader(segments_20iv:267485:nrt _2849(4.3.1):C7517434/102330 _28e8(4.3.1):C7363/115 _28h0(4.3.1):C5430/77 _28kw(4.3.1):C5984 _28k2(4.3.1):C6510/12 _28g6(4.3.1):C4537/25 _28ha(4.3.1):C5529/25 _28i4(4.3.1):C5087/42 _28js(4.3.1):C5823/10 _28ix(4.3.1):C5627/18 _28kc(4.3.1):C6710/14 _28kl(4.3.1):C7179/10 _28hk(4.3.1):C5149/65 _28j7(4.3.1):C5643/28 _28ht(4.3.1):C5428/9 _28ji(4.3.1):C5150/15 _28gq(4.3.1):C4989/9 _28ie(4.3.1):C5460/8 _28io(4.3.1):C5165/3 _28kv(4.3.1):C51 _28kx(4.3.1):C55)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core1/index.20160607000000124 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@646d42ce",
          "indexVersion": 267485
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "description": "index searcher",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrIndexSearcher"
      }
    },
    "QUERYHANDLER",
    {
      "org.apache.solr.handler.CSVRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270458,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.CSVRequestHandler"
      },
      "/admin/": {
        "stats": null,
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "description": "Register Standard Admin Handlers",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.AdminHandlers"
      },
      "/admin/mbeans": {
        "stats": {
          "999thPcRequestTime": 127.79069480400044,
          "99thPcRequestTime": 9.033666420000003,
          "95thPcRequestTime": 5.586449799999999,
          "75thPcRequestTime": 4.68247075,
          "medianRequestTime": 0.03985,
          "avgTimePerRequest": 1.5857040673599807,
          "handlerStart": 1482259270585,
          "requests": 230969,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 366246.89703,
          "avgRequestsPerSecond": 0.09671315555928528,
          "5minRateReqsPerSecond": 0.545082835587804,
          "15minRateReqsPerSecond": 0.5414280756665533
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SolrInfoMBeanHandler.java $",
        "description": "Get Info (and statistics) for registered SolrInfoMBeans",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.SolrInfoMBeanHandler"
      },
      "/debug/dump": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270462,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "description": "Dump handler (debug)",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.DumpRequestHandler"
      },
      "/admin/logging": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LoggingHandler.java $",
        "description": "Logging Handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.LoggingHandler"
      },
      "/admin/plugins": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PluginInfoHandler.java $",
        "description": "Registry",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.PluginInfoHandler"
      },
      "/admin/system": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SystemInfoHandler.java $",
        "description": "Get System Info",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.SystemInfoHandler"
      },
      "/select": {
        "stats": {
          "999thPcRequestTime": 0.518583,
          "99thPcRequestTime": 0.518583,
          "95thPcRequestTime": 0.518583,
          "75thPcRequestTime": 0.518583,
          "medianRequestTime": 0.518583,
          "avgTimePerRequest": 0.518583,
          "handlerStart": 1482259270455,
          "requests": 1,
          "errors": 1,
          "timeouts": 0,
          "totalTime": 0.518583,
          "avgRequestsPerSecond": 4.187296521163843e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 4.44659081257e-313
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,spellcheck,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/tvrh": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "org.apache.solr.handler.component.SearchHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270455,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/admin/luke": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "docs": [
          "http://wiki.apache.org/solr/LukeRequestHandler"
        ],
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LukeRequestHandler.java $",
        "description": "Lucene Index Browser.  Inspired and modeled after Luke: http://www.getopt.org/luke/",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.LukeRequestHandler"
      },
      "/update/json": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270457,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "description": "Add documents with JSON",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler"
      },
      "org.apache.solr.handler.admin.AdminHandlers": {
        "stats": null,
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "description": "Register Standard Admin Handlers",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.AdminHandlers"
      },
      "org.apache.solr.handler.ReplicationHandler": {
        "stats": {
          "lastCycleBytesDownloaded": "2578996",
          "timesIndexReplicated": "468",
          "timesFailed": "2",
          "replicationFailedAt": "Fri Feb 12 00:00:00 UTC 2016",
          "indexReplicatedAt": "Sat Jan 14 12:00:00 UTC 2017",
          "previousCycleTimeInSeconds": "0",
          "isReplicating": "false",
          "isPollingDisabled": "false",
          "pollInterval": "12:00:00",
          "masterUrl": "http://solr-s1:8983/solr/core1",
          "isSlave": "true",
          "isMaster": "false",
          "indexPath": "/srv/solr-core1/index.20160607000000124",
          "generation": 93991,
          "15minRateReqsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "avgRequestsPerSecond": 0,
          "totalTime": 0,
          "timeouts": 0,
          "errors": 0,
          "requests": 0,
          "handlerStart": 1482259270463,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0,
          "indexSize": "1.66 GB",
          "indexVersion": 1484389807419
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.ReplicationHandler"
      },
      "org.apache.solr.handler.JsonUpdateRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270457,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "description": "Add documents with JSON",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler"
      },
      "org.apache.solr.handler.DumpRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270462,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "description": "Dump handler (debug)",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.DumpRequestHandler"
      },
      "org.apache.solr.handler.RealTimeGetHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270456,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "description": "The realtime get handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.RealTimeGetHandler"
      },
      "/get": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270456,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "description": "The realtime get handler",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.RealTimeGetHandler"
      },
      "/admin/properties": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PropertiesRequestHandler.java $",
        "description": "Get System Properties",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.PropertiesRequestHandler"
      },
      "/query": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270455,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.component.SearchHandler"
      },
      "/admin/threads": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ThreadDumpHandler.java $",
        "description": "Thread Dump",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.ThreadDumpHandler"
      },
      "/analysis/field": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.FieldAnalysisRequestHandler]",
        "version": null,
        "class": "Lazy[solr.FieldAnalysisRequestHandler]"
      },
      "org.apache.solr.handler.PingRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270461,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "description": "Reports application health to a load-balancer",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.PingRequestHandler"
      },
      "/analysis/document": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "version": null,
        "class": "Lazy[solr.DocumentAnalysisRequestHandler]"
      },
      "/spell": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "/update/csv": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270458,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.CSVRequestHandler"
      },
      "org.apache.solr.handler.UpdateRequestHandler": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270457,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.UpdateRequestHandler"
      },
      "/replication": {
        "stats": {
          "lastCycleBytesDownloaded": "2578996",
          "timesIndexReplicated": "468",
          "timesFailed": "2",
          "replicationFailedAt": "Fri Feb 12 00:00:00 UTC 2016",
          "indexReplicatedAt": "Sat Jan 14 12:00:00 UTC 2017",
          "previousCycleTimeInSeconds": "0",
          "isReplicating": "false",
          "isPollingDisabled": "false",
          "pollInterval": "12:00:00",
          "masterUrl": "http://solr-s1:8983/solr/core1",
          "isSlave": "true",
          "isMaster": "false",
          "indexPath": "/srv/solr-core1/index.20160607000000124",
          "generation": 93991,
          "15minRateReqsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "avgRequestsPerSecond": 0,
          "totalTime": 0,
          "timeouts": 0,
          "errors": 0,
          "requests": 0,
          "handlerStart": 1482259270463,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0,
          "indexSize": "1.66 GB",
          "indexVersion": 1484389807419
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.ReplicationHandler"
      },
      "/admin/ping": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270461,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "description": "Reports application health to a load-balancer",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.PingRequestHandler"
      },
      "/update": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270457,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.UpdateRequestHandler"
      },
      "/terms": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.SearchHandler]",
        "version": null,
        "class": "Lazy[solr.SearchHandler]"
      },
      "/admin/file": {
        "stats": {
          "999thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "75thPcRequestTime": 0,
          "medianRequestTime": 0,
          "avgTimePerRequest": 0,
          "handlerStart": 1482259270585,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ShowFileRequestHandler.java $",
        "description": "Admin Get File -- view config files directly",
        "version": "4.3.1",
        "class": "org.apache.solr.handler.admin.ShowFileRequestHandler"
      },
      "/update/extract": {
        "stats": {
          "note": "not initialized yet"
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "description": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "version": null,
        "class": "Lazy[solr.extraction.ExtractingRequestHandler]"
      }
    },
    "UPDATEHANDLER",
    {
      "updateHandler": {
        "stats": {
          "cumulative_errors": 0,
          "expungeDeletes": 0,
          "rollbacks": 0,
          "optimizes": 0,
          "soft autocommits": 0,
          "autocommits": 0,
          "autocommit maxTime": "900ms",
          "autocommit maxDocs": 500,
          "commits": 0,
          "docsPending": 0,
          "adds": 0,
          "deletesById": 0,
          "deletesByQuery": 0,
          "errors": 0,
          "cumulative_adds": 0,
          "cumulative_deletesById": 0,
          "cumulative_deletesByQuery": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/update/DirectUpdateHandler2.java $",
        "description": "Update handler that efficiently directly updates the on-disk main lucene index",
        "version": "1.0",
        "class": "org.apache.solr.update.DirectUpdateHandler2"
      }
    },
    "CACHE",
    {
      "fieldCache": {
        "stats": {
          "insanity_count": 0,
          "entries_count": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrFieldCacheMBean.java $",
        "description": "Provides introspection of the Lucene FieldCache, this is **NOT** a cache that is managed by Solr.",
        "version": "1.0",
        "class": "org.apache.solr.search.SolrFieldCacheMBean"
      },
      "fieldValueCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_hits": 0,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=4096, initialSize=4096, minSize=3686, acceptableSize=3891, cleanupThread=false, autowarmCount=128, regenerator=org.apache.solr.search.SolrIndexSearcher$1@58c2e3e9)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "documentCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_hits": 0,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=32768, initialSize=32768, minSize=29491, acceptableSize=31129, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "queryResultCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_hits": 0,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=4096, initialSize=4096, minSize=3686, acceptableSize=3891, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      },
      "filterCache": {
        "stats": {
          "cumulative_evictions": 0,
          "cumulative_inserts": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_hits": 0,
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0
        },
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "description": "Concurrent LRU Cache(maxSize=512, initialSize=512, minSize=460, acceptableSize=486, cleanupThread=false)",
        "version": "1.0",
        "class": "org.apache.solr.search.FastLRUCache"
      }
    }
  ],
  "responseHeader": {
    "QTime": 5,
    "status": 0
  }
}
`

var solrAdminMainCoreStatusExpected = map[string]interface{}{
	"num_docs":      int64(168943425),
	"max_docs":      int64(169562700),
	"deleted_docs":  int64(619275),
	"size_in_bytes": int64(247497521642),
}

var solrAdminCore1StatusExpected = map[string]interface{}{
	"num_docs":      int64(7517488),
	"max_docs":      int64(7620303),
	"deleted_docs":  int64(102815),
	"size_in_bytes": int64(1784635686),
}

var solrCoreExpected = map[string]interface{}{
	"num_docs":     int64(168962621),
	"max_docs":     int64(169647870),
	"deleted_docs": int64(685249),
}

var solrQueryHandlerExpected = map[string]interface{}{
	"15min_rate_reqs_per_second": float64(0),
	"5min_rate_reqs_per_second":  float64(0),
	"75th_pc_request_time":       float64(0),
	"95th_pc_request_time":       float64(0),
	"999th_pc_request_time":      float64(0),
	"99th_pc_request_time":       float64(0),
	"avg_requests_per_second":    float64(0),
	"avg_time_per_request":       float64(0),
	"errors":                     int64(0),
	"handler_start":              int64(1482259270810),
	"median_request_time":        float64(0),
	"requests":                   int64(0),
	"timeouts":                   int64(0),
	"total_time":                 float64(0),
}

var solrUpdateHandlerExpected = map[string]interface{}{
	"adds":                        int64(0),
	"autocommit_max_docs":         int64(500),
	"autocommit_max_time":         int64(900),
	"autocommits":                 int64(0),
	"commits":                     int64(0),
	"cumulative_adds":             int64(0),
	"cumulative_deletes_by_id":    int64(0),
	"cumulative_deletes_by_query": int64(0),
	"cumulative_errors":           int64(0),
	"deletes_by_id":               int64(0),
	"deletes_by_query":            int64(0),
	"docs_pending":                int64(0),
	"errors":                      int64(0),
	"expunge_deletes":             int64(0),
	"optimizes":                   int64(0),
	"rollbacks":                   int64(0),
	"soft_autocommits":            int64(0),
}

var solrCacheExpected = map[string]interface{}{
	"cumulative_evictions": int64(0),
	"cumulative_hitratio":  float64(0),
	"cumulative_hits":      int64(55),
	"cumulative_inserts":   int64(14),
	"cumulative_lookups":   int64(69),
	"evictions":            int64(0),
	"hitratio":             float64(0.01),
	"hits":                 int64(0),
	"inserts":              int64(0),
	"lookups":              int64(0),
	"size":                 int64(0),
	"warmup_time":          int64(0),
}
