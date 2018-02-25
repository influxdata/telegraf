package oracle

import (
	"database/sql"
	"regexp"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"gopkg.in/goracle.v2"
)

type Oracle struct {
	Connection
	InstanceStateMetrics bool
	SystemMetrics        bool
	TablespaceMetrics    bool
	WaitClassMetrics     bool
	WaitEventMetrics     bool
}

type Connection struct {
	DB            *sql.DB
	MaxLifetime   internal.Duration
	MinSessions   int
	MaxSessions   int
	Password      string
	PoolIncrement int
	SID           string
	Username      string
}

var sampleConfig = `
## Username used to connect to Oracle.
username = "telegraf"
## Password used to connect to Oracle.
password = "telegraf"
## SID used to connect to Oracle.
sid = "localhost/sid"
## Minimum number of database connections that the connection pool can contain. Defaults to 10.
min_sessions = 10
## Maximum number of database connections that the connection pool can contain. Defaults to 20.
max_sessions = 20
## Increment by which the connection pool capacity is expanded. Defaults to 1.
pool_increment = 1
## Maximum amount of time a connection may be reused. Defaults to 0 or forever.
max_lifetime = "0s"

## Collect instance state metrics from V$INSTANCE. Defaults to true.
instance_state_metrics = true
## Collect system metrics from V$SYSMETRIC. Defaults to true.
system_metrics = true
## Collect tablespace metrics from DBA_TABLESPACE_USAGE_METRICS. Defaults to true.
tablespace_metrics = true
## Collect wait event metrics from V$EVENTMETRIC. Defaults to true.
wait_event_metrics = true
## Collect wait class metrics from V$WAITCLASSMETRIC. Defaults to true.
wait_class_metrics = true
`

func (o Oracle) SampleConfig() string {
	return sampleConfig
}

func (o Oracle) Description() string {
	return "Read metrics about an Oracle database."
}

func (o Oracle) Gather(acc telegraf.Accumulator) error {
	var (
		err error
		wg  sync.WaitGroup
	)

	if o.DB == nil {
		o.DB, err = sql.Open("goracle", goracle.ConnectionParams{
			Username:      o.Username,
			Password:      o.Password,
			SID:           o.SID,
			MinSessions:   o.MinSessions,
			MaxSessions:   o.MaxSessions,
			PoolIncrement: o.PoolIncrement,
		}.String())
		if err != nil {
			return err
		}
		o.DB.SetConnMaxLifetime(o.MaxLifetime.Duration)
	}

	wg.Add(5)

	go func() {
		defer wg.Done()
		if o.InstanceStateMetrics {
			acc.AddError(gatherInstanceStateMetrics(o.DB, acc))
		}
	}()

	go func() {
		defer wg.Done()
		if o.SystemMetrics {
			acc.AddError(gatherSystemMetrics(o.DB, acc))
		}
	}()

	go func() {
		defer wg.Done()
		if o.TablespaceMetrics {
			acc.AddError(gatherTablespaceMetrics(o.DB, acc))
		}
	}()

	go func() {
		defer wg.Done()
		if o.WaitClassMetrics {
			acc.AddError(gatherWaitClassMetrics(o.DB, acc))
		}
	}()

	go func() {
		defer wg.Done()
		if o.WaitEventMetrics {
			acc.AddError(gatherWaitEventMetrics(o.DB, acc))
		}
	}()

	wg.Wait()

	return nil
}

func gatherCommonTags(db *sql.DB) (map[string]string, error) {
	var (
		instance_name string
		host_name     string
		name          string
		version       string
		instance_role string
		queryInstance = "SELECT INSTANCE_NAME,HOST_NAME,VERSION,INSTANCE_ROLE FROM V$INSTANCE"
		queryDatabase = "SELECT NAME FROM V$DATABASE"
	)

	if err := db.QueryRow(queryInstance).Scan(&instance_name, &host_name, &version, &instance_role); err != nil {
		return map[string]string{}, err
	}

	if err := db.QueryRow(queryDatabase).Scan(&name); err != nil {
		return map[string]string{}, err
	}

	return map[string]string{
		"database_name": name,
		"instance_name": instance_name,
		"db_host":       host_name,
		"version":       version,
		"instance_role": instance_role,
	}, nil
}

