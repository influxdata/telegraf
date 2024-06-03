//go:build !windows

package main

import (
	"io"

	"github.com/urfave/cli/v2"
)

func cliFlags() []cli.Flag {
	return []cli.Flag{}
}

func getServiceCommands(io.Writer) []*cli.Command {
	return nil
}
