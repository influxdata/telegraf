# particle webhooks

You should configure your Rollbar's Webhooks to point at the `webhooks` service. To do this go to `particle.com/` and click `Settings > Notifications > Webhook`. In the resulting page set `URL` to `http://<my_ip>:1619/particle`, and click on `Enable Webhook Integration`.

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
Escaping the "" is required in the source file.
The number of tag values and field values is not restrictied so you can send as many values per webhook call as you'd like.

You will need to enable JSON messages in the Webhooks setup of Particle.io

See [webhook doc](https://docs.particle.io/reference/webhooks/)

