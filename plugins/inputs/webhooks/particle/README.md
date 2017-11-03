# particle webhooks

You should configure your Particle.io's Webhooks to point at the `webhooks` service. To do this go to `(https://console.particle.io/)[https://console.particle.io]` and click `Integrations > New Integration > Webhook`. In the resulting page set `URL` to `http://<my_ip>:1619/particle`, and  under `Advanced Settings` click on `JSON` and add:

```
{
    "influx_db": "your_measurement_name"
}
```

If required, enter your username and password, etc. and then click `Save`


## Events

Your Particle device should publish an event that contains a JSON in the form of:
```
String data = String::format("{ \"tags\" : {
	    \"tag_name\": \"tag_value\", 
	    \"other_tag\": \"other_value\"
    }, 
	\"values\": {
	    \"value_name\": %f, 
		\"other_value\": %f, 
    }
    }",  value_value, other_value
	);
    Particle.publish("event_name", data, PRIVATE);
```
Escaping the "" is required in the source file on the Particle device.
The number of tag values and field values is not restricted so you can send as many values per webhook call as you'd like.



See [webhook doc](https://docs.particle.io/reference/webhooks/)
