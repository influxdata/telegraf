# Example of a merge aggregator implemented with a starlark script.
load('time.star', 'time')
state = {}
def add(metric):
    metrics = state.get("metrics")
    if metrics == None:
        metrics = {}
        state["metrics"] = metrics
        state["ordered"] = []
    gId = groupID(metric)
    m = metrics.get(gId)
    if m == None:
        m = deepcopy(metric)
        metrics[gId] = m 
        state["ordered"].append(m)
    else:
        for k, v in metric.fields.items():
            m.fields[k] = v

def push():
    return state.get("ordered")

def reset():
    state.clear()

def groupID(metric):
    key = metric.name + "-"
    for k, v in metric.tags.items():
        key = key + k + "-" + v + "-"
    key = key + "-" + str(metric.time)
    return hash(key)