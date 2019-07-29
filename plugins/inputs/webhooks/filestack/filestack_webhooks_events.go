package filestack

import "strconv"

type FilestackEvent struct {
	Action    string `json:"action"`
	TimeStamp int64  `json:"timestamp"`
	Id        int    `json:"id"`
}

func (fe *FilestackEvent) Tags() map[string]string {
	return map[string]string{
		"action": fe.Action,
	}
}

func (fe *FilestackEvent) Fields() map[string]interface{} {
	return map[string]interface{}{
		"id": strconv.Itoa(fe.Id),
	}
}
