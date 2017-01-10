package solr

const coreStatsResponse = `
{
  "responseHeader": {
    "status": 0,
    "QTime": 0
  },
  "solr-mbeans": [
    "CORE",
    {
      "searcher": {
        "class": "org.apache.solr.search.SolrIndexSearcher",
        "version": "1.0",
        "description": "index searcher",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "stats": {
          "searcherName": "Searcher@eee4e51 main",
          "caching": true,
          "numDocs": 1822549,
          "maxDoc": 2814012,
          "deletedDocs": 991463,
          "reader": "StandardDirectoryReader(segments_x2t9:4913665:nrt _zilg(4.3.1):C1798549/651601 _101qi(4.3.1):C314957/149516 _10dk8(4.3.1):C248175/79443 _10a6t(4.3.1):C138149/50435 _10hkl(4.3.1):C115963/41723 _10lm0(4.3.1):C86704/17303 _10nkv(4.3.1):C97442/1015 _10n6z(4.3.1):C2670/186 _10njr(4.3.1):C2505/23 _10nkb(4.3.1):C2610/43 _10nkl(4.3.1):C2697/2 _10nl6(4.3.1):C427/20 _10nlp(4.3.1):C457/1 _10nlf(4.3.1):C319/7 _10nlz(4.3.1):C412/1 _10nn3(4.3.1):C408 _10nm9(4.3.1):C315/1 _10nmj(4.3.1):C354/42 _10nmt(4.3.1):C431 _10nmu(4.3.1):C47 _10nn4(4.3.1):C50/20 _10nn5(4.3.1):C50/32 _10nn6(4.3.1):C50/25 _10nn7(4.3.1):C50/19 _10nn8(4.3.1):C50/1 _10nn9(4.3.1):C50/3 _10nna(4.3.1):C50 _10nnb(4.3.1):C50/1 _10nnc(4.3.1):C21)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-randomapi/index.20160606235000107 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@798c8d0e",
          "indexVersion": 4913665,
          "openedAt": "2016-09-06T10:10:01.849Z",
          "registeredAt": "2016-09-06T10:10:01.85Z",
          "warmupTime": 0
        }
      },
      "core": {
        "class": "randomapi",
        "version": "1.0",
        "description": "SolrCore",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/SolrCore.java $",
        "stats": {
          "coreName": "randomapi",
          "startTime": "2016-08-05T00:18:17.689Z",
          "refCount": 2,
          "indexDir": "/srv/solr-randomapi/index.20160606235000107",
          "aliases": [
            "randomapi"
          ]
        }
      },
      "Searcher@eee4e51 main": {
        "class": "org.apache.solr.search.SolrIndexSearcher",
        "version": "1.0",
        "description": "index searcher",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "stats": {
          "searcherName": "Searcher@eee4e51 main",
          "caching": true,
          "numDocs": 1822549,
          "maxDoc": 2814012,
          "deletedDocs": 991463,
          "reader": "StandardDirectoryReader(segments_x2t9:4913665:nrt _zilg(4.3.1):C1798549/651601 _101qi(4.3.1):C314957/149516 _10dk8(4.3.1):C248175/79443 _10a6t(4.3.1):C138149/50435 _10hkl(4.3.1):C115963/41723 _10lm0(4.3.1):C86704/17303 _10nkv(4.3.1):C97442/1015 _10n6z(4.3.1):C2670/186 _10njr(4.3.1):C2505/23 _10nkb(4.3.1):C2610/43 _10nkl(4.3.1):C2697/2 _10nl6(4.3.1):C427/20 _10nlp(4.3.1):C457/1 _10nlf(4.3.1):C319/7 _10nlz(4.3.1):C412/1 _10nn3(4.3.1):C408 _10nm9(4.3.1):C315/1 _10nmj(4.3.1):C354/42 _10nmt(4.3.1):C431 _10nmu(4.3.1):C47 _10nn4(4.3.1):C50/20 _10nn5(4.3.1):C50/32 _10nn6(4.3.1):C50/25 _10nn7(4.3.1):C50/19 _10nn8(4.3.1):C50/1 _10nn9(4.3.1):C50/3 _10nna(4.3.1):C50 _10nnb(4.3.1):C50/1 _10nnc(4.3.1):C21)",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-randomapi/index.20160606235000107 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@798c8d0e",
          "indexVersion": 4913665,
          "openedAt": "2016-09-06T10:10:01.849Z",
          "registeredAt": "2016-09-06T10:10:01.85Z",
          "warmupTime": 0
        }
      }
    }
  ]
}
`

