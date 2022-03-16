# Example of parsing a date out of a field and modifying the metric to inject the year, month and day.
#
# Example Input:
# time value="2009-06-12T12:06:10.000000099" 1465839830100400201
#
# Example Output:
# time year=2009i,month=6i,day=12i 1465839830100400201

load('time.star', 'time')
# loads time.parse_duration(), time.is_valid_timezone(), time.now(), time.time(), 
# time.parse_time() and time.from_timestamp()

def apply(metric):
    date = time.parse_time(metric.fields.get('value'), format="2006-01-02T15:04:05.999999999", location="UTC")
    metric.fields.pop('value')
    metric.fields["year"] = date.year
    metric.fields["month"] = date.month
    metric.fields["day"] = date.day
    return metric
