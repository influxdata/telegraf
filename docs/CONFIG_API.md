# Config API

## Endpoints

### GET /plugins/list

List all known plugins with default config. Each plugin is listed once.

**request params**

None

**response**

An array of plugin-config schemas.

eg:
```json
[
  {
    "name": "mqtt_consumer",
    "config": {
      "servers": {
        "type": "[string]", // another example: Map[string, SomeSchema]
        "default": ["http://127.0.0.1"],
        "required": true,
      },
      "topics": {
        "type": "[string]",
        "default": [
          "telegraf/host01/cpu",
          "telegraf/+/mem",
          "sensors/#",
        ],
        "required": true,
      },
      "topic_tag": {
        "type": "string",
        "default": "topic",
      },
      "username": {
        "type": "string",
        "required": false,
      },
      "password": {
        "type": "string",
        "required": false,
      },
      "qos": {
        "type": "integer",
        "format": "int64",
      },
      "connection_timeout": {
        "type": "integer",
        "format": "duration"
      },
      "max_undelivered_messages": {
        "type": "integer",
        "format": "int32",
      }
    }
  },
  // ...
]
```

### GET /plugins/running

List all currently running plugins. If there are 5 copies of a plugin, all 5 will be returned.

**request params**

none

**response**

```json
[
  {
    "id": "unique-id-here",
    "name": "mqtt_consumer",
    "config": {
        "servers": ["tcp://127.0.0.1:1883"],
        "topics": [
          "telegraf/host01/cpu",
          "telegraf/+/mem",
          "sensors/#",
        ],
        "topic_tag": "topic",
        "qos": 0,
        "connection_timeout": 300000000000,
        "max_undelivered_messages": 1000,
        "persistent_session": false,
        "client_id": "",
        "username": "telegraf",
        // password not returned
        "tls_ca": "/etc/telegraf/ca.pem",
        "tls_cert": "/etc/telegraf/cert.pem",
        "tls_key": "/etc/telegraf/key.pem",
        "insecure_skip_verify": false,
        "data_format": "influx",
    },
  },
]
```

### POST /plugins/create

Create a new plugin. It will be started upon creation.

**request params**

```json
  {
    "name": "mqtt_consumer",
    "config": {.
      // ..
    },
  },
```

**response**

```json
  {"id": "unique-id-here"}
```

### GET /plugins/{id}/status

Get the status of a launched plugin

**request params**

None. ID in url

**response**

```json
  {
    "status": "", // starting, running, notfound, or error
    "reason": "", // extended reason code containing error details.
  }
```

### DELETE /plugins/{id}

Stop an existing running plugin given its `id`. It will be allowed to finish
any metrics in-progress.

**request params**

None

**response**

200 OK
```json
{}
```

## Schemas

### plugin-config

A plugin-config is a plugin name and details about the config fields.

```
  {
    name: string,
    config: Map[string, FieldConfig]
  }
```

### FieldConfig

```
  {
    type: string, // eg "string", "integer", "[string]", or "Map[string, SomeSchema]"
    default: object, // whatever the default value is
    required: bool,
    format: string, // type-specific format info.
  }
```

### plugin

An instance of a plugin running with a specific configuration

```
{
  id: string,
  name: string,
  config: Map[string, object],
}
```