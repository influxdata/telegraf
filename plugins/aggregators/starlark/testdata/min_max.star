# Example of a min_max aggregator implemented with a starlark script.

supported_types = (["int", "float"])
state = {}
def add(metric):
    gId = groupID(metric)
    aggregate = state.get(gId)
    if aggregate == None:
        aggregate = {
            "name": metric.name, 
            "tags": metric.tags, 
            "fields": {}
        }
        for k, v in metric.fields.items():
            if type(v) in supported_types:
                aggregate["fields"][k] = {
				    "min": v,
				    "max": v,
			    }
        state[gId] = aggregate
    else:
        for k, v in metric.fields.items():
            if type(v) in supported_types:
                min_max = aggregate["fields"].get(k)
                if min_max == None:
                    aggregate["fields"][k] = {
				        "min": v,
				        "max": v,
			        }
                elif v < min_max["min"]:
                    aggregate["fields"][k]["min"] = v
                elif v > min_max["max"]:
                    aggregate["fields"][k]["max"] = v
        
def push():
    metrics = []
    for a in state:
        fields = {}
        for k in state[a]["fields"]:
            fields[k + "_min"] = state[a]["fields"][k]["min"]
            fields[k + "_max"] = state[a]["fields"][k]["max"]
        m = Metric(state[a]["name"], state[a]["tags"], fields)
        metrics.append(m)
    return metrics

def reset():
    state.clear()

def groupID(metric):
    key = metric.name + "-"
    for k, v in metric.tags.items():
        key = key + k + "-" + v
    return hash(key)