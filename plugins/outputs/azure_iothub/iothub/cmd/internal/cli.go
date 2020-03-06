package internal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidUsage when returned by a Handler the usage message is displayed.
var ErrInvalidUsage = errors.New("invalid usage")

// Command is a cli subcommand.
type Command struct {
	Name      string
	Args      []string
	Desc      string
	Handler   HandlerFunc
	ParseFunc func(*flag.FlagSet)
}

// HandlerFunc is a subcommand handler, fs is already parsed.
type HandlerFunc func(args []string) error

// FlagFunc prepares fs for parsing, setting flags.
type FlagFunc func(fs *flag.FlagSet)

// CLI is a cli subcommands executor.
type CLI struct {
	desc string
	cmds []*Command
	main FlagFunc
}

// New creates new cli executor.
func New(desc string, f FlagFunc, cmds []*Command) *CLI {
	return &CLI{
		desc: desc,
		cmds: cmds,
		main: f,
	}
}

// Run runs one or the given commands based on argv.
//
// Panics if argv is empty.
//
// If ErrInvalidUsage is returned there's no need to print it,
// the usage message is already sent to STDERR.
func (r *CLI) Run(argv []string) error {
	if len(argv) == 0 {
		panic("empty argv")
	}

	sm := flag.NewFlagSet(filepath.Base(argv[0]), flag.ContinueOnError)
	if r.main != nil {
		r.main(sm)
	}
	sm.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [option...] COMMAND [option...] [arg...]

%s

Commands:
`, sm.Name(), r.desc)
		for _, cmd := range r.cmds {
			fmt.Fprintf(os.Stderr, "  %-25s %s\n", cmd.Name, cmd.Desc)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Common options: ")
		sm.PrintDefaults()
	}

	if err := sm.Parse(argv[1:]); err != nil {
		if err == flag.ErrHelp {
			return ErrInvalidUsage
		}
		return err
	}

	if sm.NArg() == 0 {
		sm.Usage()
		return ErrInvalidUsage
	}

	cmd := r.findCommand(sm.Arg(0))
	if cmd == nil {
		sm.Usage()
		return ErrInvalidUsage
	}

	sc := flag.NewFlagSet(sm.Arg(0), flag.ContinueOnError)
	if cmd.ParseFunc != nil {
		cmd.ParseFunc(sc)
	}
	sc.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [option...] %s ", sm.Name(), sm.Arg(0))
		if hasFlags(sc) {
			fmt.Fprintf(os.Stderr, "[option...] ")
		}
		fmt.Fprintln(os.Stderr, strings.Join(cmd.Args, " "))
		if hasFlags(sc) {
			fmt.Fprintln(os.Stderr, "\nOptions:")
			sc.PrintDefaults()
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Common options: ")
		sm.PrintDefaults()
	}
	if err := sc.Parse(sm.Args()[1:]); err != nil {
		if err == flag.ErrHelp {
			return ErrInvalidUsage
		}
		return err
	}
	if len(cmd.Args) != sc.NArg() {
		sc.Usage()
		return ErrInvalidUsage
	}
	if err := cmd.Handler(sc.Args()); err != nil {
		if err == ErrInvalidUsage {
			sc.Usage()
		}
		return err
	}
	return nil
}

func hasFlags(fs *flag.FlagSet) bool {
	var has bool
	fs.VisitAll(func(f *flag.Flag) {
		has = true
	})
	return has
}

func (r *CLI) findCommand(k string) *Command {
	for _, cmd := range r.cmds {
		if cmd.Name == k {
			return cmd
		}
	}
	return nil
}

// OutputLine prints the given string to stdout appending a new-line char.
func OutputLine(format string) error {
	_, err := fmt.Println(format)
	return err
}

func Output(v interface{}, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(v)
	case "json-pretty":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "\t")
		return enc.Encode(v)
	default:
		return fmt.Errorf("unknown output format: %q", format)
	}
}
