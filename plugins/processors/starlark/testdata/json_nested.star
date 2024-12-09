# 
# This code assumes the value parser with data_type='string' is used
# in the input collecting the JSON data. The entire JSON obj/doc will
# be set to a Field named `value` with which this code will work.

# JSON:
#  ```
#     {
#         "fields": {
#             "LogEndOffset": 339238,
#             "LogStartOffset": 339238,
#             "NumLogSegments": 1,
#             "Size": 0,
#             "UnderReplicatedPartitions": 0
#         },
#         "name": "partition",
#         "tags": {
#             "host": "CUD1-001559",
#             "jolokia_agent_url": "http://localhost:7777/jolokia",
#             "partition": "1",
#             "topic": "qa-kafka-connect-logs"
#         },
#         "timestamp": 1591124461
#     } ```
# 
# Example Input:
# json value="[{\"fields\": {\"LogEndOffset\": 339238, \"LogStartOffset\": 339238, \"NumLogSegments\": 1, \"Size\": 0, \"UnderReplicatedPartitions\": 0}, \"name\": \"partition\", \"tags\": {\"host\": \"CUD1-001559\", \"jolokia_agent_url\": \"http://localhost:7777/jolokia\", \"partition\": \"1\", \"topic\": \"qa-kafka-connect-logs\"}, \"timestamp\": 1591124461}]"

# Example Output:
# partition,host=CUD1-001559,jolokia_agent_url=http://localhost:7777/jolokia,partition=1,topic=qa-kafka-connect-logs LogEndOffset=339238i,LogStartOffset=339238i,NumLogSegments=1i,Size=0i,UnderReplicatedPartitions=0i 1591124461000000000


load("json.star", "json")

def apply(metric):
  j_list = json.decode(metric.fields.get('value')) # input JSON may be an arrow of objects
  metrics = []
  for obj in j_list:
    new_metric = Metric("partition") # We want a new InfluxDB/Telegraf metric each iteration
    for tag in obj["tags"].items(): # 4 Tags to iterate through
      new_metric.tags[str(tag[0])] = tag[1]
    for field in obj["fields"].items(): # 5 Fields to iterate through
      new_metric.fields[str(field[0])] = field[1]
    new_metric.time = int(obj["timestamp"] * 1e9)
    metrics.append(new_metric)
  return metrics
