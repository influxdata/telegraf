
# This Starlark processor is used when loading Sparkplug B protobuf #
# messages into InfluxDB.  The data source is a Opto22 Groov EPIC controller.  
#
# This processor does the following:
#   - Resolves the metric name using a numeric alias.
#     When the EPIC MQTT client is started it sends a DBIRTH message 
#     that lists all metrics configured on the controller and includes 
#     a sequential numeric alias to reference it by.  
#     This processor stores that information in the array states["aliases"].

#     When subsequent DDATA messages are published, the numeric alias is 
#     used to find the stored metric name in the array states["aliases"].
#  - Splits the MQTT topic into 5 fields which can be used as tags in InfluxDB.
#  - Splits the metric name into 6 fields which are be used as tags in InfluxDB.
#  - Deletes the host, type, topic, name and alias tags
#
# TODO:
#   The requirment that a DBIRTH message has to be received before DDATA messages
#   can be used creates a significant reliability issue and a debugging mess.
#   I have to go into the Groov EPIC controller and restart the MQTT client everytime
#   I restart the telegraf loader.  This has caused many hours of needless frustration.
#   
#   I see two possible solutions:
#      - Opto 22 changes their software making it optional to drop the alias 
#        and simply include the name in the DDATA messages.  In my case it's never more
#        than 15 characters.  This is the simplest and most reliable solution.
#      - Make a system call from telegraf and using SSH to remotely restart the MQTT client.
#      - Have telegraf send a message through MQTT requesting a DBIRTH message from the EPIC Controller.
#
# Example Input:
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 type=9i,value=22.247711,alias=10i 1626475876000000000
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 alias=10i,type=9i,value=22.231323 1626475877000000000
# edge,host=firefly,topic=spBv1.0/SF/DBIRTH/epiclc/Exp501 type=9i,name="Strategy/IO/I_Ch_TC_Right",alias=9i 1626475880000000000
# edge,host=firefly,topic=spBv1.0/SF/DBIRTH/epiclc/Exp501 value=22.200958,name="Strategy/IO/I_Ch_TC_Top_C",type=9i,alias=10i 1626475881000000000
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 alias=10i,type=9i,value=22.177643 1626475884000000000
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 type=9i,value=22.231903,alias=10i 1626475885000000000
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 value=22.165192,alias=10i,type=9i 1626475895000000000
# edge,host=firefly,topic=spBv1.0/SF/DDATA/epiclc/Exp501 alias=10i,type=9i,value=22.127106 1626475896000000000
#
# Example Output:
# C,Component=Ch,Datatype=IO,Device=TC,EdgeID=epiclc,Experiment=Exp501,Metric=I_Ch_TC_Top_C,MsgType=DBIRTH,Position=Top,Reactor=SF,Source=Strategy value=22.200958 1626475881000000000
# C,Component=Ch,Datatype=IO,Device=TC,EdgeID=epiclc,Experiment=Exp501,Metric=I_Ch_TC_Top_C,MsgType=DDATA,Position=Top,Reactor=SF,Source=Strategy value=22.177643 1626475884000000000
# C,Component=Ch,Datatype=IO,Device=TC,EdgeID=epiclc,Experiment=Exp501,Metric=I_Ch_TC_Top_C,MsgType=DDATA,Position=Top,Reactor=SF,Source=Strategy value=22.231903 1626475885000000000
# C,Component=Ch,Datatype=IO,Device=TC,EdgeID=epiclc,Experiment=Exp501,Metric=I_Ch_TC_Top_C,MsgType=DDATA,Position=Top,Reactor=SF,Source=Strategy value=22.165192 1626475895000000000
# C,Component=Ch,Datatype=IO,Device=TC,EdgeID=epiclc,Experiment=Exp501,Metric=I_Ch_TC_Top_C,MsgType=DDATA,Position=Top,Reactor=SF,Source=Strategy value=22.127106 1626475896000000000

#############################################
# The following is the telegraf.conf used when calling this processor

