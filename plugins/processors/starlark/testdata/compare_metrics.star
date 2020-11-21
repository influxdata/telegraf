# Example showing how to keep the last metric in order to compare it with the new one.
#
# Example Input:
# cpu value=10i,i=10i,f=2.35,s="before" 1465839830100400201
# cpu value=8i,i=20i,f=1.23,s="after" 1465839830100400301
#
# Example Output:
# cpu_diff value=2i,i=-10i,f=1.12,s="before" 1465839830100400301

def apply(metric):
    # Load from the shared state the metric assigned to the key "last"
    last = Load("last")
    # Load from the shared state the integer assigned to the key "ilast"
    ilast = Load("ilast")
    # Load from the shared state the float assigned to the key "flast"
    flast = Load("flast")
    # Load from the shared state the string assigned to the key "slast"
    slast = Load("slast")    
    # Store the new metric into the shared state and assign it to the key "last"
    Store("last", metric)
    # Store the value of the field "i" into the shared state and assign it to the key "ilast"
    Store("ilast", metric.fields["i"])
    # Store the value of the field "f" into the shared state and assign it to the key "flast"
    Store("flast", metric.fields["f"])
    # Store the value of the field "s" into the shared state and assign it to the key "slast"
    Store("slast", metric.fields["s"])
    if last != None:
        # Create a new metric named "cpu_diff"
        result = Metric("cpu_diff")
        # Set the field "value" to the difference between the value of the last metric and the current one
        result.fields["value"] = last.fields["value"] - metric.fields["value"]
        # Set the field "i" to the difference between ilast and the value of the field "i" of the current metric
        result.fields["i"] = ilast - metric.fields["i"]
        # Set the field "f" to the difference between flast and the value of the field "f" of the current metric
        result.fields["f"] = flast - metric.fields["f"]
        # Set the field "s" to the value of slast
        result.fields["s"] = slast
        return result
