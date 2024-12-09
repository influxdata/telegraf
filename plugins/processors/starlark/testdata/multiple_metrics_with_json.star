# Example showing how to create several metrics from a json array.
#
# Example Input:
# json value="[{\"label\": \"hello\"}, {\"label\": \"world\"}]"
#
# Example Output:
# json value="hello" 1618488000000000999
# json value="world" 1618488000000000999

# loads json.encode(), json.decode(), json.indent()
load("json.star", "json")
load("time.star", "time")

def apply(metric):
    # Initialize a list of metrics
    metrics = []
    # Loop over the json array stored into the field 
    for obj in json.decode(metric.fields['value']):
        # Create a new metric whose name is "json"
        current_metric = Metric("json")
        # Set the field "value" to the label extracted from the current json object
        current_metric.fields["value"] = obj["label"]
        # Reset the time (only needed for testing purpose)
        current_metric.time = time.now().unix_nano
        # Add metric to the list of metrics
        metrics.append(current_metric)
    return metrics
