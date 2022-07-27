# Apache IoTDB

[![Main Mac and Linux](https://github.com/apache/iotdb/actions/workflows/main-unix.yml/badge.svg)](https://github.com/apache/iotdb/actions/workflows/main-unix.yml)
[![Main Win](https://github.com/apache/iotdb/actions/workflows/main-win.yml/badge.svg)](https://github.com/apache/iotdb/actions/workflows/main-win.yml)
[![coveralls](https://coveralls.io/repos/github/apache/iotdb/badge.svg?branch=master)](https://coveralls.io/repos/github/apache/iotdb/badge.svg?branch=master)
[![GitHub release](https://img.shields.io/github/release/apache/iotdb.svg)](https://github.com/apache/iotdb/releases)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
![size](https://github-size-badge.herokuapp.com/apache/iotdb.svg)
![downloads](https://img.shields.io/github/downloads/apache/iotdb/total.svg)
![platform](https://img.shields.io/badge/platform-win10%20%7C%20macox%20%7C%20linux-yellow.svg)
![java-language](https://img.shields.io/badge/java--language-1.8-blue.svg)
[![Language grade: Java](https://img.shields.io/lgtm/grade/java/g/apache/iotdb.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/apache/iotdb/context:java)
[![IoTDB Website](https://img.shields.io/website-up-down-green-red/https/shields.io.svg?label=iotdb-website)](https://iotdb.apache.org/)
[![Maven Version](https://maven-badges.herokuapp.com/maven-central/org.apache.iotdb/iotdb-parent/badge.svg)](http://search.maven.org/#search|gav|1|g:"org.apache.iotdb")
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://join.slack.com/t/apacheiotdb/shared_invite/zt-qvso1nj8-7715TpySZtZqmyG5qXQwpg)

Apache IoTDB（物联网数据库）是一个物联网原生数据库，在数据管理和分析方面表现良好，可部署在边缘设备和云上。
由于其轻量级架构、高性能和丰富的功能集，以及与Apache Hadoop、Spark和Flink的深度集成，
Apache IoTDB可以满足物联网工业领域的海量数据存储、高速数据摄取和复杂数据分析的要求。

Apache IoTDB website: <https://iotdb.apache.org>
Apache IoTDB Github: <https://github.com/apache/iotdb>

# Apache IoTDB Go语言客户端

[![E2E Tests](https://github.com/apache/iotdb-client-go/actions/workflows/e2e.yml/badge.svg)](https://github.com/apache/iotdb-client-go/actions/workflows/e2e.yml)
[![GitHub release](https://img.shields.io/github/release/apache/iotdb-client-go.svg)](https://github.com/apache/iotdb-client-go/releases)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
![size](https://github-size-badge.herokuapp.com/apache/iotdb-client-go.svg)
![platform](https://img.shields.io/badge/platform-win10%20%7C%20macos%20%7C%20linux-yellow.svg)
[![IoTDB Website](https://img.shields.io/website-up-down-green-red/https/shields.io.svg?label=iotdb-website)](https://iotdb.apache.org/)

Apache IoTDB 有一个go语言客户端，能够使用go语言原生接口支持 IoTDB 的数据增删改查。

Apache IoTDB Golang Client Github: <https://github.com/apache/iotdb>

# IoTDB 输出插件

[English](./README.md) | [中文](./README_ZH.md)

IoTDB 输出插件可以把 Telegraf 采集到的数据保存到IoTDB数据库。该插件使用了go语言客户端的接口，能够支持会话连接、数据插入。

## 快速上手

使用该插件前，需要配置数据库服务器的ip地址、所使用的端口号、用户名、密码等信息，以及一些数据类型转换、时间单位等配置。

英文的配置文件请参考：[English Configuration](./sample.conf)，中文配置文件请参考[中文配置样例](./sample_zh.conf). 或者，对应版本的配置内容也在后文中列出。

## 注意事项

1. IoTDB 0.13.x版本以及之前的版本，**不支持无符号整数**。所以本插件提供了三种可选的无符号整数处理方式，只需要指定参数`convertUint64To`的取值即可。该参数有三个取值，分别对应不同的处理方式，分别是：
   - `ToInt64`，默认的处理方式。对于未超出`int64`表示范围的无符号整数，以`int64`类型存储；如果超出表示范围，则保存`math.MaxInt64`，也即9223372036854775807。
   - `ForceToInt64`，强制类型转换为`int64`。如果数字超过`int64`表示范围，可能会抛出异常。
   - `Text`，强制转换为字符串。无论无符号整型多大，都会被转换为字符串保存，不丢失精度。

2. IoTDB支持多种时间精度，但无论何种精度，都以`int64`类型存储，所以用户需要指定时间戳的语义。用户需要指定参数`convertUint64To`的取值，该参数默认取值为`nanosecond`。

3. IoTDB目前不能很好地支持标签（Tag）索引，目前采用的处理方式请参考[InfluxDB-Protocol适配器](https://iotdb.apache.org/zh/UserGuide/Master/API/InfluxDB-Protocol.html)。用户需要指定参数`treateTagsAs`的取值，来决定如何处理标签：

   - `Measurements`，Tag会被看做一个普通的物理量，等同于Field。只不过Tag的取值总是字符串。
   - `DeviceID_subtree`，默认的处理方式。Tag会被看做设备标识路径（device id）的一部分。Tags的顺序是有序的，该顺序由Telegraf决定，一般为字典序升序排列。

   举例：当一个metric的取值为，`Name="root.sg.device", Tags={tag1="private", tag2="working"}, Fields={s1=100, s2="hello"}`。此时不同参数对应的处理结果为：

   - `Measurements`，处理结果：`root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"`
   - `DeviceID_subtree`，处理结果：`root.sg.device.private.working, s1=100, s2="hello"`

## 测试

本插件自带测试。但测试前**首先需要开启IoTDB数据库**，默认地址为`localhost:6667`。若要修改测试地址，可以修改`iotdb_test.go`中的`test_host`等全局变量。

测试内容主要包括：数据库连接、配置纠错、数据类型转换、数据写入等。

## 配置文件

```properties
# 将采集到的数据保存到IoTDB
[[outputs.iotdb]]
  ## IoTDB 服务器配置
  ## host 是IoTDB服务器的ip地址
  ## port 是IoTDB服务器的端口号
  host = "127.0.0.1"
  port = "6667"

  ## 认证配置
  ## user是用户名，默认是'root', password是密码，默认是'root'
  user = "root"
  password = "root"

  ## 会话相关配置
  ## timeout是连接超时时间，单位是毫秒(ms)。该数值类型必须是int，0表示不设定超时。
  ## 负数会被当做0来看待。
  timeout = 5000

  ## 无符号整型的转换配置
  ## IoTDB 不支持无符号整数(版本 13.x).
  ## 对 uint32, 此插件会直接将其转换为 int64.
  ## 但是遇到 unit64 类型的数字时，该操作可能导致溢出，所以用户必须制定下面的一个可选的转换方案。
  ## 
  ## 本插件支持3种转换uint64的方式 : 
  ## - "ForceToInt64": 无论其数值多大，强制转换为 int64.
  ## - "ToInt64"(默认): 如果数字比 MAXINT64 小, 转换为 int64; 否则，转换为 MAXINT64。
  ##              math.MaxInt64 = 9223372036854775807
  ## - "Text": 无论数值多大，总是转换为字符串保存。在IoTDB中，字符串类型称为TEXT.
  convertUint64To = "ToInt64"

  ## 时间戳(timestamp)的相关配置
  ## 时间戳总是以int64的形式存储。timeStampUnit 指定了时间戳使用的单位. 如下是该变量的可用值:
  ## "second", "millisecond", "microsecond", "nanosecond"(默认)
  timeStampUnit = "nanosecond"

  ## 处理标签(Tags)的相关配置
  ## IoTDB不完全支持Tag的索引，但是也有为了兼容InfluxDB-Protocol所设计的解决方案，可以看这里：
  ##     https://iotdb.apache.org/zh/UserGuide/Master/API/InfluxDB-Protocol.html
  ## 
  ## 本插件提供两种可用的方式来处理标签：
  ## - "Measurements": 将标签看做物理量(measurements). 每个标签都有Key到Value的对构成，
  ##                        所以把Key当做被测物理量的名称，Value当做其取值，转化为
  ##                        Measurement, Value, DataType.
  ## - "DeviceID_subtree"(默认): 把标签看做设备标识号(deviceID)的组成部分. 标签是'Name'的子树.
  ##
  ## 例如，当一个metric的取值为:
  ##      Name="root.sg.device", Tags={tag1="private", tag2="working"}, Fields={s1=100, s2="hello"}
  ## - 在"Measurements"模式下得到的record为:
  ##      root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"
  ## - 在"DeviceID_subtree"模式下得到的record为:
  ##      root.sg.device.private.working, s1=100, s2="hello"
  treateTagsAs = "DeviceID_subtree"

```
