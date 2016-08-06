# Particle Webhook
[Particle](https://www.particle.io) Webhook Plug-in for Telegraf

Particle events can be sent to Webhooks. Webhooks are configured using [Integrations](https://dashboard.particle.io/user/integrations).
Assuming your Particle publishes an event called MY_EVENT, you could configure an Integration with:

1. "Event Name" set to MY_EVENT
2. "URL" set to http://MY_IP:1619/particle
3. Leave "Request Type" set to "POST"
4. Leave "Device" set to "Any"

Replace MY_EVENT and MY_IP appropriately.

Then click "Create Webhook".

You may watch the stream of Particle events including the hook-sent/MY_EVENT and hook-response/MY_EVENT entries in the [Logs](https://dashboard.particle.io/user/logs)

## Particle

See Particle [Webhooks](https://docs.particle.io/guide/tools-and-features/webhooks/) documentation.

The default data is:

```
{
    "event": MY_EVENT,
    "data": MY_EVENT_DATA,
    "published_at": MY_EVENT_TIMESTAMP,
    "coreid": DEVICE_ID
}
```

The following Particle (trivial) sample publishes a random number as an event called "randomnumber" every 10 seconds:

```
void loop() {
    Particle.publish("randomnumber", String(random(1000)), PRIVATE);
    delay(10000);
}
```

## Events

**Tags:**
* 'event' = `event` string
* 'coreid' = `coreid` string

**Fields:**
* 'data' = `data` int

**Time:**
* 'published' = `published_at` time.Time ([ISO-8601](https://en.wikipedia.org/wiki/ISO_8601))