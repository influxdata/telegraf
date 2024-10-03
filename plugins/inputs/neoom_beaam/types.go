package neoom_beaam

type site struct {
	EnergyFlow struct {
		DataPoints map[string]datapoint `json:"dataPoints"`
	} `json:"energyFlow"`
	Things map[string]thingDefinition `json:"things"`
}

type siteState struct {
	EnergyFlow struct {
		States []state `json:"states"`
	} `json:"energyFlow"`
}

type thingDefinition struct {
	id         string
	Name       string               `json:"type"`
	DataPoints map[string]datapoint `json:"dataPoints"`
}

type thingState struct {
	ID     string  `json:"thingId"`
	States []state `json:"states"`
}

type datapoint struct {
	Key      string `json:"key"`
	DataType string `json:"dataType"`
	Unit     string `json:"unitOfMeasure"`
}

type state struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	DataPointID string      `json:"dataPointId"`
	Timestamp   float64     `json:"ts"`
}