# [[inputs.mqtt_consumer]]
#   servers = ["tcp://your_server:1883"]
#   qos = 0
#   connection_timeout = "30s"
#   topics = ["spBv1.0/#"]
#   persistent_session = false
#   client_id = ""
#   username = "your username"
#   password = "your password"
# 
#   # Sparkplug protobuf configuration
#   data_format = "xpath_protobuf"
#    
#   # URL of sparkplug protobuf prototype
#   xpath_protobuf_type = "org.eclipse.tahu.protobuf.Payload"
#    
#   # Location of sparkplug_b.proto file
#   xpath_protobuf_file = "/apps/telegraf/config/sparkplug_b.proto"
# 
#   [[inputs.mqtt_consumer.xpath_protobuf]]
#     metric_selection = "metrics[not(template_value)]"
#     metric_name = "concat('edge', substring-after(name, ' '))"
#     timestamp = "timestamp"
#     timestamp_format = "unix_ms"
#     [inputs.mqtt_consumer.xpath_protobuf.tags]
#       name = "substring-after(name, ' ')"
#     [inputs.mqtt_consumer.xpath_protobuf.fields_int]
#       type = "datatype"
#       alias = "alias"
#     [inputs.mqtt_consumer.xpath_protobuf.fields]
#       # A metric value must be numeric
#       value = "number((int_value | long_value | float_value | double_value | boolean_value))"
#       name = "name"
# 
# # Starlark processor
# [[processors.starlark]]
#   script = "sparkplug.star"
# 
#   # Optionally Define constants used in sparkplug.star
#   # Constants can be defined here or they can be defined in the 
#   # sparkplug_b.star file.
#
#   [processors.starlark.constants]     
#
#     # NOTE: The remaining fields can be specified either here or in the starlark script.
#     
#     # Tags used to identify message type - 3rd field of topic
#     BIRTH_TAG = "BIRTH/"
#     DEATH_TAG = "DEATH/"
#     DATA_TAG = "DATA/"
# 
#     # Number of messages to hold if alias cannot be resolved 
#     MAX_UNRESOLVED = 3
# 
#     # Provide alternate names for the 5 sparkplug topic fields.  
#     # The topic contains 5 fields separated by the '/' character.  
#     # Define the tag name for each of these fields.
#     MSG_FORMAT = "false"        #0
#     GROUP_ID   = "reactor"      #1
#     MSG_TYPE   = "false"        #2
#     EDGE_ID    = "edgeid"       #3
#     DEVICE_ID  = "experiment"   #4
#

BIRTH_TAG = "BIRTH/"
DEATH_TAG = "DEATH/"
DATA_TAG  = "DATA/"
  
# Number of messages to hold if alias cannot be resolved 
MAX_UNRESOLVED = 3

# Provide alternate names for the 5 sparkplug topic fields.  
# The topic contains 5 fields separated by the '/' character.  
# Define the tag name for each of these fields.
MSG_FORMAT = "false"        #0
GROUP_ID   = "Reactor"      #1
MSG_TYPE   = "MsgType"      #2
EDGE_ID    = "EdgeID"       #3
DEVICE_ID  = "Experiment"   #4
 
########### Begin sparkplug.star script


load("logging.star", "log")

state = {
    "aliases":    dict(),
    "devices":    dict(),
    "unresolved": list()
}

def extractTopicTags(metric):
    msg_format   = ''
    groupid      = ''
    msg_type     = ''
    edgeid       = ''
    deviceid     = ''

    topic = metric.tags.get("topic", "");
    fields = topic.split("/");
    nfields = len(fields)
    if nfields > 0: msg_format = fields[0]
    if nfields > 1: groupid    = fields[1]
    if nfields > 2: msg_type   = fields[2]
    if nfields > 3: edgeid     = fields[3]
    if nfields > 4: deviceid   = fields[4]
    return [msg_format, groupid, msg_type, edgeid, deviceid]
   
                
def buildTopicTags(metric, topicFields):
    # Remove topic and host tags - they are not useful for analysis
    metric.tags.pop("topic")
    metric.tags.pop("host")

    if MSG_FORMAT != "false": metric.tags[MSG_FORMAT] = topicFields[0] 
    if GROUP_ID   != "false": metric.tags[GROUP_ID]   = topicFields[1] 
    if MSG_TYPE   != "false": metric.tags[MSG_TYPE]   = topicFields[2] 
    if EDGE_ID    != "false": metric.tags[EDGE_ID]    = topicFields[3]
    if DEVICE_ID  != "false": metric.tags[DEVICE_ID]  = topicFields[4]


