# Apache Input Plugin

The Apache plugin collects server performance information using the [`mod_status`](https://httpd.apache.org/docs/2.4/mod/mod_status.html) module of the [Apache HTTP Server](https://httpd.apache.org/).

Typically, the `mod_status` module is configured to expose a page at the `/server-status?auto` location of the Apache server.  The [ExtendedStatus](https://httpd.apache.org/docs/2.4/mod/core.html#extendedstatus) option must be enabled in order to collect all available fields.  For information about how to configure your server reference the [module documentation](https://httpd.apache.org/docs/2.4/mod/mod_status.html#enable).

### Configuration:

```toml
# Read Apache status information (mod_status)
[[inputs.apache]]
  ## An array of URLs to gather from, must be directed at the machine
  ## readable version of the mod_status page including the auto query string.
  ## Default is "http://localhost/server-status?auto".
  urls = ["http://localhost/server-status?auto"]

  ## Credentials for basic HTTP authentication.
  # username = "myuser"
  # password = "mypassword"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Measurements & Fields:

- apache
  - BusyWorkers (float)
  - BytesPerReq (float)
  - BytesPerSec (float)
  - ConnsAsyncClosing (float)
  - ConnsAsyncKeepAlive (float)
  - ConnsAsyncWriting (float)
  - ConnsTotal (float)
  - CPUChildrenSystem (float)
  - CPUChildrenUser (float)
  - CPULoad (float)
  - CPUSystem (float)
  - CPUUser (float)
  - IdleWorkers (float)
  - Load1 (float)
  - Load5 (float)
  - Load15 (float)
  - ParentServerConfigGeneration (float)
  - ParentServerMPMGeneration (float)
  - ReqPerSec (float)
  - ServerUptimeSeconds (float)
  - TotalAccesses (float)
  - TotalkBytes (float)
  - Uptime (float)

The following fields are collected from the `Scoreboard`, and represent the number of requests in the given state:

- apache
  - scboard_closing (float)
  - scboard_dnslookup (float)
  - scboard_finishing (float)
  - scboard_idle_cleanup (float)
  - scboard_keepalive (float)
  - scboard_logging (float)
  - scboard_open (float)
  - scboard_reading (float)
  - scboard_sending (float)
  - scboard_starting (float)
  - scboard_waiting (float)

### Tags:

- All measurements have the following tags:
    - port
    - server

### Example Output:

```
apache,port=80,server=debian-stretch-apache BusyWorkers=1,BytesPerReq=0,BytesPerSec=0,CPUChildrenSystem=0,CPUChildrenUser=0,CPULoad=0.00995025,CPUSystem=0.01,CPUUser=0.01,ConnsAsyncClosing=0,ConnsAsyncKeepAlive=0,ConnsAsyncWriting=0,ConnsTotal=0,IdleWorkers=49,Load1=0.01,Load15=0,Load5=0,ParentServerConfigGeneration=3,ParentServerMPMGeneration=2,ReqPerSec=0.00497512,ServerUptimeSeconds=201,TotalAccesses=1,TotalkBytes=0,Uptime=201,scboard_closing=0,scboard_dnslookup=0,scboard_finishing=0,scboard_idle_cleanup=0,scboard_keepalive=0,scboard_logging=0,scboard_open=100,scboard_reading=0,scboard_sending=1,scboard_starting=0,scboard_waiting=49 1502489900000000000
```

### Example Grafana Dashboard

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 1,
  "links": [],
  "panels": [
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": null,
      "fill": 2,
      "fillGradient": 0,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "hiddenSeries": false,
      "id": 15,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "nullPointMode": "null",
      "options": {
        "dataLinks": []
      },
      "percentage": false,
      "pointradius": 2,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "groupBy": [
            {
              "params": [
                "$__interval"
              ],
              "type": "time"
            },
            {
              "params": [
                "null"
              ],
              "type": "fill"
            }
          ],
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "direction_num"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "distinct"
              }
            ]
          ],
          "tags": [
            {
              "key": "person",
              "operator": "=",
              "value": "batman"
            }
          ]
        }
      ],
      "thresholds": [
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "gt",
          "value": 2,
          "yaxis": "left"
        },
        {
          "colorMode": "warning",
          "fill": true,
          "line": true,
          "op": "gt",
          "value": 1,
          "yaxis": "left"
        },
        {
          "colorMode": "warning",
          "fill": true,
          "line": true,
          "op": "lt",
          "value": -1,
          "yaxis": "left"
        },
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "lt",
          "value": -2,
          "yaxis": "left"
        }
      ],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Trending Arrows",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": null,
      "fill": 1,
      "fillGradient": 0,
      "gridPos": {
        "h": 8,
        "w": 11,
        "x": 12,
        "y": 0
      },
      "hiddenSeries": false,
      "id": 6,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "nullPointMode": "null",
      "options": {
        "dataLinks": []
      },
      "percentage": false,
      "pointradius": 2,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "groupBy": [
            {
              "params": [
                "$__interval"
              ],
              "type": "time"
            },
            {
              "params": [
                "null"
              ],
              "type": "fill"
            }
          ],
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "sgv"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "distinct"
              }
            ]
          ],
          "tags": [
            {
              "key": "person",
              "operator": "=",
              "value": "batman"
            }
          ]
        }
      ],
      "thresholds": [
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "gt",
          "value": 270,
          "yaxis": "left"
        },
        {
          "colorMode": "warning",
          "fill": true,
          "line": true,
          "op": "gt",
          "value": 150,
          "yaxis": "left"
        },
        {
          "colorMode": "warning",
          "fill": true,
          "line": true,
          "op": "lt",
          "value": 80,
          "yaxis": "left"
        },
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "lt",
          "value": 70,
          "yaxis": "left"
        }
      ],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Trending Numbers",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 8
      },
      "id": 4,
      "panels": [],
      "title": "Batman",
      "type": "row"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": true,
      "colors": [
        "#C4162A",
        "#37872D",
        "#C4162A"
      ],
      "datasource": null,
      "format": "none",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 0,
        "y": 9
      },
      "id": 8,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "pluginVersion": "6.7.1",
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": false,
        "lineColor": "rgb(31, 120, 193)",
        "show": false,
        "ymax": null,
        "ymin": null
      },
      "tableColumn": "",
      "targets": [
        {
          "groupBy": [],
          "hide": false,
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "table",
          "select": [
            [
              {
                "params": [
                  "direction_num"
                ],
                "type": "field"
              }
            ]
          ],
          "tags": [
            {
              "key": "person",
              "operator": "=",
              "value": "batman"
            }
          ]
        }
      ],
      "thresholds": "-1,1",
      "timeFrom": null,
      "timeShift": null,
      "title": "Direction",
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "Flat",
          "value": "0"
        },
        {
          "op": "=",
          "text": "Single Up",
          "value": "2"
        },
        {
          "op": "=",
          "text": "Double Up",
          "value": "3"
        },
        {
          "op": "=",
          "text": "Forty-Five Up",
          "value": "1"
        },
        {
          "op": "=",
          "text": "Single Down",
          "value": "-2"
        },
        {
          "op": "=",
          "text": "Double Down",
          "value": "-3"
        },
        {
          "op": "=",
          "text": "Forty-Five Down",
          "value": "-1"
        }
      ],
      "valueName": "total"
    },
    {
      "cacheTimeout": null,
      "datasource": null,
      "gridPos": {
        "h": 5,
        "w": 7,
        "x": 8,
        "y": 9
      },
      "id": 19,
      "links": [],
      "options": {
        "colorMode": "value",
        "fieldOptions": {
          "calcs": [
            "lastNotNull"
          ],
          "defaults": {
            "mappings": [
              {
                "id": 0,
                "op": "=",
                "text": "N/A",
                "type": 1,
                "value": "null"
              }
            ],
            "nullValueMode": "connected",
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {
                  "color": "rgb(255, 255, 255)",
                  "value": null
                }
              ]
            },
            "unit": "dateTimeFromNow"
          },
          "overrides": [],
          "values": false
        },
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "horizontal"
      },
      "pluginVersion": "6.7.1",
      "targets": [
        {
          "groupBy": [],
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "query": "SELECT distinct(\"date\") FROM \"nightscout\" WHERE (\"person\" = 'batman') AND \"date\" - (time.now - 3600)",
          "rawQuery": true,
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "date"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "distinct"
              }
            ]
          ],
          "tags": [
            {
              "key": "person",
              "operator": "=",
              "value": "batman"
            }
          ]
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Data Freshness",
      "type": "stat"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": true,
      "colors": [
        "#C4162A",
        "#37872D",
        "#C4162A"
      ],
      "datasource": null,
      "format": "none",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 17,
        "y": 9
      },
      "id": 2,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "270",
          "text": "N/A",
          "to": "500"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": false,
        "lineColor": "rgb(31, 120, 193)",
        "show": false,
        "ymax": null,
        "ymin": null
      },
      "tableColumn": "",
      "targets": [
        {
          "groupBy": [],
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "sgv"
                ],
                "type": "field"
              }
            ]
          ],
          "tags": [
            {
              "key": "person",
              "operator": "=",
              "value": "batman"
            }
          ]
        }
      ],
      "thresholds": "80,270",
      "timeFrom": null,
      "timeShift": null,
      "title": "Current Blood Sugar",
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    },
    {
      "collapsed": true,
      "datasource": null,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 17
      },
      "id": 11,
      "panels": [
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": true,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "minValue": 0,
            "show": false,
            "thresholdLabels": false,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 8,
            "w": 6,
            "x": 0,
            "y": 1
          },
          "id": 9,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "pluginVersion": "6.7.1",
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "full": false,
            "lineColor": "rgb(31, 120, 193)",
            "show": false,
            "ymax": null,
            "ymin": null
          },
          "tableColumn": "",
          "targets": [
            {
              "groupBy": [
                {
                  "params": [
                    "$__interval"
                  ],
                  "type": "time"
                },
                {
                  "params": [
                    "null"
                  ],
                  "type": "fill"
                }
              ],
              "hide": false,
              "measurement": "nightscout",
              "orderByTime": "ASC",
              "policy": "default",
              "refId": "A",
              "resultFormat": "table",
              "select": [
                [
                  {
                    "params": [
                      "direction_num"
                    ],
                    "type": "field"
                  },
                  {
                    "params": [],
                    "type": "distinct"
                  }
                ]
              ],
              "tags": []
            }
          ],
          "thresholds": "",
          "timeFrom": null,
          "timeShift": null,
          "title": "Direction",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "Flat",
              "value": "0"
            },
            {
              "op": "=",
              "text": "Single Up",
              "value": "1"
            },
            {
              "op": "=",
              "text": "Double Up",
              "value": "2"
            },
            {
              "op": "=",
              "text": "Forty-Five Up",
              "value": "3"
            },
            {
              "op": "=",
              "text": "Single Down",
              "value": "-1"
            },
            {
              "op": "=",
              "text": "Double Down",
              "value": "-2"
            },
            {
              "op": "=",
              "text": "Forty-Five Down",
              "value": "-3"
            }
          ],
          "valueName": "first"
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": null,
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 8,
            "w": 11,
            "x": 6,
            "y": 1
          },
          "hiddenSeries": false,
          "id": 12,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pointradius": 2,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "groupBy": [
                {
                  "params": [
                    "$__interval"
                  ],
                  "type": "time"
                },
                {
                  "params": [
                    "null"
                  ],
                  "type": "fill"
                }
              ],
              "measurement": "nightscout",
              "orderByTime": "ASC",
              "policy": "default",
              "refId": "A",
              "resultFormat": "time_series",
              "select": [
                [
                  {
                    "params": [
                      "sgv"
                    ],
                    "type": "field"
                  },
                  {
                    "params": [],
                    "type": "mean"
                  }
                ]
              ],
              "tags": [
                {
                  "key": "person",
                  "operator": "=",
                  "value": "batman"
                }
              ]
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "Time Lapse",
          "tooltip": {
            "shared": true,
            "sort": 0,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            },
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": true,
          "colors": [
            "#C4162A",
            "#37872D",
            "#C4162A"
          ],
          "datasource": null,
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "minValue": 0,
            "show": false,
            "thresholdLabels": false,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 8,
            "w": 6,
            "x": 17,
            "y": 1
          },
          "id": 13,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "270",
              "text": "N/A",
              "to": "500"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "full": false,
            "lineColor": "rgb(31, 120, 193)",
            "show": false,
            "ymax": null,
            "ymin": null
          },
          "tableColumn": "",
          "targets": [
            {
              "groupBy": [
                {
                  "params": [
                    "$__interval"
                  ],
                  "type": "time"
                },
                {
                  "params": [
                    "null"
                  ],
                  "type": "fill"
                }
              ],
              "measurement": "nightscout",
              "orderByTime": "ASC",
              "policy": "default",
              "refId": "A",
              "resultFormat": "time_series",
              "select": [
                [
                  {
                    "params": [
                      "sgv"
                    ],
                    "type": "field"
                  },
                  {
                    "params": [],
                    "type": "distinct"
                  }
                ]
              ],
              "tags": []
            }
          ],
          "thresholds": "80,270",
          "timeFrom": null,
          "timeShift": null,
          "title": "Current Blood Sugar",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        }
      ],
      "title": "Indy",
      "type": "row"
    }
  ],
  "refresh": "30s",
  "schemaVersion": 22,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ]
  },
  "timezone": "",
  "title": "Hacker Daycare",
  "uid": "G4U55MrWz",
  "variables": {
    "list": []
  },
  "version": 18
}
```
