# Example of filtering metrics based on the timestamp in seconds.
#
# Example Input:
# time result="KO" 1616020365100400201
# time result="OK" 1616150517100400201
#
# Example Output:
# time result="OK" 1616150517100400201

load('time.star', 'time')
# loads time.parse_duration(), time.is_valid_timezone(), time.now(), time.time(), 
# time.parse_time() and time.from_timestamp()

def apply(metric):
    # 1616198400 sec = Saturday, March 20, 2021 0:00:00 GMT
    refDate = time.from_timestamp(1616198400)
    # 1616020365 sec = Wednesday, March 17, 2021 22:32:45 GMT
    # 1616150517 sec = Friday, March 19, 2021 10:41:57 GMT
    metric_date = time.from_timestamp(int(metric.time / 1e9))
    # Only keep metrics with a timestamp that is not more than 24 hours before the reference date 
    if refDate - time.parse_duration("24h") < metric_date:
        return metric
