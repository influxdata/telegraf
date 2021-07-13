# Example of parsing a duration out of a field and modifying the metric to inject the equivalent in seconds.
#
# Example Input:
# time value="3m35s" 1465839830100400201
#
# Example Output:
# time seconds=215 1465839830100400201

load('time.star', 'time')
# loads time.parse_duration(), time.is_valid_timezone(), time.now(), time.time(), 
# time.parse_time() and time.from_timestamp()

def apply(metric):
    duration = time.parse_duration(metric.fields.get('value'))
    metric.fields.pop('value')
    metric.fields["seconds"] = duration.seconds
    return metric
