package hc2

// LinkRoomsSections links rooms to sections
type LinkRoomsSections struct {
	Name      string
	SectionID uint16
}

// Sections contains sections informations
type Sections struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

// Rooms contains rooms informations
type Rooms struct {
	ID        uint16 `json:"id"`
	Name      string `json:"name"`
	SectionID uint16 `json:"sectionID"`
}

// Devices contains devices informations
type Devices struct {
	ID         uint16 `json:"id"`
	Name       string `json:"name"`
	RoomID     uint16 `json:"roomID"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	Properties struct {
		BatteryLevel *string     `json:"batteryLevel"`
		Dead         string      `json:"dead"`
		Energy       *string     `json:"energy"`
		Power        *string     `json:"power"`
		Value        interface{} `json:"value"`
		Value2       *string     `json:"value2"`
	} `json:"properties"`
}
