# Time parsing
#
# Example Input:
# time,host=hostname t="2020-09-01T12:00:00Z",value=42i 1600955566000000000
#
# Example Output:
# time,host=hostname value=42i 1600955566000000000

load("time.star", "time")

def apply(metric):
    t_in = metric.fields.pop("t")
    t_out = time.parse_time(t_in, "2006-01-02T15:04:05Z")
    metric.time = t_out.unix_nano

    return metric