const queryHandlerStatsResponse = `
{
  "responseHeader": {
    "status": 0,
    "QTime": 6
  },
  "solr-mbeans": [
    "QUERYHANDLER",
    {
      "org.apache.solr.handler.component.SearchHandler": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "4.3.1",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "stats": {
          "handlerStart": 1470356299096,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/luke": {
        "class": "org.apache.solr.handler.admin.LukeRequestHandler",
        "version": "4.3.1",
        "description": "Lucene Index Browser.  Inspired and modeled after Luke: http://www.getopt.org/luke/",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LukeRequestHandler.java $",
        "docs": [
          "http://wiki.apache.org/solr/LukeRequestHandler"
        ],
        "stats": {
          "handlerStart": 1470356300106,
          "requests": 1,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 1.354916,
          "avgRequestsPerSecond": 3.471186326369445e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.8675124772e-313,
          "avgTimePerRequest": 1.354916,
          "medianRequestTime": 1.354916,
          "75thPcRequestTime": 1.354916,
          "95thPcRequestTime": 1.354916,
          "99thPcRequestTime": 1.354916,
          "999thPcRequestTime": 1.354916
        }
      },
      "/update/json": {
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler",
        "version": "4.3.1",
        "description": "Add documents with JSON",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299104,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "org.apache.solr.handler.admin.AdminHandlers": {
        "class": "org.apache.solr.handler.admin.AdminHandlers",
        "version": "4.3.1",
        "description": "Register Standard Admin Handlers",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "stats": null
      },
      "org.apache.solr.handler.ReplicationHandler": {
        "class": "org.apache.solr.handler.ReplicationHandler",
        "version": "4.3.1",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "stats": {
          "handlerStart": 1470356299265,
          "requests": 1,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 13.991382,
          "avgRequestsPerSecond": 3.471185312149739e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.8675124772e-313,
          "avgTimePerRequest": 13.991382,
          "medianRequestTime": 13.991382,
          "75thPcRequestTime": 13.991382,
          "95thPcRequestTime": 13.991382,
          "99thPcRequestTime": 13.991382,
          "999thPcRequestTime": 13.991382,
          "indexSize": "2.96 GB",
          "indexVersion": 1473156102523,
          "generation": 1543293,
          "indexPath": "/srv/solr-randomapi/index.20160606235000107",
          "isMaster": "false",
          "isSlave": "true",
          "masterUrl": "http://localhost:8983/solr/randomapi",
          "pollInterval": "00:10:00",
          "isPollingDisabled": "false",
          "isReplicating": "false",
          "previousCycleTimeInSeconds": "1",
          "indexReplicatedAt": "Tue Sep 06 10:10:01 UTC 2016",
          "replicationFailedAt": "Mon Jun 06 23:10:00 UTC 2016",
          "timesFailed": "41",
          "timesIndexReplicated": "328",
          "lastCycleBytesDownloaded": "128462700"
        }
      },
      "org.apache.solr.handler.JsonUpdateRequestHandler": {
        "class": "org.apache.solr.handler.JsonUpdateRequestHandler",
        "version": "4.3.1",
        "description": "Add documents with JSON",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299104,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "org.apache.solr.handler.DumpRequestHandler": {
        "class": "org.apache.solr.handler.DumpRequestHandler",
        "version": "4.3.1",
        "description": "Dump handler (debug)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299184,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "org.apache.solr.handler.RealTimeGetHandler": {
        "class": "org.apache.solr.handler.RealTimeGetHandler",
        "version": "4.3.1",
        "description": "The realtime get handler",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "stats": {
          "handlerStart": 1470356299099,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/tvrh": {
        "class": "Lazy[solr.SearchHandler]",
        "version": null,
        "description": "Lazy[solr.SearchHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/select": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "4.3.1",
        "description": "Search using components: query,facet,mlt,highlight,stats,spellcheck,debug,",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "stats": {
          "handlerStart": 1470356299058,
          "requests": 1,
          "errors": 1,
          "timeouts": 0,
          "totalTime": 169.655217,
          "avgRequestsPerSecond": 3.471185059164947e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 4.44659081257e-313,
          "avgTimePerRequest": 169.655217,
          "medianRequestTime": 169.655217,
          "75thPcRequestTime": 169.655217,
          "95thPcRequestTime": 169.655217,
          "99thPcRequestTime": 169.655217,
          "999thPcRequestTime": 169.655217
        }
      },
      "/admin/system": {
        "class": "org.apache.solr.handler.admin.SystemInfoHandler",
        "version": "4.3.1",
        "description": "Get System Info",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SystemInfoHandler.java $",
        "stats": {
          "handlerStart": 1470356300108,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/plugins": {
        "class": "org.apache.solr.handler.admin.PluginInfoHandler",
        "version": "4.3.1",
        "description": "Registry",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PluginInfoHandler.java $",
        "stats": {
          "handlerStart": 1470356300110,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/logging": {
        "class": "org.apache.solr.handler.admin.LoggingHandler",
        "version": "4.3.1",
        "description": "Logging Handler",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/LoggingHandler.java $",
        "stats": {
          "handlerStart": 1470356300113,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/debug/dump": {
        "class": "org.apache.solr.handler.DumpRequestHandler",
        "version": "4.3.1",
        "description": "Dump handler (debug)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299184,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/mbeans": {
        "class": "org.apache.solr.handler.admin.SolrInfoMBeanHandler",
        "version": "4.3.1",
        "description": "Get Info (and statistics) for registered SolrInfoMBeans",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/SolrInfoMBeanHandler.java $",
        "stats": {
          "handlerStart": 1470356300109,
          "requests": 537,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 2373.411334,
          "avgRequestsPerSecond": 0.00018605558705804686,
          "5minRateReqsPerSecond": 0.0024283836550347184,
          "15minRateReqsPerSecond": 0.0031135268227272966,
          "avgTimePerRequest": 4.428006220149253,
          "medianRequestTime": 0.180612,
          "75thPcRequestTime": 4.9944925,
          "95thPcRequestTime": 6.2053377999999935,
          "99thPcRequestTime": 13.340844199999996,
          "999thPcRequestTime": 94.803079
        }
      },
      "/admin/": {
        "class": "org.apache.solr.handler.admin.AdminHandlers",
        "version": "4.3.1",
        "description": "Register Standard Admin Handlers",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "stats": null
      },
      "/get": {
        "class": "org.apache.solr.handler.RealTimeGetHandler",
        "version": "4.3.1",
        "description": "The realtime get handler",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/RealTimeGetHandler.java $",
        "stats": {
          "handlerStart": 1470356299099,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/properties": {
        "class": "org.apache.solr.handler.admin.PropertiesRequestHandler",
        "version": "4.3.1",
        "description": "Get System Properties",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/PropertiesRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356300111,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/query": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "4.3.1",
        "description": "Search using components: query,facet,mlt,highlight,stats,debug,",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "stats": {
          "handlerStart": 1470356299096,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/admin/threads": {
        "class": "org.apache.solr.handler.admin.ThreadDumpHandler",
        "version": "4.3.1",
        "description": "Thread Dump",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ThreadDumpHandler.java $",
        "stats": {
          "handlerStart": 1470356300111,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/analysis/field": {
        "class": "Lazy[solr.FieldAnalysisRequestHandler]",
        "version": null,
        "description": "Lazy[solr.FieldAnalysisRequestHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "org.apache.solr.handler.PingRequestHandler": {
        "class": "org.apache.solr.handler.PingRequestHandler",
        "version": "4.3.1",
        "description": "Reports application health to a load-balancer",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299182,
          "requests": 1,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0.222296,
          "avgRequestsPerSecond": 3.471185208104538e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.9169190418e-313,
          "avgTimePerRequest": 0.222296,
          "medianRequestTime": 0.222296,
          "75thPcRequestTime": 0.222296,
          "95thPcRequestTime": 0.222296,
          "99thPcRequestTime": 0.222296,
          "999thPcRequestTime": 0.222296
        }
      },
      "/analysis/document": {
        "class": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "version": null,
        "description": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/spell": {
        "class": "Lazy[solr.SearchHandler]",
        "version": null,
        "description": "Lazy[solr.SearchHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/update/csv": {
        "class": "org.apache.solr.handler.CSVRequestHandler",
        "version": "4.3.1",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299106,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "org.apache.solr.handler.UpdateRequestHandler": {
        "class": "org.apache.solr.handler.UpdateRequestHandler",
        "version": "4.3.1",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299102,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/replication": {
        "class": "org.apache.solr.handler.ReplicationHandler",
        "version": "4.3.1",
        "description": "ReplicationHandler provides replication of index and configuration files from Master to Slaves",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/ReplicationHandler.java $",
        "stats": {
          "handlerStart": 1470356299265,
          "requests": 1,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 13.991382,
          "avgRequestsPerSecond": 3.471185307602568e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.8675124772e-313,
          "avgTimePerRequest": 13.991382,
          "medianRequestTime": 13.991382,
          "75thPcRequestTime": 13.991382,
          "95thPcRequestTime": 13.991382,
          "99thPcRequestTime": 13.991382,
          "999thPcRequestTime": 13.991382,
          "indexSize": "2.96 GB",
          "indexVersion": 1473156102523,
          "generation": 1543293,
          "indexPath": "/srv/solr-randomapi/index.20160606235000107",
          "isMaster": "false",
          "isSlave": "true",
          "masterUrl": "http://localhost:8983/solr/randomapi",
          "pollInterval": "00:10:00",
          "isPollingDisabled": "false",
          "isReplicating": "false",
          "previousCycleTimeInSeconds": "1",
          "indexReplicatedAt": "Tue Sep 06 10:10:01 UTC 2016",
          "replicationFailedAt": "Mon Jun 06 23:10:00 UTC 2016",
          "timesFailed": "41",
          "timesIndexReplicated": "328",
          "lastCycleBytesDownloaded": "128462700"
        }
      },
      "/admin/ping": {
        "class": "org.apache.solr.handler.PingRequestHandler",
        "version": "4.3.1",
        "description": "Reports application health to a load-balancer",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299182,
          "requests": 1,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0.222296,
          "avgRequestsPerSecond": 3.471185205347671e-07,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.9169190418e-313,
          "avgTimePerRequest": 0.222296,
          "medianRequestTime": 0.222296,
          "75thPcRequestTime": 0.222296,
          "95thPcRequestTime": 0.222296,
          "99thPcRequestTime": 0.222296,
          "999thPcRequestTime": 0.222296
        }
      },
      "/update": {
        "class": "org.apache.solr.handler.UpdateRequestHandler",
        "version": "4.3.1",
        "description": "Add documents using XML (with XSLT), CSV, JSON, or javabin",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/UpdateRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299102,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      },
      "/terms": {
        "class": "Lazy[solr.SearchHandler]",
        "version": null,
        "description": "Lazy[solr.SearchHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/admin/file": {
        "class": "org.apache.solr.handler.admin.ShowFileRequestHandler",
        "version": "4.3.1",
        "description": "Admin Get File -- view config files directly",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/admin/ShowFileRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356300115,
          "requests": 3,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0.731451,
          "avgRequestsPerSecond": 1.0413558987622932e-06,
          "5minRateReqsPerSecond": 1.4821969375e-313,
          "15minRateReqsPerSecond": 6.9169190418e-313,
          "avgTimePerRequest": 0.243817,
          "medianRequestTime": 0.11456,
          "75thPcRequestTime": 0.530102,
          "95thPcRequestTime": 0.530102,
          "99thPcRequestTime": 0.530102,
          "999thPcRequestTime": 0.530102
        }
      },
      "/update/extract": {
        "class": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "version": null,
        "description": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "stats": {
          "note": "not initialized yet"
        }
      },
      "org.apache.solr.handler.CSVRequestHandler": {
        "class": "org.apache.solr.handler.CSVRequestHandler",
        "version": "4.3.1",
        "description": "Add/Update multiple documents with CSV formatted rows",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/handler/CSVRequestHandler.java $",
        "stats": {
          "handlerStart": 1470356299106,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgRequestsPerSecond": 0,
          "5minRateReqsPerSecond": 0,
          "15minRateReqsPerSecond": 0,
          "avgTimePerRequest": 0,
          "medianRequestTime": 0,
          "75thPcRequestTime": 0,
          "95thPcRequestTime": 0,
          "99thPcRequestTime": 0,
          "999thPcRequestTime": 0
        }
      }
    }
  ]
}
`

