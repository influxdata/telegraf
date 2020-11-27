# Complete test case of the shared state.
#
# Example Input:
# cpu i=10i,f=2.35,s="before",b=true 1465839830100400201
# cpu i=20i,f=1.23,s="after",b=false 1465839830100400301
#
# Example Output:
# cpu i=10i,f=2.35,s="before",b=true 1465839830100400201
# cpu i=20i,f=1.23,s="after",b=false 1465839830100400301

def apply(metric):
    first = metric.fields["b"]
    # Test Booleans - START
    if first:
        if Load("b") != None:
            return fail("1. Load b: Should be None")
        # Set to True
        Store("b", True)
    else:
        if Load("b") != True:
            return fail("2. Load b: Should be True")
        # Set to False
        Store("b", False)
        if Load("b") != False:
            return fail("3. Load b: Should be False")
        # Remove b
        Store("b", None)
        if Load("b") != None:
            return fail("4. Load b: Should be None")
    # Test Booleans - END
    # Test Integers - START
    if first:
        if Load("i") != None:
            return fail("1. Load i: Should be None")
        # Set to 10
        Store("i", metric.fields["i"])
    else:
        if Load("i") != 10:
            return fail("2. Load i: Should be 10")
        # Set to 20
        Store("i", metric.fields["i"])
        if Load("i") != 20:
            return fail("3. Load i: Should be 20")
        # Remove i
        Store("i", None)
        if Load("i") != None:
            return fail("4. Load i: Should be None")
    # Test Integers - END
    # Test Floats - START
    if first:
        if Load("f") != None:
            return fail("1. Load f: Should be None")
        # Set to 2.35
        Store("f", metric.fields["f"])
    else:
        if Load("f") != 2.35:
            return fail("2. Load f: Should be 2.35")
        # Set to 1.23
        Store("f", metric.fields["f"])
        if Load("f") != 1.23:
            return fail("3. Load f: Should be 1.23")
        # Remove f
        Store("f", None)
        if Load("f") != None:
            return fail("4. Load f: Should be None")
    # Test Floats - END
    # Test Strings - START
    if first:
        if Load("s") != None:
            return fail("1. Load s: Should be None")
        # Set to before
        Store("s", metric.fields["s"])
    else:
        if Load("s") != "before":
            return fail("2. Load s: Should be before")
        # Set to after
        Store("s", metric.fields["s"])
        if Load("s") != "after":
            return fail("3. Load s: Should be after")
        # Remove s
        Store("s", None)
        if Load("s") != None:
            return fail("4. Load s: Should be None")
    # Test Strings - END
    # Test Metrics - START
    if first:
        if Load("m") != None:
            return fail("1. Load m: Should be None")
        # Set to first metric
        Store("m", metric)
    else:
        if Load("m").fields["i"] != 10:
            return fail("2. Load m: Should be first metric")
        # Set to second metric
        Store("m", metric)
        if Load("m").fields["i"] != 20:
            return fail("3. Load m: Should be second metric")
        # Remove m
        Store("m", None)
        if Load("m") != None:
            return fail("4. Load m: Should be None")
    # Test Metrics - END
    # Test List - START
    if first:
        if Load("l") != None:
            return fail("1. Load l: Should be None")
        # Set to the first list
        Store("l", [1, 2.3, True, "v1", metric, f1])
    else:
        if Load("l")[0] != 1 or Load("l")[1] != 2.3 or Load("l")[2] != True or Load("l")[3] != "v1" or Load("l")[4].fields["i"] != 10 or Load("l")[5]() != 1:
            return fail("2. Load l: Should be first list")
        # Set to second list
        Store("l", [2, 3.3, False, "v2", metric, f2])
        if Load("l")[0] != 2 or Load("l")[1] != 3.3 or Load("l")[2] != False or Load("l")[3] != "v2" or Load("l")[4].fields["i"] != 20 or Load("l")[5]() != 2:
            return fail("3. Load l: Should be second list")
        # Remove l
        Store("l", None)
        if Load("l") != None:
            return fail("4. Load l: Should be None")
    # Test List - END
    # Test Dictionary - START
    if first:
        if Load("d") != None:
            return fail("1. Load d: Should be None")
        # Set to the first dictionary
        Store("d", {"i": 1, "f": 2.3, "b": True, "s": "v1", "m": metric})
    else:
        if Load("d")["i"] != 1 or Load("d")["f"] != 2.3 or Load("d")["b"] != True or Load("d")["s"] != "v1" or Load("d")["m"].fields["i"] != 10:
            return fail("2. Load d: Should be first dictionary")
        # Set to second dictionary
        Store("d", {"i": 2, "f": 3.3, "b": False, "s": "v2", "m": metric})
        if Load("d")["i"] != 2 or Load("d")["f"] != 3.3 or Load("d")["b"] != False or Load("d")["s"] != "v2" or Load("d")["m"].fields["i"] != 20:
            return fail("3. Load d: Should be second dictionary")
        # Remove d
        Store("d", None)
        if Load("d") != None:
            return fail("4. Load d: Should be None")
    # Test Dictionary - END
    # Test Set - START
    if first:
        if Load("se") != None:
            return fail("1. Load se: Should be None")
        # Set to the first set
        Store("se", set([1, 2.3, True, "v1"]))
    else:
        if 1 not in Load("se") or 2.3 not in Load("se") or True not in Load("se") or "v1" not in Load("se"):
            return fail("2. Load se: Should be first set")
        # Set to second set
        Store("se", set([2, 3.3, False, "v2"]))
        if 2 not in Load("se") or 3.3 not in Load("se") or False not in Load("se") or "v2" not in Load("se"):
            return fail("3. Load se: Should be second set")
        # Remove se
        Store("se", None)
        if Load("se") != None:
            return fail("4. Load se: Should be None")
    # Test Set - END
    # Test Tuple - START
    if first:
        if Load("t") != None:
            return fail("1. Load t: Should be None")
        # Set to the first tuple
        Store("t", (1, 2.3, True, "v1"))
    else:
        if 1 not in Load("t") or 2.3 not in Load("t") or True not in Load("t") or "v1" not in Load("t"):
            return fail("2. Load t: Should be first tuple")
        # Set to second tuple
        Store("t", (2, 3.3, False, "v2"))
        if 2 not in Load("t") or 3.3 not in Load("t") or False not in Load("t") or "v2" not in Load("t"):
            return fail("3. Load t: Should be second tuple")
        # Remove t
        Store("t", None)
        if Load("t") != None:
            return fail("4. Load t: Should be None")
    # Test Tuple - END
    # Test Function - START
    if first:
        if Load("fu") != None:
            return fail("1. Load fu: Should be None")
        # Set to the first function
        Store("fu", f1)
    else:
        if Load("fu")() != 1:
            return fail("2. Load fu: Should be first function")
        # Set to second function
        Store("fu", f2)
        if Load("fu")() != 2:
            return fail("3. Load fu: Should be second function")
        # Remove fu
        Store("fu", None)
        if Load("fu") != None:
            return fail("4. Load fu: Should be None")
    # Test Function - END    
    return metric

def f1():
    return 1

def f2():
    return 2