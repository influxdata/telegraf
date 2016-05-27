# rollbar_webhooks

This is a Telegraf service plugin that listens for events kicked off by Rollbar Webhooks service and persists data from them into configured outputs. To set up the listener first generate the proper configuration:
```sh
$ telegraf -sample-config -input-filter rollbar_webhooks -output-filter influxdb > config.conf.new
```
Change the config file to point to the InfluxDB server you are using and adjust the settings to match your environment. Once that is complete:
```sh
$ cp config.conf.new /etc/telegraf/telegraf.conf
$ sudo service telegraf start
```
Once the server is running you should configure your Rollbar's Webhooks to point at the `rollbar_webhooks` service. To do this go to `rollbar.com/` and click `Settings > Notifications > Webhook`. In the resulting page set `URL` to `http://<my_ip>:1619`, and click on `Enable Webhook Integration`.

## Events

The titles of the following sections are links to the full payloads and details for each event. The body contains what information from the event is persisted. The format is as follows:
```
# TAGS
* 'tagKey' = `tagValue` type
# FIELDS
* 'fieldKey' = `fieldValue` type
```
The tag values and field values show the place on the incoming JSON object where the data is sourced from.

See [webhook doc](https://rollbar.com/docs/webhooks/)

#### `new_item` event

**Tags:**
* 'event' = `event.event_name` string
* 'environment' = `event.data.item.environment` string
* 'project_id = `event.data.item.project_id` int
* 'language' = `event.data.item.last_occurence.language` string
* 'level' = `event.data.item.last_occurence.level` string

**Fields:**
* 'id' = `event.data.item.id` int

#### `deploy` event

**Tags:**
* 'event' = `event.event_name` string
* 'environment' = `event.data.deploy.environment` string
* 'project_id = `event.data.deploy.project_id` int

**Fields:**
* 'id' = `event.data.item.id` int
