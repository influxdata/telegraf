# Multiply any float fields by 10

def apply(metric):
    for k, v in metric.fields.items():
        if type(v) == "float":
            metric.fields[k] = v * 10
    return metric
