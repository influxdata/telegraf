# Example of a min_max aggregator implemented with a starlark script.

supported_types = (["int", "float"])

def add(cache, metric):
    aggregate = cache.get(metric)
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
        cache[metric] = aggregate
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
        
def push(cache, accumulator):
    for a in cache:
        fields = {}
        for k in cache[a]["fields"]:
            fields[k + "_min"] = cache[a]["fields"][k]["min"]
            fields[k + "_max"] = cache[a]["fields"][k]["max"]
        accumulator.add_fields(cache[a]["name"], fields, cache[a]["tags"])