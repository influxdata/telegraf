package ts3

import (
	"fmt"
	"strings"
)

// Cmd represents a TeamSpeak 3 ServerQuery command.
type Cmd struct {
	cmd      string
	args     []CmdArg
	options  []string
	response interface{}
}

// NewCmd creates a new Cmd.
func NewCmd(cmd string) *Cmd {
	return &Cmd{cmd: cmd}
}

// WithArgs sets the command Args.
func (c *Cmd) WithArgs(args ...CmdArg) *Cmd {
	c.args = args
	return c
}

// WithOptions sets the command Options.
func (c *Cmd) WithOptions(options ...string) *Cmd {
	c.options = options
	return c
}

// WithResponse sets the command Response which will have the data returned from the server decoded into it.
func (c *Cmd) WithResponse(r interface{}) *Cmd {
	c.response = r
	return c
}

func (c *Cmd) String() string {
	args := make([]interface{}, 1, len(c.args)+len(c.options)+1)
	args[0] = c.cmd
	for _, v := range c.args {
		args = append(args, v.ArgString())
	}
	for _, v := range c.options {
		args = append(args, v)
	}
	return fmt.Sprintln(args...)
}

// CmdArg is implemented by types which can be used as a command argument.
type CmdArg interface {
	ArgString() string
}

// ArgGroup represents a group of TeamSpeak 3 ServerQuery command arguments.
type ArgGroup struct {
	grp []CmdArg
}

// NewArgGroup returns a new ArgGroup.
func NewArgGroup(args ...CmdArg) *ArgGroup {
	return &ArgGroup{grp: args}
}

// ArgString implements CmdArg.
func (ag *ArgGroup) ArgString() string {
	args := make([]string, len(ag.grp))
	for i, arg := range ag.grp {
		args[i] = arg.ArgString()
	}
	return strings.Join(args, "|")
}

// ArgSet represents a set of TeamSpeak 3 ServerQuery command arguments.
type ArgSet struct {
	set []CmdArg
}

// NewArgSet returns a new ArgSet.
func NewArgSet(args ...CmdArg) *ArgSet {
	return &ArgSet{set: args}
}

// ArgString implements CmdArg.
func (ag *ArgSet) ArgString() string {
	args := make([]string, len(ag.set))
	for i, arg := range ag.set {
		args[i] = arg.ArgString()
	}
	return strings.Join(args, " ")
}

// Arg represents a TeamSpeak 3 ServerQuery command argument.
// Args automatically escape white space and special characters before being sent to the server.
type Arg struct {
	key string
	val string
}

// NewArg returns a new Arg with key val.
func NewArg(key string, val interface{}) *Arg {
	return &Arg{key: key, val: fmt.Sprint(val)}
}

// ArgString implements CmdArg
func (a *Arg) ArgString() string {
	return fmt.Sprintf("%v=%v", encoder.Replace(a.key), encoder.Replace(a.val))
}
