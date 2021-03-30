# Rename the measurement name to the fieldname. Rename the fieldname to value. 
#
# Example Input:
# prometheus_remote_write,instance=localhost:9090,job=prometheus,quantile=0.99 go_gc_duration_seconds=4.63 1614889298859000000
#
# Example Output:
# go_gc_duration_seconds,instance=localhost:9090,job=prometheus,quantile=0.99 value=4.63 1614889299000000000

def apply(metric):
   for k, v in metric.fields.items():
       metric.name = k
       metric.fields["value"] = v
       metric.fields.pop(k)
   return metric