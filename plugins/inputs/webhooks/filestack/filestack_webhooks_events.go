package filestack

import "strconv"

type filestackEvent struct {
	Action    string `json:"action"`
	TimeStamp int64  `json:"timestamp"`
	ID        int    `json:"id"`
}

func (fe *filestackEvent) tags() map[string]string {
	return map[string]string{
		"action": fe.Action,
	}
}

func (fe *filestackEvent) fields() map[string]interface{} {
	return map[string]interface{}{
		"id": strconv.Itoa(fe.ID),
	}
}
