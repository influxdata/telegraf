###############################################################################
#                                 DESCRIPTION                                 #
###############################################################################
#
# Possible configurations:
# -Metrics without a watchdog link will replicate last value forever
# -Metrics with a watchdog that has not timed out will be replicated
# -Metrics with a watchdog that has timed out will halt replication until new update of watchdog
# -Metrics with an unknown watchdog link will be added but not replicated
#     -A watchdog without timeout tag will be discarded after the aggregation period (timeout is considered as the aggregation period)
#
# All metrics passed into this aggregator will be replicated every period.
# You can configure multiple replication aggregators with different intervals.
#
# Optional there are watchdog metrics that can be passed into the aggregator and linked to the other metrics.
# Watchdog metrics will not be replicated
# Using this watchdog ensures no replication when communication has stopped or the datasource is frozen.
# This is needed because communication failure is not passed through an aggregator and this is the only way to capture this.

# When the watchdog timeout occurs, the replication of metrics is halted and will resume when the watchdog metric is refreshed.
# Old values are not removed on timeout because freezing and unfreezing the datasource might not trigger an update of the values.
# If there is a reconnection event, an update will be triggered autmatically.
#
# Example of tags from metric to be replicated:
# default_tags = [replication = "5s"]

# Optional watchdog:
# Example of tags of a watchdog metric
# default_tags = [template = "watchdog", timeout = "10s"]
# Example of tags from a metric with linked watchdog
# default_tags = [replication = "5s", watchdog_name = "plc_watchdog", watchdog_field = "counter"]

###############################################################################
#                          EXAMPLE CONFIGURATION                              #
###############################################################################
#[[aggregators.starlark]]
#  # Alias for debugging purposes
#  alias = "replication 5s"
#  # The replication interval
#  period = "5s"
#  # Replication script
#  script = "./replication.star"
#  
#  [aggregators.starlark.tagpass]
#    # This tag adds every metric marked with the correct replication rate [optional].
#    replication = ["5s"]
#    # Make sure the watchdogs are passed in as well [optional].
#    template = ["watchdog"]
#    
#  [aggregators.starlark.constants]
#    # ALL THESE CONSTANTS ARE OPTIONAL TO USE BUT HAVE TO BE INCLUDED FOR THE SCRIPT TO RUN.
#    # Optional watchdog configuration
#    # This is a tag and value to identify a watchdog metric. 
#    # Metrics marked with this tag are stored in a different manner for easy lookup (only name and field are used to identify the watchdog).
#    # Within a single replication aggregator there can be multiple watchdog variables
#    wd_identifier_tag = "template"
#    wd_identifier_value = "watchdog"
#    
#    # Optional custom timeout
#    # To specify a timeout of the watchdog that is greater than the replication interval you can specify this in a timeout tag.
#    # The default tag is "timeout" but can be changed
#    wd_timeout_tag = "timeout"
#    
#    # Every metric that has to be linked to a watchdog needs to have these two tags to identify the metric by name and field.
#    # This is used to link a metric to the correct watchdog metric.
#    linked_wd_metric_name_tag = "watchdog_name"
#    linked_wd_metric_field_tag = "watchdog_field"
#
###############################################################################
#                            REPLICATION SCRIPT                               #
###############################################################################
# State object layout:
#  {
#    "watchdogs": {
#      "wd1_field1_hash": {
#        "name": "plc_watchdog",
#        "tags": {
#          "template": "watchdog",
#          "timeout": "10s",
#          "...": "..."
#        },
#        "fields": {
#          "counter": 999
#        },
#        "time": 700790400000000000
#      },
#      "wd1_field2_hash": "wd2_obj"
#    },
#    "metrics": {
#      "metric1_tags_field1_hash": {
#        "name": "LEVEL_Tank1",
#        "tags": {
#          "replication": "5s",
#          "watchdog_name": "Watchdog_Test",
#          "watchdog_field": "counter",
#          "...": "..."
#        },
#        "fields": {
#          "field1": 999
#        },
#        "time": 700790400000000000
#      },
#      "metric1_tags_field2_hash": "metric1_field2_obj",
#      "metric2_tags_field1_hash": "metric2_field1_obj",
#      "...": "..."
#    }
#  }

state = {
  "watchdogs": {},
  "metrics": {}
}

load("logging.star", "log")
load("time.star", "time")

def add(metric):
  # Store metrics with single field at a time
  for fieldname, fieldvalue in metric.fields.items():
    # Check if metric contains the watchdog identifier tag and follow up by checking it's value.
    if metric.tags.get(wd_identifier_tag) == wd_identifier_value:
      # Watchdog metric
      gId = watchdogHash(metric.name, fieldname)
      state["watchdogs"][gId] = deepcopy(metric)
      log.debug("Watchdog added/updated: " + metric.name + "." + fieldname + " gId: " + str(gId))
    else:
      # Regular metric
      gId = metricHash(metric)
      state["metrics"][gId] = deepcopy(metric)
      log.debug("Metric added/updated: " + metric.name + "." + fieldname + " gId: " + str(gId))

def push():
  metrics = []
  for metricKey, storedMetric in state["metrics"].items():
    if storedMetric.tags.get(linked_wd_metric_name_tag) == None or storedMetric.tags.get(linked_wd_metric_field_tag) == None:
      # metric has no correct combination of linked watchog tags -> always replicate
      metrics.append(storedMetric)
    else:
      gId = watchdogHash(storedMetric.tags[linked_wd_metric_name_tag], storedMetric.tags[linked_wd_metric_field_tag])
      if state["watchdogs"].get(gId) != None:
        # Watchdog metric found so replication allowed
        metrics.append(storedMetric)
  log.debug("push: "+ str(len(metrics)))
  return metrics

def reset():
  for watchdogKey, watchdogMetric in state["watchdogs"].items():
    # Discard watchdogs that do not have a timeout tag or have timed out
    if watchdogMetric.tags.get(wd_timeout_tag) == None:
      # No identifier tag found
      state["watchdogs"].pop(watchdogKey)
    else:
      # Timed out: if value of wd_timeout_tag is less than the difference between now and the last update.
      if time.parse_duration(watchdogMetric.tags[wd_timeout_tag]).nanoseconds < (time.now().unix_nano - watchdogMetric.time):
        state["watchdogs"].pop(watchdogKey)
  log.debug("Discard temporary watchdogs. Remaining: " + str(len(state["watchdogs"])))

def watchdogHash(metricName,field):
  # name-field hash for watchdogs
  key = metricName + field
  return hash(key)

def metricHash(metric):
  # name-tags-field hash for metrics
  key = metric.name
  for k, v in metric.tags.items():
    key = key + "|" + k + "-" + v
  for k, v in metric.fields.items():
    key = key + "|" + k
  return hash(key)