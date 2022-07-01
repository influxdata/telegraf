# Example showing how to obtain IOPS (to aggregate, to produce max_iops).  Input can be produced by:
#
#[[inputs.diskio]]
#  alias = "diskio1s"
#  interval = "1s"
#  fieldpass = ["reads", "writes"]
#  name_suffix = "1s"
#
# Example Input:
# diskio1s,host=hostname,name=diska reads=0i,writes=0i 1554079521000000000
# diskio1s,host=hostname,name=diska reads=0i,writes=0i 1554079522000000000
# diskio1s,host=hostname,name=diska reads=110i,writes=0i 1554079523000000000
# diskio1s,host=hostname,name=diska reads=110i,writes=30i 1554079524000000000
# diskio1s,host=hostname,name=diska reads=160i,writes=70i 1554079525000000000
#
# Example Output:
# diskiops,host=hostname,name=diska readsps=0,writesps=0,iops=0 1554079522000000000
# diskiops,host=hostname,name=diska readsps=110,writesps=0,iops=110 1554079523000000000
# diskiops,host=hostname,name=diska readsps=0,writesps=30,iops=30 1554079524000000000
# diskiops,host=hostname,name=diska readsps=50,writesps=40,iops=90 1554079525000000000

state = { }

def apply(metric):
    disk_name = metric.tags["name"]
    # Load from the shared last_state the metric for the disk name
    last = state.get(disk_name)
    # Store the deepcopy of the new metric into the shared last_state and assign it to the key "last"
    # NB: To store a metric into the shared last_state you have to deep copy it
    state[disk_name] = deepcopy(metric)
    if last != None:
        # Create the new metrics
        diskiops = Metric("diskiops")
        # Calculate reads/writes per second
        reads = metric.fields["reads"] - last.fields["reads"]
        writes = metric.fields["writes"] - last.fields["writes"]
        io = reads + writes
        interval_seconds = ( metric.time - last.time ) / 1000000000
        diskiops.fields["readsps"] = ( reads / interval_seconds )
        diskiops.fields["writesps"] = ( writes / interval_seconds )
        diskiops.fields["iops"] = ( io / interval_seconds )
        diskiops.tags["name"] = disk_name
        diskiops.tags["host"] = metric.tags["host"]
        diskiops.time = metric.time
        return diskiops

# This could be aggregated to obtain max IOPS using:
#
# [[aggregators.basicstats]]
#  namepass = ["diskiops"]
#  period = "60s"
#  drop_original = true
#  stats = ["max"]
#
# diskiops,host=hostname,name=diska readsps_max=110,writesps_max=40,iops_max=110 1554079525000000000
