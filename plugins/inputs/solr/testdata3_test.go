package solr

const mBeansSolr3MainResponse = `{
  "solr-mbeans": [
    "CORE",
    {
      "searcher": {
        "class": "org.apache.solr.search.SolrIndexSearcher",
        "version": "1.0",
        "description": "index searcher",
        "srcId": "$Id: SolrIndexSearcher.java 1201291 2011-11-12 18:02:03Z simonw $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "docs": null,
        "stats": {
          "searcherName": "Searcher@4eea69e8 main",
          "caching": true,
          "numDocs": 117166,
          "maxDoc": 117305,
          "reader": "SolrIndexReader{this=2ee29b0,r=ReadOnlyDirectoryReader@2ee29b0,refCnt=1,segments=5}",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/usr/solrData/search/index lockFactory=org.apache.lucene.store.NativeFSLockFactory@178671d8",
          "indexVersion": 1491861981523,
          "openedAt": "2018-01-17T20:14:54.677Z",
          "registeredAt": "2018-01-17T20:14:54.679Z",
          "warmupTime": 1
        }
      },
      "core": {
        "class": "search",
        "version": "1.0",
        "description": "SolrCore",
        "srcId": "$Id: SolrCore.java 1190108 2011-10-28 01:13:25Z yonik $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/SolrCore.java $",
        "docs": null,
        "stats": {
          "coreName": "search",
          "startTime": "2018-01-16T06:15:53.152Z",
          "refCount": 2,
          "aliases": [
            "search"
          ]
        }
      },
      "Searcher@4eea69e8 main": {
        "class": "org.apache.solr.search.SolrIndexSearcher",
        "version": "1.0",
        "description": "index searcher",
        "srcId": "$Id: SolrIndexSearcher.java 1201291 2011-11-12 18:02:03Z simonw $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/SolrIndexSearcher.java $",
        "docs": null,
        "stats": {
          "searcherName": "Searcher@4eea69e8 main",
          "caching": true,
          "numDocs": 117166,
          "maxDoc": 117305,
          "reader": "SolrIndexReader{this=2ee29b0,r=ReadOnlyDirectoryReader@2ee29b0,refCnt=1,segments=5}",
          "readerDir": "org.apache.lucene.store.MMapDirectory:org.apache.lucene.store.MMapDirectory@/usr/solrData/search/index lockFactory=org.apache.lucene.store.NativeFSLockFactory@178671d8",
          "indexVersion": 1491861981523,
          "openedAt": "2018-01-17T20:14:54.677Z",
          "registeredAt": "2018-01-17T20:14:54.679Z",
          "warmupTime": 1
        }
      }
    },
    "QUERYHANDLER",
    {
      "/admin/system": {
        "class": "org.apache.solr.handler.admin.SystemInfoHandler",
        "version": "$Revision: 1067172 $",
        "description": "Get System Info",
        "srcId": "$Id: SystemInfoHandler.java 1067172 2011-02-04 12:50:14Z uschindler $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/SystemInfoHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/plugins": {
        "class": "org.apache.solr.handler.admin.PluginInfoHandler",
        "version": "$Revision: 1052938 $",
        "description": "Registry",
        "srcId": "$Id: PluginInfoHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/PluginInfoHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/file": {
        "class": "org.apache.solr.handler.admin.ShowFileRequestHandler",
        "version": "$Revision: 1146806 $",
        "description": "Admin Get File -- view config files directly",
        "srcId": "$Id: ShowFileRequestHandler.java 1146806 2011-07-14 17:01:37Z erick $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/ShowFileRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/update/javabin": {
        "class": "org.apache.solr.handler.BinaryUpdateRequestHandler",
        "version": "$Revision: 1165749 $",
        "description": "Add/Update multiple documents with javabin format",
        "srcId": "$Id: BinaryUpdateRequestHandler.java 1165749 2011-09-06 16:20:07Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/BinaryUpdateRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353158,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/luke": {
        "class": "org.apache.solr.handler.admin.LukeRequestHandler",
        "version": "$Revision: 1201265 $",
        "description": "Lucene Index Browser.  Inspired and modeled after Luke: http://www.getopt.org/luke/",
        "srcId": "$Id: LukeRequestHandler.java 1201265 2011-11-12 14:09:28Z mikemccand $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/LukeRequestHandler.java $",
        "docs": [
          "java.net.URL:http://wiki.apache.org/solr/LukeRequestHandler"
        ],
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/dataimport": {
        "class": "org.apache.solr.handler.dataimport.DataImportHandler",
        "version": "1.0",
        "description": "Manage data import from databases to Solr",
        "srcId": "$Id: DataImportHandler.java 1171306 2011-09-15 22:43:33Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/contrib/dataimporthandler/src/java/org/apache/solr/handler/dataimport/DataImportHandler.java $",
        "docs": null,
        "stats": [
          "Status",
          "IDLE",
          "Documents Processed",
          "java.util.concurrent.atomic.AtomicLong:1",
          "Requests made to DataSource",
          "java.util.concurrent.atomic.AtomicLong:2",
          "Rows Fetched",
          "java.util.concurrent.atomic.AtomicLong:2",
          "Documents Deleted",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Documents Skipped",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Total Documents Processed",
          "java.util.concurrent.atomic.AtomicLong:351705",
          "Total Requests made to DataSource",
          "java.util.concurrent.atomic.AtomicLong:1438",
          "Total Rows Fetched",
          "java.util.concurrent.atomic.AtomicLong:876393",
          "Total Documents Deleted",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Total Documents Skipped",
          "java.util.concurrent.atomic.AtomicLong:0",
          "handlerStart",
          1516083353155,
          "requests",
          2442,
          "errors",
          0,
          "timeouts",
          0,
          "totalTime",
          1748,
          "avgTimePerRequest",
          0.7158067,
          "avgRequestsPerSecond",
          0.017792022
        ]
      },
      "/update": {
        "class": "org.apache.solr.handler.XmlUpdateRequestHandler",
        "version": "$Revision: 1165749 $",
        "description": "Add documents with XML",
        "srcId": "$Id: XmlUpdateRequestHandler.java 1165749 2011-09-06 16:20:07Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/XmlUpdateRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353157,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/terms": {
        "class": "Lazy[solr.SearchHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.SearchHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "org.apache.solr.handler.XmlUpdateRequestHandler": {
        "class": "org.apache.solr.handler.XmlUpdateRequestHandler",
        "version": "$Revision: 1165749 $",
        "description": "Add documents with XML",
        "srcId": "$Id: XmlUpdateRequestHandler.java 1165749 2011-09-06 16:20:07Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/XmlUpdateRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353157,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "org.apache.solr.handler.PingRequestHandler": {
        "class": "org.apache.solr.handler.PingRequestHandler",
        "version": "$Revision: 1142180 $",
        "description": "Reports application health to a load-balancer",
        "srcId": "$Id: PingRequestHandler.java 1142180 2011-07-02 09:04:29Z uschindler $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353163,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/threads": {
        "class": "org.apache.solr.handler.admin.ThreadDumpHandler",
        "version": "$Revision: 1052938 $",
        "description": "Thread Dump",
        "srcId": "$Id: ThreadDumpHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/ThreadDumpHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "org.apache.solr.handler.BinaryUpdateRequestHandler": {
        "class": "org.apache.solr.handler.BinaryUpdateRequestHandler",
        "version": "$Revision: 1165749 $",
        "description": "Add/Update multiple documents with javabin format",
        "srcId": "$Id: BinaryUpdateRequestHandler.java 1165749 2011-09-06 16:20:07Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/BinaryUpdateRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353158,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "org.apache.solr.handler.dataimport.DataImportHandler": {
        "class": "org.apache.solr.handler.dataimport.DataImportHandler",
        "version": "1.0",
        "description": "Manage data import from databases to Solr",
        "srcId": "$Id: DataImportHandler.java 1171306 2011-09-15 22:43:33Z janhoy $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/contrib/dataimporthandler/src/java/org/apache/solr/handler/dataimport/DataImportHandler.java $",
        "docs": null,
        "stats": [
          "Status",
          "IDLE",
          "Documents Processed",
          "java.util.concurrent.atomic.AtomicLong:1",
          "Requests made to DataSource",
          "java.util.concurrent.atomic.AtomicLong:2",
          "Rows Fetched",
          "java.util.concurrent.atomic.AtomicLong:2",
          "Documents Deleted",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Documents Skipped",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Total Documents Processed",
          "java.util.concurrent.atomic.AtomicLong:351705",
          "Total Requests made to DataSource",
          "java.util.concurrent.atomic.AtomicLong:1438",
          "Total Rows Fetched",
          "java.util.concurrent.atomic.AtomicLong:876393",
          "Total Documents Deleted",
          "java.util.concurrent.atomic.AtomicLong:0",
          "Total Documents Skipped",
          "java.util.concurrent.atomic.AtomicLong:0",
          "handlerStart",
          1516083353155,
          "requests",
          2442,
          "errors",
          0,
          "timeouts",
          0,
          "totalTime",
          1748,
          "avgTimePerRequest",
          0.7158067,
          "avgRequestsPerSecond",
          0.017792022
        ]
      },
      "/analysis/field": {
        "class": "Lazy[solr.FieldAnalysisRequestHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.FieldAnalysisRequestHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/browse": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "$Revision: 1052938 $",
        "description": "Search using components: org.apache.solr.handler.component.QueryComponent,org.apache.solr.handler.component.FacetComponent,org.apache.solr.handler.component.MoreLikeThisComponent,org.apache.solr.handler.component.HighlightComponent,org.apache.solr.handler.component.StatsComponent,org.apache.solr.handler.component.SpellCheckComponent,org.apache.solr.handler.component.DebugComponent,",
        "srcId": "$Id: SearchHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353156,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/ping": {
        "class": "org.apache.solr.handler.PingRequestHandler",
        "version": "$Revision: 1142180 $",
        "description": "Reports application health to a load-balancer",
        "srcId": "$Id: PingRequestHandler.java 1142180 2011-07-02 09:04:29Z uschindler $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/PingRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353163,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/admin/mbeans": {
        "class": "org.apache.solr.handler.admin.SolrInfoMBeanHandler",
        "version": "$Revision: 1065312 $",
        "description": "Get Info (and statistics) about all registered SolrInfoMBeans",
        "srcId": "$Id: SolrInfoMBeanHandler.java 1065312 2011-01-30 16:08:25Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/SolrInfoMBeanHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 1078,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 547,
          "avgTimePerRequest": 0.50742114,
          "avgRequestsPerSecond": 0.00785414
        }
      },
      "/analysis/document": {
        "class": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.DocumentAnalysisRequestHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "search": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "$Revision: 1052938 $",
        "description": "Search using components: org.apache.solr.handler.component.QueryComponent,org.apache.solr.handler.component.FacetComponent,org.apache.solr.handler.component.MoreLikeThisComponent,org.apache.solr.handler.component.HighlightComponent,org.apache.solr.handler.component.StatsComponent,org.apache.solr.handler.component.DebugComponent,",
        "srcId": "$Id: SearchHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353156,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/update/csv": {
        "class": "Lazy[solr.CSVRequestHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.CSVRequestHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/update/json": {
        "class": "Lazy[solr.JsonUpdateRequestHandler]",
        "version": "$Revision: 1086822 $ :: $Revision: 1102081 $",
        "description": "Add documents with JSON",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $ :: $Id: JsonUpdateRequestHandler.java 1102081 2011-05-11 20:37:04Z yonik $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $\n$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/JsonUpdateRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516103486630,
          "requests": 2530,
          "errors": 26,
          "timeouts": 0,
          "totalTime": 132438,
          "avgTimePerRequest": 52.347034,
          "avgRequestsPerSecond": 0.02160195
        }
      },
      "/admin/": {
        "class": "org.apache.solr.handler.admin.AdminHandlers",
        "version": "$Revision: 953887 $",
        "description": "Register Standard Admin Handlers",
        "srcId": "$Id: AdminHandlers.java 953887 2010-06-11 21:53:43Z hossman $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "docs": null,
        "stats": null
      },
      "standard": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "$Revision: 1052938 $",
        "description": "Search using components: org.apache.solr.handler.component.QueryComponent,org.apache.solr.handler.component.FacetComponent,org.apache.solr.handler.component.MoreLikeThisComponent,org.apache.solr.handler.component.HighlightComponent,org.apache.solr.handler.component.StatsComponent,org.apache.solr.handler.component.DebugComponent,",
        "srcId": "$Id: SearchHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353155,
          "requests": 11480,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 318753,
          "avgTimePerRequest": 27.765942,
          "avgRequestsPerSecond": 0.08364145
        }
      },
      "org.apache.solr.handler.admin.AdminHandlers": {
        "class": "org.apache.solr.handler.admin.AdminHandlers",
        "version": "$Revision: 953887 $",
        "description": "Register Standard Admin Handlers",
        "srcId": "$Id: AdminHandlers.java 953887 2010-06-11 21:53:43Z hossman $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/AdminHandlers.java $",
        "docs": null,
        "stats": null
      },
      "tvrh": {
        "class": "Lazy[solr.SearchHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.SearchHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "org.apache.solr.handler.DumpRequestHandler": {
        "class": "org.apache.solr.handler.DumpRequestHandler",
        "version": "$Revision: 1067172 $",
        "description": "Dump handler (debug)",
        "srcId": "$Id: DumpRequestHandler.java 1067172 2011-02-04 12:50:14Z uschindler $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353163,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/update/extract": {
        "class": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.extraction.ExtractingRequestHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/admin/properties": {
        "class": "org.apache.solr.handler.admin.PropertiesRequestHandler",
        "version": "$Revision: 898152 $",
        "description": "Get System Properties",
        "srcId": "$Id: PropertiesRequestHandler.java 898152 2010-01-12 02:19:56Z ryan $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/admin/PropertiesRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353227,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "org.apache.solr.handler.component.SearchHandler": {
        "class": "org.apache.solr.handler.component.SearchHandler",
        "version": "$Revision: 1052938 $",
        "description": "Search using components: org.apache.solr.handler.component.QueryComponent,org.apache.solr.handler.component.FacetComponent,org.apache.solr.handler.component.MoreLikeThisComponent,org.apache.solr.handler.component.HighlightComponent,org.apache.solr.handler.component.StatsComponent,org.apache.solr.handler.component.SpellCheckComponent,org.apache.solr.handler.component.DebugComponent,",
        "srcId": "$Id: SearchHandler.java 1052938 2010-12-26 20:21:48Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/component/SearchHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353156,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      },
      "/spell": {
        "class": "Lazy[solr.SearchHandler]",
        "version": "$Revision: 1086822 $",
        "description": "Lazy[solr.SearchHandler]",
        "srcId": "$Id: RequestHandlers.java 1086822 2011-03-30 02:23:07Z koji $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/core/RequestHandlers.java $",
        "docs": null,
        "stats": {
          "note": "not initialized yet"
        }
      },
      "/debug/dump": {
        "class": "org.apache.solr.handler.DumpRequestHandler",
        "version": "$Revision: 1067172 $",
        "description": "Dump handler (debug)",
        "srcId": "$Id: DumpRequestHandler.java 1067172 2011-02-04 12:50:14Z uschindler $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/handler/DumpRequestHandler.java $",
        "docs": null,
        "stats": {
          "handlerStart": 1516083353163,
          "requests": 0,
          "errors": 0,
          "timeouts": 0,
          "totalTime": 0,
          "avgTimePerRequest": "NaN",
          "avgRequestsPerSecond": 0
        }
      }
    },
    "UPDATEHANDLER",
    {
      "updateHandler": {
        "class": "org.apache.solr.update.DirectUpdateHandler2",
        "version": "1.0",
        "description": "Update handler that efficiently directly updates the on-disk main lucene index",
        "srcId": "$Id: DirectUpdateHandler2.java 1203770 2011-11-18 17:55:52Z mikemccand $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/update/DirectUpdateHandler2.java $",
        "docs": null,
        "stats": {
          "commits": 3220,
          "autocommits": 0,
          "optimizes": 3,
          "rollbacks": 0,
          "expungeDeletes": 0,
          "docsPending": 0,
          "adds": 0,
          "deletesById": 0,
          "deletesByQuery": 0,
          "errors": 0,
          "cumulative_adds": 354209,
          "cumulative_deletesById": 0,
          "cumulative_deletesByQuery": 3,
          "cumulative_errors": 0
        }
      }
    },
    "CACHE",
    {
      "queryResultCache": {
        "class": "org.apache.solr.search.LRUCache",
        "version": "1.0",
        "description": "LRU Cache(maxSize=512, initialSize=512)",
        "srcId": "$Id: LRUCache.java 1065312 2011-01-30 16:08:25Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/LRUCache.java $",
        "docs": null,
        "stats": {
          "lookups": 4,
          "hits": 2,
          "hitratio": "0.50",
          "inserts": 2,
          "evictions": 0,
          "size": 2,
          "warmupTime": 0,
          "cumulative_lookups": 10630,
          "cumulative_hits": 5509,
          "cumulative_hitratio": "0.51",
          "cumulative_inserts": 5626,
          "cumulative_evictions": 0
        }
      },
      "fieldCache": {
        "class": "org.apache.solr.search.SolrFieldCacheMBean",
        "version": "1.0",
        "description": "Provides introspection of the Lucene FieldCache, this is **NOT** a cache that is managed by Solr.",
        "srcId": "$Id: SolrFieldCacheMBean.java 984594 2010-08-11 21:42:04Z yonik $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/SolrFieldCacheMBean.java $",
        "docs": null,
        "stats": {
          "entries_count": 174,
          "entry#0": "'MMapIndexInput(path=\"/usr/solrData/search/index/_9eir.frq\")'=>'latlng_0_coordinate',double,org.apache.lucene.search.FieldCache.NUMERIC_UTILS_DOUBLE_PARSER=>[D#661647869",
	   "insanity_count": 1,
           "insanity#0": "SUBREADER: Found caches for descendants of ReadOnlyDirectoryReader(segments_1wo _3kl(3.5):C133115/12 _3kw(3.5):C17/2 _3kx(3.5):C6 _3ky(3.5):C1 _3kz(3.5):C2 _3l0(3.5):C2 _3l1(3.5):C1 _3l2(3.5):C1 _3l3(3.5):C1 _3l4(3.5):C1)+owner\n\t'ReadOnlyDirectoryReader(segments_1wo _3kl(3.5):C133115/12 _3kw(3.5):C17/2 _3kx(3.5):C6 _3ky(3.5):C1 _3kz(3.5):C2 _3l0(3.5):C2 _3l1(3.5):C1 _3l2(3.5):C1 _3l3(3.5):C1 _3l4(3.5):C1)'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#927712538\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3kx.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#969886745\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3kz.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#495952608\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3ky.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1581258843\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3l1.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#359550090\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3kl.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1748227582\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3l4.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1084424163\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3l3.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1116912780\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3l0.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1187916045\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3l2.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#62119827\n\t'MMapIndexInput(path=\"/usr/solrData/search/index/_3kw.frq\")'=>'owner',class org.apache.lucene.search.FieldCache$StringIndex,null=>org.apache.lucene.search.FieldCache$StringIndex#1756606907\n"
        }
      },
      "documentCache": {
        "class": "org.apache.solr.search.LRUCache",
        "version": "1.0",
        "description": "LRU Cache(maxSize=512, initialSize=512)",
        "srcId": "$Id: LRUCache.java 1065312 2011-01-30 16:08:25Z rmuir $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/LRUCache.java $",
        "docs": null,
        "stats": {
          "lookups": 0,
          "hits": 0,
          "hitratio": "0.00",
          "inserts": 0,
          "evictions": 0,
          "size": 0,
          "warmupTime": 0,
          "cumulative_lookups": 180435,
          "cumulative_hits": 22584,
          "cumulative_hitratio": "0.12",
          "cumulative_inserts": 157851,
          "cumulative_evictions": 40344
        }
      },
      "fieldValueCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=10000, initialSize=10, minSize=9000, acceptableSize=9500, cleanupThread=false)",
        "srcId": "$Id: FastLRUCache.java 1170772 2011-09-14 19:09:56Z sarowe $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "docs": null,
        "stats": {
          "lookups": 5,
          "hits": 3,
          "hitratio": "0.60",
          "inserts": 1,
          "evictions": 0,
          "size": 1,
          "warmupTime": 0,
          "cumulative_lookups": 8529,
          "cumulative_hits": 5432,
          "cumulative_hitratio": "0.63",
          "cumulative_inserts": 1437,
          "cumulative_evictions": 0,
          "item_parentCompanyId": "{field=parentCompanyId,memSize=785156,tindexSize=13056,time=136,phase1=135,nTerms=75696,bigTerms=0,termInstances=117166,uses=4}"
        }
      },
      "filterCache": {
        "class": "org.apache.solr.search.FastLRUCache",
        "version": "1.0",
        "description": "Concurrent LRU Cache(maxSize=512, initialSize=512, minSize=460, acceptableSize=486, cleanupThread=false)",
        "srcId": "$Id: FastLRUCache.java 1170772 2011-09-14 19:09:56Z sarowe $",
        "src": "$URL: https://svn.apache.org/repos/asf/lucene/dev/branches/lucene_solr_3_5/solr/core/src/java/org/apache/solr/search/FastLRUCache.java $",
        "docs": null,
        "stats": {
          "lookups": 2,
          "hits": 2,
          "hitratio": "1.00",
          "inserts": 2,
          "evictions": 0,
          "size": 2,
          "warmupTime": 0,
          "cumulative_lookups": 4041,
          "cumulative_hits": 4041,
          "cumulative_hitratio": "1.00",
          "cumulative_inserts": 2828,
          "cumulative_evictions": 0
        }
      }
    }
  ]
}
`

