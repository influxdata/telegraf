# Google Core IoT service input plugin

The Google Core IoT listener is a service input plugin that listens for messages sent via HTTP POST from Google's Pub/Sub service.
The plugin expects messages in Google's Pub/Sub JSON Format ONLY.
The intent of the plugin is to allow Telegraf to serve as an endpoint of the Google Pub/Sub 'Push' service.
This service will **only** run over HTTPS/TLS

Enable TLS by specifying the file names of a service TLS certificate and key.

The data payload sent to Google Core IoT's MQTT broker can be in either Influx Line Protocol or in JSON with the following format: 

```json
{"measurement": "measurementName", 
	"tags": {
	"tag_1": "tag_value",
	"tag_2": "tag_value",
	...
    }, 
	"fields": {
    	"field_1": value,
		"field_2" : value
		}, "time": unix_time_stamp 
}
```

Enable mutually authenticated TLS and authorize client connections by signing certificate authority by including a list of allowed CA certificate file names in ````tls_allowed_cacerts````.


**Example:**
```
An example program for sending data from a Google Core IoT device (Raspberry Pi) will be provided
```

### Configuration:

This is a sample configuration for the plugin.

```toml
  ## Server listens on <server name>:port/write
  ## Address and port to host HTTP listener on
  service_address = ":9999"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  # precision of the time stamps. can be one of the following:
  # second, millisecond, microsecond, nanosecond
  # Default is nanosecond
  
  precision = "nanosecond"
  
  # Data Format is either line protocol or json
  protocol="line protocol" 

  ## Set one or more allowed client CA certificate file names to 
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

```
