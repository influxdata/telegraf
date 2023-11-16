package dpdk

import (
	"fmt"
	"strings"
)

type linkStatus int64

const (
	DOWN linkStatus = iota
	UP
)

const (
	ethdevLinkStatusCommand    = "/ethdev/link_status"
	linkStatusStringFieldName  = "status"
	linkStatusIntegerFieldName = "status_i"
)

var (
	linkStatusMap = map[string]linkStatus{
		"down": DOWN,
		"up":   UP,
	}

	processCmdMap = map[string]func(map[string]interface{}) error{
		ethdevLinkStatusCommand: processLinkStatusCmd,
	}
)

func processCommandResponse(command string, data map[string]interface{}) error {
	handler, ok := processCmdMap[command]
	if ok {
		return handler(data)
	}
	return nil
}

func processLinkStatusCmd(data map[string]interface{}) error {
	status, ok := data[linkStatusStringFieldName].(string)
	if !ok {
		return fmt.Errorf("can't find or parse '%s' field", linkStatusStringFieldName)
	}

	parsedLinkStatus, ok := parseLinkStatus(status)
	if !ok {
		return fmt.Errorf("can't parse linkStatus: unknown value: %s", status)
	}

	data[linkStatusIntegerFieldName] = parsedLinkStatus.asInt64()
	return nil
}

func parseLinkStatus(s string) (linkStatus, bool) {
	value, ok := linkStatusMap[strings.ToLower(s)]
	return value, ok
}

func (ls linkStatus) asInt64() int64 {
	return int64(ls)
}
