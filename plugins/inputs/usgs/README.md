# USGS Telegraf plugin

This plugin gathers the recent earthquake data from the USGS and turns it into Telegraf metric format. The JSON polled from USGS looks as follows:

```json
{
  "type": "FeatureCollection",
  "metadata": {
    "generated": 1481144380000,
    "url": "http://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson",
    "title": "USGS All Earthquakes, Past Hour",
    "status": 200,
    "api": "1.5.2",
    "count": 4
  },
  "features": [
    {
      "type": "Feature",
      "properties": {
        "mag": 1.82,
        "place": "15km ENE of Hawaiian Ocean View, Hawaii",
        "time": 1481143731250,
        "updated": 1481143943070,
        "tz": -600,
        "url": "http://earthquake.usgs.gov/earthquakes/eventpage/hv61510176",
        "detail": "http://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/hv61510176.geojson",
        "felt": null,
        "cdi": null,
        "mmi": null,
        "alert": null,
        "status": "automatic",
        "tsunami": 0,
        "sig": 51,
        "net": "hv",
        "code": "61510176",
        "ids": ",hv61510176,",
        "sources": ",hv,",
        "types": ",general-link,geoserve,origin,phase-data,",
        "nst": 32,
        "dmin": 0.07161,
        "rms": 0.24,
        "gap": 106,
        "magType": "md",
        "type": "earthquake",
        "title": "M 1.8 - 15km ENE of Hawaiian Ocean View, Hawaii"
      },
      "geometry": {
        "type": "Point",
        "coordinates": [
          -155.6236725,
          19.1058331,
          0.87
        ]
      },
      "id": "hv61510176"
    }
  ],
  "bbox": [
    -155.6236725,
    19.1058331,
    0.87,
    -117.025,
    64.9877,
    13.47
  ]
}
```

Each `Feature` is then converted into a point in InfluxDB:

```yaml
measurement: "earthquakes"
tags:
- magnitude: 1.82,
- url: "http://earthquake.usgs.gov/earthquakes/eventpage/hv61510176",
- detail: "http://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/hv61510176.geojson",
- felt: null,
- cdi: null,
- mmi: null,
- alert: null,
- status: "automatic",
- tsunami: 0,
- sig: 51,
- net: "hv",
- nst: 32,
- dmin: 0.07161,
- rms: 0.24,
- gap: 106,
- magType: "md",
- type: "earthquake",
- title: "M 1.8 - 15km ENE of Hawaiian Ocean View, Hawaii"
```