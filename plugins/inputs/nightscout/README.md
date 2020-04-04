# Nightscout Input Plugin

The nightscout plugin collects blood glucose related telemtry from a nightscout site. Not all available data is collected at this time, simply the basics needed for active monitoring.

### Configuration:

```toml
  ## This is a list of nightscout sites to pull data from.
  ##  example: servers = ["https://my.nightscout.site/api/v1/entries"]
  servers = []
  ## The person is the owner of the nightscout site.
  # owner = ""

  ## The secret is the SHA-1 hash of the api secret used to access the site.
  # secret = ""

  ## The count is the number of results to pull at a time. Unless you have a specific need, this should remain as 1.
  # count = "1"

  ## The token is the unique token that matches a role within your nightscout site(s).
  token = ""
```

### Collected Measurements:

- Type
- DateString
- Date
- SGV
- Direction
- Noise
- Filtered
- Unfiltered
- RSSI

### Tags:

- All measurements have the following tags:
    - owner

### Example Output:

```
nightscout,host=localhost,owner=bob noise=0i,filtered=0i,rssi=0i,direction_num="-1",dateString="2020-04-04T14:12:51.000Z",date=1586009571000i,direction="FortyFiveDown",unfiltered=0i,type="sgv",sgv=277i 1586009730000000000
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
      "datasource": null,
      "gridPos": {
        "h": 8,
        "w": 11,
        "x": 6,
        "y": 9
      },
      "id": 17,
      "options": {
        "fieldOptions": {
          "calcs": [
            "delta"
          ],
          "defaults": {
            "mappings": [],
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {
                  "color": "green",
                  "value": null
                },
                {
                  "color": "red",
                  "value": 80
                }
              ]
            }
          },
          "overrides": [],
          "values": false
        },
        "orientation": "auto",
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "pluginVersion": "6.7.1",
      "targets": [
        {
          "groupBy": [],
          "measurement": "nightscout",
          "orderByTime": "ASC",
          "policy": "default",
          "query": "SELECT \"date\" FROM \"nightscout\" WHERE (\"person\" = 'batman') AND $timeFilter",
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
      "title": "Staleness",
      "type": "gauge"
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
