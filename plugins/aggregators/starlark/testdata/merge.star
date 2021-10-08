# Example of a merge aggregator implemented with a starlark script.

load('time.star', 'time')

def add(cache, metric):
    metrics = cache.get("metrics")
    if metrics == None:
        metrics = {}
        cache["metrics"] = metrics
        cache["ordered"] = []
    m = metrics.get(metric)
    if m == None:
        m = deepcopy(metric)
        metrics[metric] = m 
        cache["ordered"].append(m)
    else:
        for k, v in metric.fields.items():
            m.fields[k] = v

def apply(cache):
    return cache.get("ordered")
