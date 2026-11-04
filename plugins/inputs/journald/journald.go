//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package journald

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Message printed by journalctl if the given cursor cannot be used
const seekCursorFailed = "Failed to seek to cursor"

type Journald struct {
	Binary                string          `toml:"binary"`
	UseSudo               bool            `toml:"use_sudo"`
	Units                 []string        `toml:"units"`
	UserUnits             []string        `toml:"user_units"`
	Priority              string          `toml:"priority"`
	Matches               []string        `toml:"matches"`
	Directory             string          `toml:"directory"`
	Namespace             string          `toml:"namespace"`
	InitialReadOffset     string          `toml:"initial_read_offset"`
	TagFields             []string        `toml:"tag_fields"`
	JournalFields         []string        `toml:"journal_fields"`
	MaxUndeliveredEntries int             `toml:"max_undelivered_entries"`
	RestartDelay          config.Duration `toml:"restart_delay"`
	Log                   telegraf.Logger `toml:"-"`

	command   string
	baseArgs  []string
	tagKeys   map[string]string
	fieldKeys filter.Filter

	acc     telegraf.TrackingAccumulator
	sem     chan struct{}
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	tracker *cursorTracker

	// Cursor of the last entry handed to the accumulator, used to continue
	// seamlessly if journalctl needs to be restarted.
	readCursor string
}

func (*Journald) SampleConfig() string {
	return sampleConfig
}

func (j *Journald) Init() error {
	// Set defaults
	if j.Binary == "" {
		j.Binary = "journalctl"
	}
	if j.InitialReadOffset == "" {
		j.InitialReadOffset = "saved-or-end"
	}
	if j.MaxUndeliveredEntries == 0 {
		j.MaxUndeliveredEntries = 1000
	}
	if j.RestartDelay == 0 {
		j.RestartDelay = config.Duration(5 * time.Second)
	}

	// Check the settings
	switch j.InitialReadOffset {
	case "beginning", "end", "saved-or-beginning", "saved-or-end":
	default:
		return fmt.Errorf("invalid 'initial_read_offset' setting %q", j.InitialReadOffset)
	}
	if j.MaxUndeliveredEntries < 0 {
		return errors.New("'max_undelivered_entries' must be positive")
	}
	if j.RestartDelay < 0 {
		return errors.New("'restart_delay' must not be negative")
	}
	if err := checkPriority(j.Priority); err != nil {
		return err
	}
	for _, m := range j.Matches {
		if strings.TrimSpace(m) == "" {
			return errors.New("empty entry in 'matches'")
		}
	}

	j.tagKeys = make(map[string]string, len(j.TagFields))
	for _, f := range j.TagFields {
		if f == "" {
			return errors.New("empty entry in 'tag_fields'")
		}
		j.tagKeys[f] = keyName(f)
	}
	fieldKeys, err := filter.Compile(j.JournalFields)
	if err != nil {
		return fmt.Errorf("creating filter for 'journal_fields' failed: %w", err)
	}
	j.fieldKeys = fieldKeys

	// Check for the required executables and construct the command
	j.command = j.Binary
	if j.UseSudo {
		j.command = "sudo"
		j.baseArgs = append(j.baseArgs, j.Binary)
	}
	if _, err := exec.LookPath(j.command); err != nil {
		return fmt.Errorf("looking up %q failed: %w", j.command, err)
	}

	// Always use the JSON output with all fields (journalctl replaces fields
	// larger than 4096 bytes by null otherwise) and follow the journal.
	j.baseArgs = append(j.baseArgs, "--output=json", "--all", "--follow", "--no-pager")
	for _, u := range j.Units {
		j.baseArgs = append(j.baseArgs, "--unit="+u)
	}
	for _, u := range j.UserUnits {
		j.baseArgs = append(j.baseArgs, "--user-unit="+u)
	}
	if j.Priority != "" {
		j.baseArgs = append(j.baseArgs, "--priority="+j.Priority)
	}
	if j.Directory != "" {
		j.baseArgs = append(j.baseArgs, "--directory="+j.Directory)
	}
	if j.Namespace != "" {
		j.baseArgs = append(j.baseArgs, "--namespace="+j.Namespace)
	}

	j.tracker = newCursorTracker()

	return nil
}

