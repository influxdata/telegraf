# Example showing how the math module can be used to compute the value of a field.
#
# Example Input:
# math value=10000i 1465839830100400201
#
# Example Output:
# math result=4 1465839830100400201

load('math.star', 'math')
# loads all the functions and constants defined in the math module

def apply(metric):
    metric.fields["result"] = math.log(metric.fields.pop('value'), 10)
    return metric
