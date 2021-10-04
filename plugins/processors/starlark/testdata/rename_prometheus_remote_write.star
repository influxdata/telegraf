# Specifically for prometheus remote write - renames the measurement name to the fieldname. Renames the fieldname to value. 
# Assumes there is only one field as is the case for prometheus remote write. 
#
# Example Input:
# prometheus_remote_write,instance=localhost:9090,job=prometheus,quantile=0.99 go_gc_duration_seconds=4.63 1618488000000000999
#
# Example Output:
# go_gc_duration_seconds,instance=localhost:9090,job=prometheus,quantile=0.99 value=4.63 1618488000000000999

def apply(metric):
   if metric.name == "prometheus_remote_write":
        for k, v in metric.fields.items():
            metric.name = k
            metric.fields["value"] = v
            metric.fields.pop(k)
   return metric