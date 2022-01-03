# Example of setting the metric timestamp to the current time.
#
# Example Input:
# time result="OK" 1515581000000000000
#
# Example Output:
# time result="OK" 1618488000000000999

load('time.star', 'time')

def apply(metric):    
    # You can set the timestamp by using the current time.
    metric.time = time.now().unix_nano

    return metric