# Fibaro Input Plugin

The Fibaro plugin makes HTTP calls to the Fibaro controller API to gather values of hooked devices.
Those values could be true (1) or false (0) for switches, percentage for dimmers, temperature, etc.

### Configuration:

```toml
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

### Metrics:

- fibaro
  - tags:
    - section (section name)
    - room (room name)
    - name (device name)
    - type (device type)
  - fields:
    - value (float)
    - value2 (float, when available from device)


### Example Output:

```
fibaro,host=vm1,name=Escaliers,room=Dégagement,section=Pièces\ communes,type=com.fibaro.binarySwitch value=0 1523351010000000000
fibaro,host=vm1,name=Porte\ fenêtre,room=Salon,section=Pièces\ communes,type=com.fibaro.FGRM222 value=99,value2=99 1523351010000000000
fibaro,host=vm1,name=LED\ îlot\ central,room=Cuisine,section=Cuisine,type=com.fibaro.binarySwitch value=0 1523351010000000000
fibaro,host=vm1,name=Détérioration,room=Entrée,section=Pièces\ communes,type=com.fibaro.heatDetector value=0 1523351010000000000
fibaro,host=vm1,name=Température,room=Cave,section=Cave,type=com.fibaro.temperatureSensor value=17.87 1523351010000000000
fibaro,host=vm1,name=Présence,room=Garde-manger,section=Cuisine,type=com.fibaro.FGMS001 value=1 1523351010000000000
fibaro,host=vm1,name=Luminosité,room=Garde-manger,section=Cuisine,type=com.fibaro.lightSensor value=92 1523351010000000000
fibaro,host=vm1,name=Etat,room=Garage,section=Extérieur,type=com.fibaro.doorSensor value=0 1523351010000000000
fibaro,host=vm1,name=CO2\ (ppm),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=880 1523351010000000000
fibaro,host=vm1,name=Humidité\ (%),room=Salon,section=Pièces\ communes,type=com.fibaro.humiditySensor value=53 1523351010000000000
fibaro,host=vm1,name=Pression\ (mb),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=1006.9 1523351010000000000
fibaro,host=vm1,name=Bruit\ (db),room=Salon,section=Pièces\ communes,type=com.fibaro.multilevelSensor value=58 1523351010000000000
```
