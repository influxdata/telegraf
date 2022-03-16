# Compute the ratio of two integer fields.
#
# Example: A new field 'usage' from an existing fields 'used' and 'total'
#
# Example Input:
# memory,host=hostname used=11038756864.4948,total=17179869184.1221 1597255082000000000
#
# Example Output:
# memory,host=hostname used=11038756864.4948,total=17179869184.1221,usage=64.25402164701573 1597255082000000000

def apply(metric):
    used = float(metric.fields['used'])
    total = float(metric.fields['total'])
    metric.fields['usage'] = (used / total) * 100
    return metric