def buildNameTags(metric,name):
    # Remove type and alias from metric.fields - They are not useful for analysis
    metric.fields.pop("type")
    metric.fields.pop("alias")
    if "name" in metric.fields:
        metric.fields.pop("name")

    # The Groov EPIC metric names are comprised of 3 fields separated by a '/'
    #   source, datatype, and metric name
    # Extract these fields and include them as tags.
    fields = name.split('/')
    nfields = len(fields)
    if nfields > 0: 
        metric.tags["Source"] = fields[0]
    if nfields > 1:
        metric.tags["Datatype"] = fields[1]
    if nfields > 2: 
        metric.tags["Metric"] = fields[2]

        # OPTIONAL
        #
        # By using underscore characters the metric name can be further
        # divided into additional tags.  
        # How this is defined is site specific.  
        # Customize this as you wish

        # The following demonstrates dividing the metric name into 3, 4 or 5 new tags
        # A metric name must have between 3-5 underscore separated fields 
        
        # If there is only one or two fields then the only tag created is 'metric' 
        # which has the full name
        #
        # The last field is Units and is filled before fields 3, 4 and 5
        # Ex: C, V, Torr, W, psi, RPM, On....
        # The units are used in Influx as the 'measurement' name.
        #
        # 
        # Fields 3, 4 and 5 (device, position, composition) are optional
        #    measurement_component_device_position_composition_units
        #
        # Ex:  I_FuelTank1_C                    (2 fields) 
        #         Measurement   I
        #         Component     FuelTank1   
        #         Units         C
        #
        #      I_FuelTank1_TC_Outlet_C          (5 fields)           
        #         Measurement   I
        #         Component     FuelTank1   
        #         Device        TC
        #         Position      Outlet
        #         Units         C
        #
        #      I_FuelTank1_TC_Outlet_Premium_C  (6 fields) 
        #         Measurement   I
        #         Component     FuelTank1   
        #         Device        TC  
        #         Position      Outlet
        #         Composition   Premium   
        #         Units         C

        # Split the metric name into fields using '_' 
        sfields = fields[2].split('_')
        nf = len(sfields)
        # Don't split the name if it's one or two fields 
        if nf <= 2:
            metric.name = "Name"
        if nf > 2:
            metric.name = sfields[nf-1]     # The Units are used for the metric name
            metric.tags["Component"] = sfields[1]
        if nf > 3:
            metric.tags["Device"] = sfields[2]
        if nf > 4:
            metric.tags["Position"] = sfields[3]
        if nf > 5:
            metric.tags["Composition"] = sfields[4]

def apply(metric):
    output = metric

    log.debug("apply metric: {}".format(metric))

    topic = metric.tags.get("topic", "")
    topicFields = extractTopicTags(metric)
    edgeid = topicFields[3]      # Sparkplug spec specifies 4th field as edgeid

    # Split the topic into fields and assign to variables
    # Determine if the message is of type birth and if so add it to the "devices" LUT.
    if DEATH_TAG in topic:
        output = None
    elif BIRTH_TAG in topic:
        log.debug("    metric msg_type: {}    edgeid: {}   topic: {}".format(BIRTH_TAG, edgeid, topic))
        if "alias" in metric.fields and "name" in metric.fields:
            # Create the lookup-table using "${edgeid}/${alias}" as the key and "${name}" as value
            alias = metric.fields.get("alias")
            name = metric.fields.get("name")
            id = "{}/{}".format(edgeid,alias)
            log.debug("  --> setting alias: {}    name: {}   id: {}'".format(alias, name, id))
            state["aliases"][id] = name
            if "value" in metric.fields:
                buildTopicTags(metric, topicFields)
                buildNameTags(metric, name)
            else:
                output = None

            # Try to resolve the unresolved if any
            if len(state["unresolved"]) > 0:
                # Filter out the matching metrics and keep the rest as unresolved
                log.debug("    unresolved")
                unresolved = [("{}/{}".format(edgeid, m.fields["alias"]), m) for m in state["unresolved"]]
                matching = [(mid, m) for mid, m in unresolved if mid == id]
                state["unresolved"] = [m for mid, m in unresolved if mid != id]

                log.debug("    found {} matching unresolved metrics".format(len(matching)))
                # Process the matching metrics and output - TODO - needs debugging
                # for mid, m in matching:
                #     buildTopicTags(m,topicFields)
                #     buildNameTags(m)
                # output = [m for _, m in matching] + [metric]

    elif DATA_TAG in topic:
        log.debug("    metric msg_type: {}    edgeid: {}   topic: {}".format(DATA_TAG, edgeid, topic))
        if "alias" in metric.fields:
            alias = metric.fields.get("alias")

            # Lookup the ID. If we know it, replace the name of the metric with the lookup value,
            # otherwise we need to keep the metric for resolving later. 
            # This can happen if the messages are out-of-order for some reason...
            id = "{}/{}".format(edgeid,alias)
            if id in state["aliases"]:
                name = state["aliases"][id]
                log.debug("    found alias: {}     name: {}".format(alias, name))
                buildTopicTags(metric,topicFields)
                buildNameTags(metric,name)
            else:
                # We want to hold the metric until we get the corresponding birth message
                log.debug("    id not found: {}".format(id))
                output = None
                if len(state["unresolved"]) >= MAX_UNRESOLVED:
                    log.warn("    metric overflow, trimming {}".format(len(state["unresolved"]) - MAX_UNRESOLVED+1))
                    # Release the unresolved metrics as raw and trim buffer
                    output = state["unresolved"][MAX_UNRESOLVED-1:]
                    state["unresolved"] = state["unresolved"][:MAX_UNRESOLVED-1]
                log.debug("    --> keeping metric")
                state["unresolved"].append(metric)
        else:
            output = None

    return output

