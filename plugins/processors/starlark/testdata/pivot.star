'''
Pivots a key's value to be the key for another key.
In this example it pivots the value of key `sensor`
to be the key of the value in key `value`

Example Input:
temperature sensor="001A0",value=111.48

Example Output:
temperature 001A0=111.48
'''

def apply(metric):
  metric.fields[str(metric.fields['sensor'])] = metric.fields['value']
  metric.fields.pop('value',None)
  metric.fields.pop('sensor',None)
  return metric
