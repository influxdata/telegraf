# Filter metrics by value
'''
In this example we look at the `value` field of the metric.
If the value is zeor, we delete all the fields, effectively dropping the metric.

Example Input:
temperature sensor="001A0",value=111.48 1618488000000000999
temperature sensor="001B0",value=0.0 1618488000000000999

Example Output:
temperature sensor="001A0",value=111.48 1618488000000000999
'''

def apply(metric):
    if metric.fields["value"] == 0.0:
        # removing all fields deletes a metric
        metric.fields.clear()
    return metric
