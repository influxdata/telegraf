package gorethink

import (
	"reflect"

	"github.com/Sirupsen/logrus"

	"github.com/dancannon/gorethink/encoding"
)

var (
	log *logrus.Logger
)

func init() {
	// Set encoding package
	encoding.IgnoreType(reflect.TypeOf(Term{}))

	log = logrus.New()
}

// SetVerbose allows the driver logging level to be set. If true is passed then
// the log level is set to Debug otherwise it defaults to Info.
func SetVerbose(verbose bool) {
	if verbose {
		log.Level = logrus.DebugLevel
		return
	}

	log.Level = logrus.InfoLevel
}
