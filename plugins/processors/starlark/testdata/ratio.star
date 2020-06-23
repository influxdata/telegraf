# Compute the ratio of two integer fields.

def apply(metric):
    used = float(metric.fields['used'])
    total = float(metric.fields['total'])
    metric.fields['usage'] = (used / total) * 100
    return metric
