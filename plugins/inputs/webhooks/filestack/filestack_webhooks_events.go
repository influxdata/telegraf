package filestack

import "strconv"

type DialogEvent struct {
	Action    string `json:"action"`
	TimeStamp int64  `json:"timestamp"`
	Id        int    `json:"id"`
}

func (de *DialogEvent) Tags() map[string]string {
	return map[string]string{
		"action": de.Action,
	}
}

func (de *DialogEvent) Fields() map[string]interface{} {
	return map[string]interface{}{
		"id": strconv.Itoa(de.Id),
	}
}
