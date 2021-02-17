# Produces a new Line of statistics about the Fields
#
# Example Input:
# logstash,environment_id=EN456,property_id=PR789,request_type=ingress,stack_id=engd asn=1313i,cache_response_code=202i,colo_code="LAX",colo_id=12i,compute_time=28736i,edge_end_timestamp=1611085500320i,edge_start_timestamp=1611085496208i,id="1b5c67ed-dfd0-4d30-99bd-84f0a9c5297b_76af1809-29d1-4b35-a0cf-39797458275c",parent_ray_id="00",processing_details="ok",rate_limit_id=0i,ray_id="76af1809-29d1-4b35-a0cf-39797458275c",request_bytes=7777i,request_host="engd-08364a825824e04f0a494115.reactorstream.dev",request_id="1b5c67ed-dfd0-4d30-99bd-84f0a9c5297b",request_result="succeeded",request_uri="/ENafcb2798a9be4bb7bfddbf35c374db15",response_code=200i,subrequest=false,subrequest_count=1i,user_agent="curl/7.64.1" 1611085496208
#
# Example Output:
# sizing,measurement=logstash,environment_id=EN456,property_id=PR789,request_type=ingress,stack_id=engd tag_count=5,tag_key_avg_length=10.2,tag_value_avg_length=10.6,int_avg_length=4.9,int_count=10,bool_avg_length=5,bool_count=1,str_avg_length=25.4,str_count=10 1611085496208

def apply(metric):
    new_metric = Metric("sizing")
    num_tags = len(metric.tags.items())
    new_metric.fields["tag_count"] = float(num_tags)
    new_metric.fields["tag_key_avg_length"] = _sum(map(len, metric.tags.keys())) / num_tags
    new_metric.fields["tag_value_avg_length"] = _sum(map(len, metric.tags.values())) / num_tags

    new_metric.tags["measurement"] =  metric.name

    new_metric.tags.update(metric.tags)

    ints, floats, bools, strs = [], [], [], []
    for field in metric.fields.items():
        value = field[1]
        if type(value) == "int":
            ints.append(field)
        elif type(value) == "float":
            floats.append(field)
        elif type(value) == "bool":
            bools.append(field)
        elif type(value) == "string":
            strs.append(field)

    if len(ints) > 0:
        int_vals = [i[1] for i in ints]
        produce_pairs(new_metric, int_vals, "int")
    if len(floats) > 0:
        float_vals = [i[1] for i in floats]
        produce_pairs(new_metric, float_vals, "float")
    if len(bools) > 0:
        bool_vals = [i[1] for i in bools]
        produce_pairs(new_metric, bool_vals, "bool")     
    if len(strs) > 0:
        str_vals = [i[1] for i in strs]
        produce_pairs(new_metric, str_vals, "str")

    return new_metric

def produce_pairs(metric, li, field_type):
    lens = elem_lengths(li)
    counts = count_lengths(lens)
    max_value = max(counts)
    metric.fields["{}_avg_length".format(field_type)]  = float(mean(lens))
    metric.fields["{}_count".format(field_type)]       = float(len(li))

def elem_lengths(li):
    if type(li[0]) in ("int", "float", "bool"):
        return [len(str(elem)) for elem in li]
    else:
        return [len(elem) for elem in li]

def get_key(di, value):
    keys = []
    for k,v in di.items():
        if v == value:
            keys.append(k)
    return keys

def count_lengths(li):
    # Returns dict of counts of each occurrence of length in a list of lengths
    lens = []
    counts = []
    for elem in li:
        if elem not in lens:
            lens.append(elem)
            counts.append(1)
        else:
            index = lens.index(elem)
            counts[index] += 1
    return dict(zip(lens, counts))

def map(f, li):
    return [f(x) for x in li]

def _sum(li):
    sum = 0
    for i in li:
        sum += i
    return sum

def mean(li):
    return _sum(li)/len(li)