var solr3CoreExpected = map[string]interface{}{
	"num_docs":     int64(117166),
	"max_docs":     int64(117305),
	"deleted_docs": int64(0),
}

var solr3QueryHandlerExpected = map[string]interface{}{
	"15min_rate_reqs_per_second": float64(0),
	"5min_rate_reqs_per_second":  float64(0),
	"75th_pc_request_time":       float64(0),
	"95th_pc_request_time":       float64(0),
	"999th_pc_request_time":      float64(0),
	"99th_pc_request_time":       float64(0),
	"avg_requests_per_second":    float64(0),
	"avg_time_per_request":       float64(0),
	"errors":                     int64(0),
	"handler_start":              int64(1516083353156),
	"median_request_time":        float64(0),
	"requests":                   int64(0),
	"timeouts":                   int64(0),
	"total_time":                 float64(0),
}

var solr3UpdateHandlerExpected = map[string]interface{}{
	"adds":                        int64(0),
	"autocommit_max_docs":         int64(0),
	"autocommit_max_time":         int64(0),
	"autocommits":                 int64(0),
	"commits":                     int64(3220),
	"cumulative_adds":             int64(354209),
	"cumulative_deletes_by_id":    int64(0),
	"cumulative_deletes_by_query": int64(3),
	"cumulative_errors":           int64(0),
	"deletes_by_id":               int64(0),
	"deletes_by_query":            int64(0),
	"docs_pending":                int64(0),
	"errors":                      int64(0),
	"expunge_deletes":             int64(0),
	"optimizes":                   int64(3),
	"rollbacks":                   int64(0),
	"soft_autocommits":            int64(0),
}

var solr3CacheExpected = map[string]interface{}{
	"cumulative_evictions": int64(0),
	"cumulative_hitratio":  float64(1.00),
	"cumulative_hits":      int64(4041),
	"cumulative_inserts":   int64(2828),
	"cumulative_lookups":   int64(4041),
	"evictions":            int64(0),
	"hitratio":             float64(1.00),
	"hits":                 int64(2),
	"inserts":              int64(2),
	"lookups":              int64(2),
	"size":                 int64(2),
	"warmup_time":          int64(0),
}
