package logger

import "time"

// Config contains the log configuration settings
type Config struct {
	// will set the log level to DEBUG
	Debug bool
	//will set the log level to ERROR
	Quiet bool
	//stderr, stdout, file or eventlog (Windows only)
	LogTarget string
	// will direct the logging output to a file. Empty string is
	// interpreted as stderr. If there is an error opening the file the
	// logger will fall back to stderr
	Logfile string
	// will rotate when current file at the specified time interval
	RotationInterval time.Duration
	// will rotate when current file size exceeds this parameter.
	RotationMaxSize int64
	// maximum rotated files to keep (older ones will be deleted)
	RotationMaxArchives int
	// pick a timezone to use when logging. or type 'local' for local time.
	LogWithTimezone string
}
