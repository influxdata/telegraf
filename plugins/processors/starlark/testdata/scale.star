# Multiply any float fields by 10
#
# Example Input:
# modbus,host=hostname Current=1.22,Energy=0,Frequency=60i,Power=0,Voltage=123.9000015258789 1554079521000000000
#
# Example Output:
# modbus,host=hostname Current=12.2,Energy=0,Frequency=60i,Power=0,Voltage=1239.000015258789 1554079521000000000

def apply(metric):
    for k, v in metric.fields.items():
        if type(v) == "float":
            metric.fields[k] = v * 10
    return metric
