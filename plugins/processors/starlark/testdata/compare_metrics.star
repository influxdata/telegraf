# Example showing how to keep the last metric in order to compare it with the new one.
#
# Example Input:
# cpu value=10i 1465839830100400201
# cpu value=8i 1465839830100400301
#
# Example Output:
# cpu_diff value=2i 1465839830100400301

def apply(metric):
    # Load from the shared state the metric assigned to the key "last"
    last = Load("last")
    # Store the new metric into the shared state and assign it to the key "last"
    Store("last", metric)
    if last != None:
        # Create a new metric named "cpu_diff"
        result = Metric("cpu_diff")
        # Set the field "value" to the difference between the value of the last metric and the current one
        result.fields["value"] = last.fields["value"] - metric.fields["value"]
        return result
