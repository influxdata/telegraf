package utility

//fields in the azure table
const COUNTER_NAME = "CounterName"
const END_TIMESTAMP = "Timestamp"
const TOTAL = "Total"
const BEGIN_TIMESTAMP = "TIMESTAMP"
const LAST_SAMPLE = "Last"
const MEAN = "Average"
const MAX_SAMPLE = "Maximum"
const MIN_SAMPLE = "Minimum"
const SAMPLE_COUNT = "Count"
const DEPLOYMENT_ID = "DeploymentId"
const HOST = "Host"

const INPUT_PLUGIN = "input_plugin"

const PERIOD = "period"

const BLOCK_JSON_KEY_COUNTER_NAME = "metricName"
const BLOCK_JSON_KEY_END_TIMESTAMP = "time"
const BLOCK_JSON_KEY_TOTAL = "total"
const BLOCK_JSON_KEY_LAST_SAMPLE = "last"
const BLOCK_JSON_KEY_MAX_SAMPLE = "maximum"
const BLOCK_JSON_KEY_MIN_SAMPLE = "minimum"
const BLOCK_JSON_KEY_SAMPLE_COUNT = "count"
const BLOCK_JSON_KEY_MEAN = "average"
const BLOCK_JSON_KEY_RESOURCE_ID = "resourceId"
const BLOCK_JSON_KEY_TIME_GRAIN = "timeGrain"
const BLOCK_JSON_KEY_DIMENSIONS = "dimensions"
const BLOCK_JSON_KEY_TENANT = "Tenant"
const BLOCK_JSON_KEY_ROLE = "Role"
const BLOCK_JSON_KEY_ROLE_INSTANCE = "RoleInstance"

//parts of azure table name
const WAD_METRICS = "WADMetrics"
const P10DV25 = "P10DV25"
const PT = "PT"
const H = "H"
const M = "M"
const S = "S"
const DATE_SUFFIX_FORMAT = "2006/01/02"

//1 tick is 100ns, 1tick=10^-7 sec, 1sec=10^7tick
const TICKS_PER_SECOND = int64(10000000)

//number of seconds lapsed between 01-01-1601 and 01-01-1970
const EPOCH_DIFFERENCE = int64(11644473600)

const LAYOUT = "02/01/2006 03:04:05 PM"
