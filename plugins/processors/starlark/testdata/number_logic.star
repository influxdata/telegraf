# Set  any 'status' field between 1 and 6 to a value of 0

def apply(metric):
    v = metric.fields.get('status')
    if v == None:
        return metric
    if 1 < v and v < 6:
        metric.fields['status'] = 0
    return metric
