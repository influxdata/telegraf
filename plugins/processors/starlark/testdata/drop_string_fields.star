# Drop fields if they contain a string.
#
# Example Input:
# measurement,host=hostname a=1,b="somestring" 1597255410000000000
#
# Example Output:
# measurement,host=hostname a=1 1597255410000000000

def apply(metric):
    for k, v in metric.fields.items():
        if type(v) == "string":
            metric.fields.pop(k)

    return metric
