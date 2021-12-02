# papertrail webhooks

Enables Telegraf to act as a [Papertrail Webhook](http://help.papertrailapp.com/kb/how-it-works/web-hooks/).

## Events

[Full documentation](http://help.papertrailapp.com/kb/how-it-works/web-hooks/#callback).

Events from Papertrail come in two forms:

* The [event-based callback](http://help.papertrailapp.com/kb/how-it-works/web-hooks/#callback):

  * A point is created per event, with the timestamp as `received_at`
  * Each point has a field counter (`count`), which is set to `1` (signifying the event occurred)
  * Each event "hostname" object is converted to a `host` tag
  * The "saved_search" name in the payload is added as an `event` tag
  * The "saved_search" id in the payload is added as a `search_id` field
  * The papertrail url to view the event is built and added as a `url` field
  * The rest of the data in the event is converted directly to fields on the point:
    * `id`
    * `source_ip`
    * `source_name`
    * `source_id`
    * `program`
    * `severity`
    * `facility`
    * `message`

When a callback is received, an event-based point will look similar to:

```shell
papertrail,host=myserver.example.com,event=saved_search_name count=1i,source_name="abc",program="CROND",severity="Info",source_id=2i,message="message body",source_ip="208.75.57.121",id=7711561783320576i,facility="Cron",url="https://papertrailapp.com/searches/42?centered_on_id=7711561783320576",search_id=42i 1453248892000000000
```

* The [count-based callback](http://help.papertrailapp.com/kb/how-it-works/web-hooks/#count-only-webhooks)

  * A point is created per timeseries object per count, with the timestamp as the "timeseries" key (the unix epoch of the event)
  * Each point has a field counter (`count`), which is set to the value of each "timeseries" object
  * Each count "source_name" object is converted to a `host` tag
  * The "saved_search" name in the payload is added as an `event` tag

When a callback is received, a count-based point will look similar to:

```shell
papertrail,host=myserver.example.com,event=saved_search_name count=3i 1453248892000000000
```