const updateHandlerStatsResponse = `
{
  "responseHeader": {
    "status": 0,
    "QTime": 0
  },
  "solr-mbeans": [
    "UPDATEHANDLER",
    {
      "updateHandler": {
        "class": "org.apache.solr.update.DirectUpdateHandler2",
        "version": "1.0",
        "description": "Update handler that efficiently directly updates the on-disk main lucene index",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/update/DirectUpdateHandler2.java $",
        "stats": {
          "commits": 0,
          "autocommit maxDocs": 500,
          "autocommit maxTime": "900ms",
          "autocommits": 0,
          "soft autocommits": 0,
          "optimizes": 0,
          "rollbacks": 0,
          "expungeDeletes": 0,
          "docsPending": 0,
          "adds": 0,
          "deletesById": 0,
          "deletesByQuery": 0,
          "errors": 0,
          "cumulative_adds": 0,
          "cumulative_deletesById": 0,
          "cumulative_deletesByQuery": 0,
          "cumulative_errors": 0
        }
      }
    }
  ]
}
`

const cacheStatsResponse = `
{
  "responseHeader": {
    "status": 0,
    "QTime": 0
  },
  "solr-mbeans": [
    "CACHE",
    {
      "filterCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=512, initialSize=512, minSize=460, acceptableSize=486, cleanupThread=false)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "stats": {
          "lookups": 0,
          "hits": 1,
          "hitratio": "1.01",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0,
          "cumulative_hits": 0,
          "cumulative_hitratio": "1.01",
          "cumulative_inserts": 0,
          "cumulative_evictions": 0
        }
      },
      "queryResultCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=4096, initialSize=4096, minSize=3686, acceptableSize=3891, cleanupThread=false)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "stats": {
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0,
          "cumulative_hits": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_inserts": 0,
          "cumulative_evictions": 0
        }
      },
      "documentCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=32768, initialSize=32768, minSize=29491, acceptableSize=31129, cleanupThread=false)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "stats": {
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0,
          "cumulative_hits": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_inserts": 0,
          "cumulative_evictions": 0
        }
      },
      "fieldValueCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=4096, initialSize=4096, minSize=3686, acceptableSize=3891, cleanupThread=false, autowarmCount=128, regenerator=org.apache.solr.search.SolrIndexSearcher$1@c7bc9c5)",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "stats": {
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 0,
          "cumulative_hits": 0,
          "cumulative_hitratio": "0.00",
          "cumulative_inserts": 0,
          "cumulative_evictions": 0
        }
      },
      "fieldCache": {
        "class": "org.apache.solr.search.SolrFieldCacheMBean",
        "version": "1.0",
        "description": "Provides introspection of the Lucene FieldCache, this is **NOT** a cache that is managed by Solr.",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_4_3/solr/core/src/java/org/apache/solr/search/SolrFieldCacheMBean.java $",
        "stats": {
          "entries_count": 0,
          "insanity_count": 0
        }
      }
    }
  ]
}
`

