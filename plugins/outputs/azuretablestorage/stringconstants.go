package azuretablestorage

//parts of table name
const WADMetrics = "WADMetrics"
const P10DV25 = "P10DV25"
const PT = "PT"
const H = "H"
const M = "M"
const S = "S"
const DATE_SUFFIX_FORMAT = "2006/01/02"

//Fields in the table
const DeploymentId = "DeploymentId"
const TIMESTAMP = "TIMESTAMP"
const CounterName = "CounterName"
const Host = "Host"
const Period = "Period"

//1 tick is 100ns, 1tick=10^-7 sec, 1sec=10^7tick
const TicksPerSecond = int64(10000000)

//number of seconds lapsed between 01-01-1601 and 01-01-1970
const EpochDifference = int64(11644473600)
