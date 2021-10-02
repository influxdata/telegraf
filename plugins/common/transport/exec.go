package transport

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

// Exec is a transport implementation for locally executing a program
type Exec struct {
	BinPath string          `toml:"bin_path"`
	Timeout config.Duration `toml:"timeout"`
	BinArgs []string        `toml:"-"`

	cmd *exec.Cmd
}

func (e *Exec) SampleConfig() string {
	return `
  ## Optional: path to executable, defaults to $PATH via exec.LookPath
  # bin_path = "/usr/bin/nvidia-smi"

  ## Optional: timeout for execution
  # timeout = "5s"
`
}

// Init performs all preparation work such as argument checks, default value handling etc. It should always
// be called before Receive().
func (e *Exec) Init() error {
	if _, err := os.Stat(e.BinPath); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("binary not at path %s", e.BinPath)
	}

	e.cmd = exec.Command(e.BinPath, e.BinArgs...)

	return nil
}

// Receive calls the given command and returns the raw data received from its output.
func (e *Exec) Receive() ([]byte, error) {
	data, err := internal.CombinedOutputTimeout(e.cmd, time.Duration(e.Timeout))
	if err != nil {
		return nil, fmt.Errorf("executing %q failed: %v", strings.Join(append([]string{e.BinPath}, e.BinArgs...), " "), err)
	}
	return data, nil
}
