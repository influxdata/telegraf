package rollbar

import "strconv"

type event interface {
	tags() map[string]string
	fields() map[string]interface{}
}

type dummyEvent struct {
	EventName string `json:"event_name"`
}

type newItemDataItemLastOccurrence struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type newItemDataItem struct {
	ID             int                           `json:"id"`
	Environment    string                        `json:"environment"`
	ProjectID      int                           `json:"project_id"`
	LastOccurrence newItemDataItemLastOccurrence `json:"last_occurrence"`
}

type newItemData struct {
	Item newItemDataItem `json:"item"`
}

type newItem struct {
	EventName string      `json:"event_name"`
	Data      newItemData `json:"data"`
}

func (ni *newItem) tags() map[string]string {
	return map[string]string{
		"event":       ni.EventName,
		"environment": ni.Data.Item.Environment,
		"project_id":  strconv.Itoa(ni.Data.Item.ProjectID),
		"language":    ni.Data.Item.LastOccurrence.Language,
		"level":       ni.Data.Item.LastOccurrence.Level,
	}
}

func (ni *newItem) fields() map[string]interface{} {
	return map[string]interface{}{
		"id": ni.Data.Item.ID,
	}
}

type occurrenceDataOccurrence struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type occurrenceDataItem struct {
	ID          int    `json:"id"`
	Environment string `json:"environment"`
	ProjectID   int    `json:"project_id"`
}

type occurrenceData struct {
	Item       occurrenceDataItem       `json:"item"`
	Occurrence occurrenceDataOccurrence `json:"occurrence"`
}

type occurrence struct {
	EventName string         `json:"event_name"`
	Data      occurrenceData `json:"data"`
}

func (o *occurrence) tags() map[string]string {
	return map[string]string{
		"event":       o.EventName,
		"environment": o.Data.Item.Environment,
		"project_id":  strconv.Itoa(o.Data.Item.ProjectID),
		"language":    o.Data.Occurrence.Language,
		"level":       o.Data.Occurrence.Level,
	}
}

func (o *occurrence) fields() map[string]interface{} {
	return map[string]interface{}{
		"id": o.Data.Item.ID,
	}
}

type deployDataDeploy struct {
	ID          int    `json:"id"`
	Environment string `json:"environment"`
	ProjectID   int    `json:"project_id"`
}

type deployData struct {
	Deploy deployDataDeploy `json:"deploy"`
}

type deploy struct {
	EventName string     `json:"event_name"`
	Data      deployData `json:"data"`
}

func (ni *deploy) tags() map[string]string {
	return map[string]string{
		"event":       ni.EventName,
		"environment": ni.Data.Deploy.Environment,
		"project_id":  strconv.Itoa(ni.Data.Deploy.ProjectID),
	}
}

func (ni *deploy) fields() map[string]interface{} {
	return map[string]interface{}{
		"id": ni.Data.Deploy.ID,
	}
}
