# Example of the way to log a message with all the supported levels
# using the logger of Telegraf.
#
# Example Input:
# log debug="a debug message" 1465839830100400201
#
# Example Output:
# log debug="a debug message" 1465839830100400201

load("logging.star", "log")
# loads log.debug(), log.info(), log.warn(), log.error()

def apply(metric):
    log.debug("debug: {}".format(metric.fields["debug"]))
    log.info("an info message")
    log.warn("a warning message")
    log.error("an error message")
    return metric
 