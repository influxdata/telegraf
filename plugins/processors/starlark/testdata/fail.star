# Example of the way to return a custom error thanks to the built-in function fail
#
# Example Input:
# fail value=1 1465839830100400201
#
# Example Output Error:
# fail: The field value should be greater than 1

def apply(metric):
    if metric.fields["value"] <= 1:
        return fail("The field value should be greater than 1")
    return metric