# Rename any tags using the mapping in the renames dict.
#
# Example Input:
# measurement,host=hostname lower=0,upper=100 1597255410000000000
#
# Example Output:
# measurement,host=hostname min=0,max=100 1597255410000000000

renames = {
    'lower': 'min',
    'upper': 'max',
}

def apply(metric):
    for k, v in metric.tags.items():
        if k in renames:
            metric.tags[renames[k]] = v
            metric.tags.pop(k)
    for k, v in metric.fields.items():
        if k in renames:
            metric.fields[renames[k]] = v
            metric.fields.pop(k)
    return metric
