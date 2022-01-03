package logrus

import (
	"io"
	"log" //nolint:revive // Allow exceptional but valid use of log here.
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var once sync.Once

type LogHook struct {
}

// InstallHook installs a logging hook into the logrus standard logger, diverting all logs
// through the Telegraf logger at debug level.  This is useful for libraries
// that directly log to the logrus system without providing an override method.
func InstallHook() {
	once.Do(func() {
		logrus.SetOutput(io.Discard)
		logrus.AddHook(&LogHook{})
	})
}

func (h *LogHook) Fire(entry *logrus.Entry) error {
	msg := strings.ReplaceAll(entry.Message, "\n", " ")
	log.Print("D! [logrus] ", msg)
	return nil
}

func (h *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
