# Example of a merge aggregator implemented with a starlark script.

load('time.star', 'time')
state = {}
def add(metric):
    metrics = state.get("metrics")
    if metrics == None:
        metrics = {}
        state["metrics"] = metrics
        state["ordered"] = []
    m = metrics.get(metric)
    if m == None:
        m = deepcopy(metric)
        metrics[metric] = m 
        state["ordered"].append(m)
    else:
        for k, v in metric.fields.items():
            m.fields[k] = v

def push():
    return state.get("ordered")

def reset():
  state.clear()