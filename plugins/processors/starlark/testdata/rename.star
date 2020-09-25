# Rename any tags using the mapping in the renames dict.

renames = {
    'lower': 'min',
    'upper': 'max',
}

def apply(metric):
    for k, v in metric.tags.items():
        if k in renames:
            metric.tags[renames[k]] = v
            metric.tags.pop(k)
    return metric
