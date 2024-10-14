//go:build linux

package dpdk

import (
	"fmt"
	"strings"
)

type linkStatus int64

const (
	down linkStatus = iota
	up
)

const (
	ethdevLinkStatusCommand    = "/ethdev/link_status"
	linkStatusStringFieldName  = "status"
	linkStatusIntegerFieldName = "link_status"
)

var (
	linkStatusMap = map[string]linkStatus{
		"down": down,
		"up":   up,
	}
)

func processCommandResponse(command string, data map[string]interface{}) error {
	if command == ethdevLinkStatusCommand {
		return processLinkStatusCmd(data)
	}
	return nil
}

func processLinkStatusCmd(data map[string]interface{}) error {
	status, ok := data[linkStatusStringFieldName].(string)
	if !ok {
		return fmt.Errorf("can't find or parse %q field", linkStatusStringFieldName)
	}

	parsedLinkStatus, ok := parseLinkStatus(status)
	if !ok {
		return fmt.Errorf("can't parse linkStatus: unknown value: %q", status)
	}

	data[linkStatusIntegerFieldName] = int64(parsedLinkStatus)
	return nil
}

func parseLinkStatus(s string) (linkStatus, bool) {
	value, ok := linkStatusMap[strings.ToLower(s)]
	return value, ok
}
