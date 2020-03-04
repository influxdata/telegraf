# Huawei Routers Telemetry

This Plugin read the Huawei Routers Telemetry Information. This first version only support UDP configuration. It was tested with Huawei NE40 and NE9000 series routers running V8R10.

The telemetry information is analized by https://github.com/DamRCorba/huawei_telemetry_sensors.

## Router configuration

```
telemetry
 #
 sensor-group DeviceManager
  sensor-path huawei-devm:devm/cpuInfos/cpuInfo
  sensor-path huawei-devm:devm/memoryInfos/memoryInfo
  sensor-path huawei-devm:devm/ports/port/opticalInfo
  sensor-path huawei-devm:devm/powerSupplys/powerSupply/powerEnvironments/powerEnvironment
  sensor-path huawei-devm:devm/temperatureInfos/temperatureInfo
 #
 sensor-group InterfacesStats
  sensor-path huawei-ifm:ifm/interfaces/interface/ifDynamicInfo
  sensor-path huawei-ifm:ifm/interfaces/interface/ifStatistics
  sensor-path huawei-ifm:ifm/interfaces/interface/ifStatistics/ethPortErrSts
 #
 destination-group TelemetryTest
  ipv4-address 192.168.92.204 port 8080 protocol udp
 #
 subscription TelemetryTestSubscripcion
  sensor-group DeviceManager sample-interval 300000
  sensor-group InterfacesStats sample-interval 180000
  destination-group TelemetryTest
```

# plugin configuration

```
[[inputs.huawei_routers_telemetry]]
  service_address = "udp://:8080"
```

# Influx Data Stored.

```
SELECT * FROM "huawei-ifm:ifm/interfaces/interface/ifStatistics" WHERE ("source" = 'HuaweiRouter' AND "ifName" = '"Eth-Trunk200"')

1583267858284000000 debian "Eth-Trunk200"              huawei-ifm:ifm/interfaces/interface/ifStatistics                1879172       1154727861   1194213864700 1156607033    145599231578  689075043  689075043     HuaweiRouter TelemetryTestSubscripcion
1583267138286000000 debian "Eth-Trunk200"              huawei-ifm:ifm/interfaces/interface/ifStatistics                1806192       1108928126   1147036920750 1110734318    139701973822  661178531  661178531     HuaweiRouter TelemetryTestSubscripcion
1583267318285000000 debian "Eth-Trunk200"              huawei-ifm:ifm/interfaces/interface/ifStatistics                1824441       1120117284   1158618977528 1121941725    141213273401  668083960  668083960     HuaweiRouter TelemetryTestSubscripcion
1583267498285000000 debian "Eth-Trunk200"              huawei-ifm:ifm/interfaces/interface/ifStatistics                1842683       1131374915   1170121507435 1133217598    142584018229  674810177  674810177     HuaweiRouter TelemetryTestSubscripcion
1583267678284000000 debian "Eth-Trunk200"              huawei-ifm:ifm/interfaces/interface/ifStatistics                1860925       1143170287   1182282684273 1145031212    144102362575  682047824  682047824     HuaweiRouter TelemetryTestSubscripcion

```