func gatherInstanceStateMetrics(db *sql.DB, acc telegraf.Accumulator) error {
	var (
		active_state     string
		archiver         string
		database_status  string
		logins           string
		shutdown_pending string
		status           string
		query            = "SELECT ACTIVE_STATE,ARCHIVER,DATABASE_STATUS,LOGINS,SHUTDOWN_PENDING,STATUS FROM V$INSTANCE"
	)

	tags, err := gatherCommonTags(db)
	if err != nil {
		return err
	}

	if err := db.QueryRow(query).Scan(&active_state, &archiver, &database_status, &logins, &shutdown_pending, &status); err != nil {
		return err
	}

	fields := map[string]interface{}{
		"active_state_normal":               0,
		"active_state_quiescing":            0,
		"active_state_quiesced":             0,
		"archiver_started":                  0,
		"archiver_stopped":                  0,
		"archiver_failed":                   0,
		"database_status_active":            0,
		"database_status_suspended":         0,
		"database_status_instance_recovery": 0,
		"logins_allowed":                    0,
		"logins_restricted":                 0,
		"shutdown_pending":                  0,
		"status_started":                    0,
		"status_mounted":                    0,
		"status_open":                       0,
		"status_open_migrate":               0,
	}

	switch active_state {
	case "NORMAL":
		fields["active_state_normal"] = 1
	case "QUIESCING":
		fields["active_state_quiescing"] = 1
	case "QUIESCED":
		fields["active_state_quiesced"] = 1
	}

	switch archiver {
	case "STARTED":
		fields["archiver_started"] = 1
	case "STOPPED":
		fields["archiver_stopped"] = 1
	case "FAILED":
		fields["archiver_failed"] = 1
	}

	switch database_status {
	case "ACTIVE":
		fields["database_status_active"] = 1
	case "SUSPENDED":
		fields["database_status_suspended"] = 1
	case "INSTANCE_RECOVERY":
		fields["database_status_instance_recovery"] = 1
	}

	switch logins {
	case "ALLOWED":
		fields["logins_allowed"] = 1
	case "RESTRICTED":
		fields["logins_restricted"] = 1
	}

	switch shutdown_pending {
	case "YES":
		fields["shutdown_pending"] = 1
	}

	switch status {
	case "STARTED":
		fields["status_started"] = 1
	case "MOUNTED":
		fields["status_mounted"] = 1
	case "OPEN":
		fields["status_open"] = 1
	case "OPEN MIGRATE":
		fields["status_open_migrate"] = 1
	}

	acc.AddFields("oracle_instance_state", fields, tags)

	return nil
}

func gatherSystemMetrics(db *sql.DB, acc telegraf.Accumulator) error {
	var (
		metric_name string
		value       float64
	)

	fields := make(map[string]interface{})

	tags, err := gatherCommonTags(db)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT METRIC_NAME,VALUE FROM V$SYSMETRIC")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&metric_name, &value); err != nil {
			acc.AddError(err)
		} else {
			fields[sanitize(metric_name)] = value
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	acc.AddFields("oracle_system", fields, tags)

	return nil
}

func gatherTablespaceMetrics(db *sql.DB, acc telegraf.Accumulator) error {
	var (
		tablespace_name string
		used_space      float64
		tablespace_size float64
		used_percent    float64
		query           = "SELECT TABLESPACE_NAME,USED_SPACE,TABLESPACE_SIZE,USED_PERCENT FROM DBA_TABLESPACE_USAGE_METRICS"
	)

	tags, err := gatherCommonTags(db)
	if err != nil {
		return err
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&tablespace_name, &used_space, &tablespace_size, &used_percent); err != nil {
			acc.AddError(err)
		} else {
			tags["tablespace"] = tablespace_name
			acc.AddFields("oracle_tablespace", map[string]interface{}{
				"used_space":      used_space,
				"tablespace_size": tablespace_size,
				"used_percent":    used_percent,
			}, tags)
		}
	}
	if rows.Err() != nil {
		return err
	}

	return nil
}

