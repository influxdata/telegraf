# Drop fields if they NOT contain values of an expected type.
#
# In this example we ignore fields with an unknown expected type and do not drop them.
#
# Example Input:
# measurement,host=hostname a=1i,b=4.2,c=42.0,d="v3.14",e=true,f=23.0 1597255410000000000
# measurement,host=hostname a=1i,b="somestring",c=42.0,d="v3.14",e=true,f=23.0 1597255410000000000
#
# Example Output:
# measurement,host=hostname a=1i,b=4.2,c=42.0,d="v3.14",e=true,f=23.0 1597255410000000000
# measurement,host=hostname a=1i,c=42.0,d="v3.14",e=true,f=23.0 1597255410000000000

load("logging.star", "log")
# loads log.debug(), log.info(), log.warn(), log.error()

expected_type = {
    "a": "int",
    "b": "float",
    "c": "float",
    "d": "string",
    "e": "bool"
}

def apply(metric):
    for k, v in metric.fields.items():
        if type(v) != expected_type.get(k, type(v)):
            metric.fields.pop(k)
            log.warn("Unexpected field type dropped: metric {} had field {} with type {}, but it is expected to be {}".format(metric.name, k, type(v), expected_type.get(k, type(v))))

    return metric