func (j *Journald) GetState() any {
	return j.tracker.committed()
}

func (j *Journald) SetState(state any) error {
	cursor, ok := state.(string)
	if !ok {
		return fmt.Errorf("invalid state type %T", state)
	}
	j.tracker.setCommitted(cursor)
	return nil
}

func (j *Journald) Start(acc telegraf.Accumulator) error {
	// Only continue at the persisted cursor if requested
	switch j.InitialReadOffset {
	case "saved-or-beginning", "saved-or-end":
		j.readCursor = j.tracker.committed()
	default:
		j.readCursor = ""
	}

	j.acc = acc.WithTracking(j.MaxUndeliveredEntries)
	j.sem = make(chan struct{}, j.MaxUndeliveredEntries)

	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel

	j.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case info := <-j.acc.Delivered():
				j.tracker.done(info.ID())
				<-j.sem
			}
		}
	})

	j.wg.Go(func() {
		j.run(ctx)
	})

	return nil
}

func (j *Journald) Stop() {
	if j.cancel != nil {
		j.cancel()
	}
	j.wg.Wait()
}

func (*Journald) Gather(telegraf.Accumulator) error {
	return nil
}

func (j *Journald) run(ctx context.Context) {
	for {
		err := j.follow(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			j.acc.AddError(err)
		} else {
			j.Log.Warn("journalctl exited unexpectedly")
		}
		j.Log.Debugf("Restarting journalctl in %s", time.Duration(j.RestartDelay))

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(j.RestartDelay)):
		}
	}
}

func (j *Journald) args() []string {
	args := make([]string, 0, len(j.baseArgs)+2+len(j.Matches))
	args = append(args, j.baseArgs...)
	switch {
	case j.readCursor != "":
		args = append(args, "--after-cursor="+j.readCursor)
	case j.InitialReadOffset == "beginning" || j.InitialReadOffset == "saved-or-beginning":
		// Following implies showing the last ten lines, so explicitly request
		// all of them to start at the oldest entry
		args = append(args, "--lines=all")
	default:
		args = append(args, "--lines=0")
	}
	return append(args, j.Matches...)
}

func (j *Journald) follow(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, j.command, j.args()...)
	// Terminate gracefully so sudo can forward the signal to journalctl and
	// kill the process if it does not stop in time.
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = 5 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("getting stdout of journalctl failed: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("getting stderr of journalctl failed: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting journalctl failed: %w", err)
	}

	// Forward the messages of journalctl, e.g. about missing permissions
	var cursorInvalid bool
	var stderrWg sync.WaitGroup
	stderrWg.Go(func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, seekCursorFailed) {
				cursorInvalid = true
			}
			j.Log.Warnf("journalctl: %s", line)
		}
	})

	reader := bufio.NewReaderSize(stdout, 64*1024)
	for {
		line, rerr := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			if !j.process(ctx, line) {
				break
			}
		}
		if rerr != nil {
			if !errors.Is(rerr, io.EOF) {
				j.acc.AddError(fmt.Errorf("reading journalctl output failed: %w", rerr))
			}
			break
		}
	}
	stderrWg.Wait()
	err = cmd.Wait()

	if cursorInvalid && ctx.Err() == nil {
		j.Log.Warnf("Cannot continue after cursor %q, ignoring the cursor", j.readCursor)
		j.readCursor = ""
	}
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("journalctl failed: %w", err)
	}
	return nil
}

// process parses a single entry and hands it to the accumulator. It returns
// false if the plugin is stopping.
func (j *Journald) process(ctx context.Context, line []byte) bool {
	m, cursor, err := j.parse(line)
	if err != nil {
		j.acc.AddError(err)
		return true
	}

	// Limit the number of undelivered entries
	select {
	case j.sem <- struct{}{}:
	case <-ctx.Done():
		return false
	}
	id := j.acc.AddTrackingMetricGroup([]telegraf.Metric{m})
	j.tracker.add(id, cursor)
	if cursor != "" {
		j.readCursor = cursor
	}
	return true
}

func init() {
	inputs.Add("journald", func() telegraf.Input {
		return &Journald{}
	})
}
