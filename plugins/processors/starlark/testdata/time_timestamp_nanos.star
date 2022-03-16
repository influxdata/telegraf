# Example of filtering metrics based on the timestamp with nanoseconds.
#
# Example Input:
# time result="KO" 1617900602123455999
# time result="OK" 1617900602123456789
#
# Example Output:
# time result="OK" 1617900602123456789

load('time.star', 'time')
# loads time.parse_duration(), time.is_valid_timezone(), time.now(), time.time(), 
# time.parse_time() and time.from_timestamp()

def apply(metric):
    # 1617900602123457000 nanosec = Thursday, April 8, 2021 16:50:02.123457000 GMT
    refDate = time.from_timestamp(1617900602, 123457000)
    # 1617900602123455999 nanosec = Thursday, April 8, 2021 16:50:02.123455999 GMT
    # 1617900602123456789 nanosec = Thursday, April 8, 2021 16:50:02.123456789 GMT
    metric_date = time.from_timestamp(int(metric.time / 1e9), int(metric.time % 1e9))
    # Only keep metrics with a timestamp that is not more than 1 microsecond before the reference date 
    if refDate - time.parse_duration("1us") < metric_date:
        return metric
