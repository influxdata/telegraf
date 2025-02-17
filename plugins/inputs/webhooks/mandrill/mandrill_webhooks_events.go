package mandrill

type mandrillEvent struct {
	EventName string `json:"event"`
	TimeStamp int64  `json:"ts"`
	ID        string `json:"_id"`
}

func (me *mandrillEvent) tags() map[string]string {
	return map[string]string{
		"event": me.EventName,
	}
}

func (me *mandrillEvent) fields() map[string]interface{} {
	return map[string]interface{}{
		"id": me.ID,
	}
}