func gatherWaitClassMetrics(db *sql.DB, acc telegraf.Accumulator) error {
	var (
		wait_class           string
		average_waiter_count float64
		dbtime_in_wait       float64
		time_waited          float64
		wait_count           float64
		time_waited_fg       float64
		wait_count_fg        float64
		query                = "SELECT b.WAIT_CLASS,a.AVERAGE_WAITER_COUNT,a.DBTIME_IN_WAIT,a.TIME_WAITED,a.WAIT_COUNT,a.TIME_WAITED_FG,a.WAIT_COUNT_FG FROM V$WAITCLASSMETRIC a, V$SYSTEM_WAIT_CLASS b where a.WAIT_CLASS_ID=b.WAIT_CLASS_ID"
	)

	tags, err := gatherCommonTags(db)
	if err != nil {
		return err
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&wait_class, &average_waiter_count, &dbtime_in_wait, &time_waited, &wait_count, &time_waited_fg, &wait_count_fg); err != nil {
			acc.AddError(err)
		} else {
			tags["class"] = sanitize(wait_class)
			acc.AddFields("oracle_wait_class", map[string]interface{}{
				"average_waiter_count": average_waiter_count,
				"dbtime_in_wait":       dbtime_in_wait,
				"time_waited":          time_waited,
				"wait_count":           wait_count,
				"time_waited_fg":       time_waited_fg,
				"wait_count_fg":        wait_count_fg,
			}, tags)
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func gatherWaitEventMetrics(db *sql.DB, acc telegraf.Accumulator) error {
	var (
		name             string
		wait_class       string
		num_sess_waiting int
		time_waited      float64
		wait_count       int
		time_waited_fg   float64
		wait_count_fg    float64
		query            = "SELECT b.NAME,b.WAIT_CLASS,a.NUM_SESS_WAITING,a.TIME_WAITED,a.WAIT_COUNT,a.TIME_WAITED_FG,a.WAIT_COUNT_FG FROM V$EVENTMETRIC a, V$EVENT_NAME b where a.EVENT_ID=b.EVENT_ID"
	)

	tags, err := gatherCommonTags(db)
	if err != nil {
		return err
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&name, &wait_class, &num_sess_waiting, &time_waited, &wait_count, &time_waited_fg, &wait_count_fg); err != nil {
			acc.AddError(err)
		} else {
			tags["event"] = sanitize(name)
			tags["class"] = sanitize(wait_class)
			acc.AddFields("oracle_wait_event", map[string]interface{}{
				"num_sess_waiting": num_sess_waiting,
				"time_waited":      time_waited,
				"wait_count":       wait_count,
				"time_waited_fg":   time_waited_fg,
				"wait_count_fg":    wait_count_fg,
			}, tags)
		}
	}
	if rows.Err() != nil {
		return err
	}

	return nil
}

func sanitize(s string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_]+")

	s = strings.NewReplacer("I/O", "io", "%", "percent").Replace(s)
	s = r.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	return strings.ToLower(s)
}

func init() {
	inputs.Add("oracle", func() telegraf.Input {
		return &Oracle{
			Connection: Connection{
				MaxLifetime:   internal.Duration{},
				MinSessions:   10,
				MaxSessions:   20,
				PoolIncrement: 1,
			},
			InstanceStateMetrics: true,
			SystemMetrics:        true,
			TablespaceMetrics:    true,
			WaitClassMetrics:     true,
			WaitEventMetrics:     true,
		}
	})
}
