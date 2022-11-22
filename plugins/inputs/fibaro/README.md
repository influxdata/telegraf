# Fibaro Input Plugin

The Fibaro plugin makes HTTP calls to the Fibaro controller API to gather values
of hooked devices. Those values could be true (1) or false (0) for switches,
percentage for dimmers, temperature, etc.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Read devices value(s) from a Fibaro controller
[[inputs.fibaro]]
  ## Required Fibaro controller address/hostname.
  ## Note: at the time of writing this plugin, Fibaro only implemented http - no https available
  url = "http://<controller>:80"

  ## Required credentials to access the API (http://<controller/api/<component>)
  username = "<username>"
  password = "<password>"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"
```

## Metrics

- fibaro
  - tags:
    - deviceId (device id)
    - section (section name)
    - room (room name)
    - name (device name)
    - type (device type)
  - fields:
    - batteryLevel (float, when available from device)
    - energy (float, when available from device)
    - power (float, when available from device)
    - value (float)
    - value2 (float, when available from device)

## Example Output

```shell
fibaro,deviceId=9,host=vm1,name=Fenêtre\ haute,room=Cuisine,section=Cuisine,type=com.fibaro.FGRM222 energy=2.04,power=0.7,value=99,value2=99 1529996807000000000
fibaro,deviceId=10,host=vm1,name=Escaliers,room=Dégagement,section=Pièces\ communes,type=com.fibaro.binarySwitch value=0 1529996807000000000
fibaro,deviceId=13,host=vm1,name=Porte\ fenêtre,room=Salon,section=Pièces\ communes,type=com.fibaro.FGRM222 energy=4.33,power=0.7,value=99,value2=99 1529996807000000000
fibaro,deviceId=21,host=vm1,name=LED\ îlot\ central,room=Cuisine,section=Cuisine,type=com.fibaro.binarySwitch value=0 1529996807000000000
fibaro,deviceId=90,host=vm1,name=Détérioration,room=Entrée,section=Pièces\ communes,type=com.fibaro.heatDetector value=0 1529996807000000000
fibaro,deviceId=163,host=vm1,name=Température,room=Cave,section=Cave,type=com.fibaro.temperatureSensor value=21.62 1529996807000000000
fibaro,deviceId=191,host=vm1,name=Présence,room=Garde-manger,section=Cuisine,type=com.fibaro.FGMS001 value=1 1529996807000000000
fibaro,deviceId=193,host=vm1,name=Luminosité,room=Garde-manger,section=Cuisine,type=com.fibaro.lightSensor value=195 1529996807000000000
fibaro,deviceId=200,host=vm1,name=Etat,room=Garage,section=Extérieur,type=com.fibaro.doorSensor value=0 1529996807000000000
fibaro,deviceId=220,host=vm1,name=CO2\ (ppm),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=536 1529996807000000000
fibaro,deviceId=221,host=vm1,name=Humidité\ (%),room=Salon,section=Pièces\ communes,type=com.fibaro.humiditySensor value=61 1529996807000000000
fibaro,deviceId=222,host=vm1,name=Pression\ (mb),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=1013.7 1529996807000000000
fibaro,deviceId=223,host=vm1,name=Bruit\ (db),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=44 1529996807000000000
fibaro,deviceId=248,host=vm1,name=Température,room=Garage,section=Extérieur,type=com.fibaro.temperatureSensor batteryLevel=85,value=10.8 1529996807000000000
```