const adminCoresResponse = `
{
  "responseHeader": {
    "status": 0,
    "QTime": 10
  },
  "defaultCoreName": "main",
  "initFailures": {},
  "status": {
    "core0": {
      "name": "core0",
      "isDefaultCore": false,
      "instanceDir": "solr/core0/",
      "dataDir": "/srv/solr-core0/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-08-05T00:18:20.871Z",
      "uptime": 3054210877,
      "index": {
        "numDocs": 38732,
        "maxDoc": 38732,
        "deletedDocs": 0,
        "version": 256479,
        "segmentCount": 3,
        "current": true,
        "hasDeletions": false,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core0/index.20160606181000006 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@2c9bc1b0",
        "userData": {
          "commitTimeMSec": "1473397332781"
        },
        "lastModified": "2016-09-09T05:02:12.781Z",
        "sizeInBytes": 5304845,
        "size": "5.06 MB"
      }
    },
    "core1": {
      "name": "core1",
      "isDefaultCore": false,
      "instanceDir": "solr/core1/",
      "dataDir": "/srv/solr-core1/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-08-05T00:18:17.689Z",
      "uptime": 3054214059,
      "index": {
        "numDocs": 1823284,
        "maxDoc": 2840737,
        "deletedDocs": 1017453,
        "version": 4916340,
        "segmentCount": 23,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core1/index.20160606235000107 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@798c8d0e",
        "userData": {
          "commitTimeMSec": "1473328888638"
        },
        "lastModified": "2016-09-08T10:01:28.638Z",
        "sizeInBytes": 3208083615,
        "size": "2.99 GB"
      }
    },
    "main": {
      "name": "main",
      "isDefaultCore": true,
      "instanceDir": "solr/main/",
      "dataDir": "/srv/solr/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-08-05T00:18:19.517Z",
      "uptime": 3054212233,
      "index": {
        "numDocs": 238785023,
        "maxDoc": 250822790,
        "deletedDocs": 12037767,
        "version": 67802912,
        "segmentCount": 43,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr/index lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@3e260545",
        "userData": {
          "commitTimeMSec": "1473400796980"
        },
        "lastModified": "2016-09-09T05:59:56.98Z",
        "sizeInBytes": 372377512357,
        "size": "346.8 GB"
      }
    },
    "core2": {
      "name": "core2",
      "isDefaultCore": false,
      "instanceDir": "solr/core2/",
      "dataDir": "/srv/solr-core2/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-08-05T00:18:20.867Z",
      "uptime": 3054210887,
      "index": {
        "numDocs": 7517469,
        "maxDoc": 7538430,
        "deletedDocs": 20961,
        "version": 266096,
        "segmentCount": 22,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core2/index.20160607000000270 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@52f4b09b",
        "userData": {
          "commitTimeMSec": "1472812205664"
        },
        "lastModified": "2016-09-02T10:30:05.664Z",
        "sizeInBytes": 1762126735,
        "size": "1.64 GB"
      }
    },
    "core3": {
      "name": "core3",
      "isDefaultCore": false,
      "instanceDir": "solr/core3/",
      "dataDir": "/srv/solr-core3/",
      "config": "solrconfig.xml",
      "schema": "schema.xml",
      "startTime": "2016-08-05T00:18:19.262Z",
      "uptime": 3054212494,
      "index": {
        "numDocs": 415176,
        "maxDoc": 485825,
        "deletedDocs": 70649,
        "version": 282990385,
        "segmentCount": 18,
        "current": true,
        "hasDeletions": true,
        "directory": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/srv/solr-core3/index.20160606235000106 lockFactory=org.apache.lucene.store.SingleInstanceLockFactory@20107d34",
        "userData": {
          "commitTimeMSec": "1473410395901"
        },
        "lastModified": "2016-09-09T08:39:55.901Z",
        "sizeInBytes": 939779160,
        "size": "896.24 MB"
      }
    }
  }
}
`

