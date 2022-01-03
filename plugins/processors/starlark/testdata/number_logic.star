# Set a logic function to transform a numerical value to another numerical value
# Example: Set  any 'status' field between 1 and 6 to a value of 0
#
# Example Input:
# lb,http_method=GET status=5i 1465839830100400201
#
# Example Output:
# lb,http_method=GET status=0i 1465839830100400201


def apply(metric):
    v = metric.fields.get('status')
    if v == None:
        return metric
    if 1 < v and v < 6:
        metric.fields['status'] = 0
    return metric
