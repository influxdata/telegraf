package azuretablestorage

//parts of table name
const WAD_METRICS = "WADMetrics"
const P10DV25 = "P10DV25"
const PT = "PT"
const H = "H"
const M = "M"
const S = "S"
const DATE_SUFFIX_FORMAT = "2006/01/02"

//Fields in the table
const DEPLOYMENT_ID = "DeploymentId"
const TIMESTAMP = "TIMESTAMP"
const COUNTER_NAME = "CounterName"
const HOST = "Host"
const PERIOD = "Period"

//1 tick is 100ns, 1tick=10^-7 sec, 1sec=10^7tick
const TICKS_PER_SECOND = int64(10000000)

//number of seconds lapsed between 01-01-1601 and 01-01-1970
const EPOCH_DIFFERENCE = int64(11644473600)