var solrCoreExpected = map[string]interface{}{
	"class_name":   string("org.apache.solr.search.SolrIndexSearcher"),
	"num_docs":     int(1822549),
	"max_docs":     int(2814012),
	"deleted_docs": int(991463),
}

var solrQueryHandlerExpected = map[string]interface{}{
	"class_name":                 string("org.apache.solr.handler.component.SearchHandler"),
	"15min_rate_reqs_per_second": float64(0),
	"5min_rate_reqs_per_second":  float64(0),
	"75th_pc_request_time":       float64(0),
	"95th_pc_request_time":       float64(0),
	"999th_pc_request_time":      float64(0),
	"99th_pc_request_time":       float64(0),
	"avg_requests_per_second":    float64(0),
	"avg_time_per_request":       float64(0),
	"errors":                     int(0),
	"handler_start":              int(1470356299096),
	"median_request_time":        float64(0),
	"requests":                   int(0),
	"timeouts":                   int(0),
	"total_time":                 float64(0),
}

var solrUpdateHandlerExpected = map[string]interface{}{
	"class_name":                  string("org.apache.solr.update.DirectUpdateHandler2"),
	"adds":                        int(0),
	"autocommit_max_docs":         int(500),
	"autocommit_max_time":         int(900),
	"autocommits":                 int(0),
	"commits":                     int(0),
	"cumulative_adds":             int(0),
	"cumulative_deletes_by_id":    int(0),
	"cumulative_deletes_by_query": int(0),
	"cumulative_errors":           int(0),
	"deletes_by_id":               int(0),
	"deletes_by_query":            int(0),
	"docs_pending":                int(0),
	"errors":                      int(0),
	"expunge_deletes":             int(0),
	"optimizes":                   int(0),
	"rollbacks":                   int(0),
	"soft_autocommits":            int(0),
}

var solrCacheExpected = map[string]interface{}{
	"class_name":           string("org.apache.solr.search.FastLRUCache"),
	"cumulative_evictions": int(0),
	"cumulative_hitratio":  float64(1.01),
	"cumulative_hits":      int(0),
	"cumulative_inserts":   int(0),
	"cumulative_lookups":   int(0),
	"evictions":            int(0),
	"hitratio":             float64(1.01),
	"hits":                 int(1),
	"inserts":              int(0),
	"lookups":              int(0),
	"size":                 int(0),
	"warmup_time":          int(0),
}
