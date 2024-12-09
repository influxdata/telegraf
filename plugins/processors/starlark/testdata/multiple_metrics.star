# Example showing how to create several metrics using the Starlark processor.
#
# Example Input:
# mm value="a" 1465839830100400201
#
# Example Output:
# mm2 value="b" 1465839830100400201
# mm1 value="a" 1465839830100400201

def apply(metric):
    # Initialize a list of metrics
    metrics = []
    # Create a new metric whose name is "mm2"
    metric2 = Metric("mm2")
    # Set the field "value" to b
    metric2.fields["value"] = "b"
    # Reset the time (only needed for testing purpose)
    metric2.time = metric.time
    # Add metric2 to the list of metrics
    metrics.append(metric2)
    # Rename the original metric to "mm1"
    metric.name = "mm1"
    # Add metric to the list of metrics
    metrics.append(metric)    
    # Return the created list of metrics
    return metrics
