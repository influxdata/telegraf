'''
Pivots `sensor` value to be the key for `value`
Input: measurement,tag_set sensor=001A0,value=111.48
Output: measurement,tag_set 001A0=111.48
'''
 
 def apply(metric):
#   metric.fields[str(metric.fields['sensor'])] = metric.fields['value']
#   metric.fields.pop('value',None)
#   metric.fields.pop('sensor',None)
#   return metric