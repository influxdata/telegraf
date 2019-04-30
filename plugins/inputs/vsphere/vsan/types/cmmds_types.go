package types

type CmmdsEntity struct {
	UUID    string      `json:"uuid"`
	Owner   string      `json:"owner"`
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Cmmds struct {
	Res []CmmdsEntity `json:"result"`
}
