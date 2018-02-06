package rollbar

import "strconv"

type Event interface {
	Tags() map[string]string
	Fields() map[string]interface{}
}

type DummyEvent struct {
	EventName string `json:"event_name"`
}

type NewItemDataItemLastOccurence struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type NewItemDataItem struct {
	Id            int                          `json:"id"`
	Environment   string                       `json:"environment"`
	ProjectId     int                          `json:"project_id"`
	LastOccurence NewItemDataItemLastOccurence `json:"last_occurrence"`
}

type NewItemData struct {
	Item NewItemDataItem `json:"item"`
}

type NewItem struct {
	EventName string      `json:"event_name"`
	Data      NewItemData `json:"data"`
}

func (ni *NewItem) Tags() map[string]string {
	return map[string]string{
		"event":       ni.EventName,
		"environment": ni.Data.Item.Environment,
		"project_id":  strconv.Itoa(ni.Data.Item.ProjectId),
		"language":    ni.Data.Item.LastOccurence.Language,
		"level":       ni.Data.Item.LastOccurence.Level,
	}
}

func (ni *NewItem) Fields() map[string]interface{} {
	return map[string]interface{}{
		"id": ni.Data.Item.Id,
	}
}

type OccurrenceDataOccurrence struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type OccurrenceDataItem struct {
	Id          int    `json:"id"`
	Environment string `json:"environment"`
	ProjectId   int    `json:"project_id"`
}

type OccurrenceData struct {
	Item       OccurrenceDataItem       `json:"item"`
	Occurrence OccurrenceDataOccurrence `json:"occurrence"`
}

type Occurrence struct {
	EventName string         `json:"event_name"`
	Data      OccurrenceData `json:"data"`
}

func (o *Occurrence) Tags() map[string]string {
	return map[string]string{
		"event":       o.EventName,
		"environment": o.Data.Item.Environment,
		"project_id":  strconv.Itoa(o.Data.Item.ProjectId),
		"language":    o.Data.Occurrence.Language,
		"level":       o.Data.Occurrence.Level,
	}
}

func (o *Occurrence) Fields() map[string]interface{} {
	return map[string]interface{}{
		"id": o.Data.Item.Id,
	}
}

type DeployDataDeploy struct {
	Id          int    `json:"id"`
	Environment string `json:"environment"`
	ProjectId   int    `json:"project_id"`
}

type DeployData struct {
	Deploy DeployDataDeploy `json:"deploy"`
}

type Deploy struct {
	EventName string     `json:"event_name"`
	Data      DeployData `json:"data"`
}

func (ni *Deploy) Tags() map[string]string {
	return map[string]string{
		"event":       ni.EventName,
		"environment": ni.Data.Deploy.Environment,
		"project_id":  strconv.Itoa(ni.Data.Deploy.ProjectId),
	}
}

func (ni *Deploy) Fields() map[string]interface{} {
	return map[string]interface{}{
		"id": ni.Data.Deploy.Id,
	}
}
