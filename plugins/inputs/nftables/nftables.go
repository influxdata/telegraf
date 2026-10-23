//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package nftables

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Nftables struct {
	UseSudo bool            `toml:"use_sudo"`
	Binary  string          `toml:"binary"`
	Tables  []string        `toml:"tables"`
	Include []string        `toml:"include"`
	Log     telegraf.Logger `toml:"-"`

	args  []string
	terse bool
}

func (*Nftables) SampleConfig() string {
	return sampleConfig
}

func (n *Nftables) Init() error {
	// Set defaults
	if len(n.Tables) == 0 {
		n.Tables = []string{"filter"}
	}
	if n.Binary == "" {
		n.Binary = "nft"
	}
	if len(n.Include) == 0 {
		n.Include = []string{"anonymous-counters"}
	}

	// Check includes
	includesSet := make(map[string]bool, len(n.Include))
	for _, include := range n.Include {
		if includesSet[include] {
			return fmt.Errorf("duplicate include %q", include)
		}
		includesSet[include] = true
		switch include {
		case "anonymous-counters", "counters", "sets":
			// Do nothing, those are valid
		default:
			return fmt.Errorf("unknown include %q", include)
		}
	}

	// Construct the command
	n.args = make([]string, 0, 6)
	if n.UseSudo {
		n.args = append(n.args, n.Binary)
		n.Binary = "sudo"
	}
	return nil
}

func (n *Nftables) Start(telegraf.Accumulator) error {
	// Use --terse to avoid dumping the elements of sets, unless sets are
	// monitored and the nft version does not support counting elements.
	if !slices.Contains(n.Include, "sets") {
		n.terse = true
	} else {
		version, err := n.version()
		if err != nil {
			n.Log.Warnf("Failed to determine nft --version, will not use --terse: %v", err)
		} else if version.GreaterThanEqual(semver.New(1, 1, 7, "", "")) {
			n.terse = true
		}
	}
	return nil
}

func (*Nftables) Stop() {}

func (n *Nftables) version() (*semver.Version, error) {
	args := append(n.args, "--version")
	out, err := exec.Command(n.Binary, args...).Output()
	if err != nil {
		if oserr, ok := errors.AsType[*exec.ExitError](err); ok {
			buf, _, _ := bytes.Cut(oserr.Stderr, []byte("\n"))
			return nil, fmt.Errorf("error executing nft --version command: %w (%s)", err, bytes.TrimSpace(buf))
		}
		return nil, fmt.Errorf("error executing nft --version command: %w", err)
	}

	// Parse version from output like "nftables v1.1.6 (Commodore Bullmoose #7)"
	fields := strings.Fields(string(out))
	if len(fields) < 2 || fields[0] != "nftables" {
		return nil, fmt.Errorf("unexpected version output %q", strings.TrimSpace(string(out)))
	}
	return semver.NewVersion(fields[1])
}

func (n *Nftables) Gather(acc telegraf.Accumulator) error {
	for _, table := range n.Tables {
		acc.AddError(n.gatherTable(acc, table))
	}
	return nil
}

func (n *Nftables) gatherTable(acc telegraf.Accumulator, name string) error {
	// Run the nft command
	args := n.args
	if n.terse {
		args = append(args, "--terse")
	}
	args = append(args, "--json", "list", "table", name)
	c := exec.Command(n.Binary, args...)
	out, err := c.Output()
	if err != nil {
		if oserr, ok := errors.AsType[*exec.ExitError](err); ok {
			buf, _, _ := bytes.Cut(oserr.Stderr, []byte("\n"))
			msg := string(bytes.TrimSpace(buf))
			if msg == "Error: No such file or directory" {
				return fmt.Errorf("table %q does not exist", name)
			}
			return fmt.Errorf("error executing nft command: %w (%s)", err, msg)
		}
		return fmt.Errorf("error executing nft command: %w", err)
	}

	// Parse the result into metrics and add them to the accumulator
	var nftable table
	if err := json.Unmarshal(out, &nftable); err != nil {
		return fmt.Errorf("parsing command output failed: %w", err)
	}

	for _, include := range n.Include {
		switch include {
		case "anonymous-counters":
			for _, rule := range nftable.Rules {
				if len(rule.Comment) == 0 {
					continue
				}
				for _, expr := range rule.Exprs {
					if expr.Cntr == nil || expr.Cntr.isNamedRef {
						continue
					}
					fields := map[string]any{
						"bytes": expr.Cntr.Bytes,
						"pkts":  expr.Cntr.Packets,
					}
					tags := map[string]string{
						"table": rule.Table,
						"chain": rule.Chain,
						"rule":  rule.Comment,
					}
					acc.AddFields("nftables", fields, tags)
				}
			}
		case "counters":
			for _, counter := range nftable.Counters {
				fields := map[string]any{
					"bytes": counter.Bytes,
					"pkts":  counter.Packets,
				}
				tags := map[string]string{
					"table":   counter.Table,
					"counter": counter.Name,
				}
				acc.AddFields("nftables", fields, tags)
			}
		case "sets":
			for _, set := range nftable.Sets {
				// With --terse nft only reports the element count for sets
				// declared with a size and only if the kernel provides it,
				// so a missing count does not mean the set is empty. List
				// the set itself to count its elements in this case.
				count := set.count()
				if n.terse && set.Count == nil {
					c, err := n.countSetElements(set)
					if err != nil {
						return err
					}
					count = c
				}
				fields := map[string]any{
					"count": count,
				}
				tags := map[string]string{
					"table": set.Table,
					"set":   set.Name,
				}
				acc.AddFields("nftables", fields, tags)
			}
		}
	}
	return nil
}

func (n *Nftables) countSetElements(set *namedSet) (int64, error) {
	// List the set without --terse to get its elements
	args := append(n.args, "--json", "list", "set", set.Family, set.Table, set.Name)
	out, err := exec.Command(n.Binary, args...).Output()
	if err != nil {
		if oserr, ok := errors.AsType[*exec.ExitError](err); ok {
			buf, _, _ := bytes.Cut(oserr.Stderr, []byte("\n"))
			return 0, fmt.Errorf("error listing set %q: %w (%s)", set.Name, err, bytes.TrimSpace(buf))
		}
		return 0, fmt.Errorf("error listing set %q: %w", set.Name, err)
	}

	var nftable table
	if err := json.Unmarshal(out, &nftable); err != nil {
		return 0, fmt.Errorf("parsing output for set %q failed: %w", set.Name, err)
	}
	if len(nftable.Sets) != 1 {
		return 0, fmt.Errorf("expected one set in output for %q but got %d", set.Name, len(nftable.Sets))
	}
	return nftable.Sets[0].count(), nil
}

func init() {
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{}
	})
}
