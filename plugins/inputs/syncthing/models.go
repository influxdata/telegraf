package syncthing

import (
	"time"
)

// SystemConnections returns the list of configured devices and some metadata associated with them. The list also contains the local device itself as not connected.
// The connection types are TCP (Client), TCP (Server), Relay (Client) and Relay (Server).
// GET /rest/system/connections
// https://docs.syncthing.net/rest/system-connections-get.html
type SystemConnections struct {
	Connections map[string]*Connection `json:"connections"`
}

type Connection struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	At            string `json:"at"`
	ClientVersion string `json:"clientVersion"`
	Connected     bool   `json:"connected"`
	Crypto        string `json:"crypto"`
	InBytesTotal  int    `json:"inBytesTotal"`
	OutBytesTotal int    `json:"outBytesTotal"`
	Paused        bool   `json:"paused"`
	Type          string `json:"type"`
	DeviceName    string `json:"device_name"`
}

// SystemVersion Returns the current version information.
// GET /rest/system/version
type SystemVersion struct {
	Arch        string `json:"arch"`
	LongVersion string `json:"longVersion"`
	Os          string `json:"os"`
	Version     string `json:"version"`
}

// Need Takes one mandatory parameter, folder, and returns lists of files which are needed by this device in order for
// it to become in sync.
// GET /rest/db/need
type Need struct {
	// DO NOT want these because we don't need the specific information
	// Progress []FileInfo `json:"progress"`
	// Queued   []FileInfo `json:"queued"`
	// Rest     []FileInfo `json:"rest"`
	Total int `json:"total"`
}

// SystemConfig is the current system configuration
// GET /rest/system/config
type SystemConfig struct {
	Version        int               `json:"version"`
	Folders        []*Folders        `json:"folders"`
	Devices        []*Device         `json:"devices"`
	PendingDevices []*PendingFolders `json:"pendingDevices"`
}

func (s *SystemConfig) DeviceByID(id string) *Device {
	for _, d := range s.Devices {
		if d.ID == id {
			return d
		}
	}
	return nil
}

type Folders struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Path   string `json:"path"`
	Paused bool   `json:"paused"`
}

type PendingFolders struct {
	Time  time.Time `json:"time"`
	ID    string    `json:"id"`
	Label string    `json:"label"`
}

type Device struct {
	ID             string           `json:"deviceID"`
	Name           string           `json:"name"`
	Paused         bool             `json:"paused"`
	PendingFolders []PendingFolders `json:"pendingFolders"`
	MaxRequestKiB  int              `json:"maxRequestKiB"`
}

// SystemStatus returns information about current system status and resource usage. The CPU percent value has been deprecated from the API and will always report 0.
// GET /rest/system/status
// https://docs.syncthing.net/rest/system-status-get.html
type SystemStatus struct {
	MyID      string    `json:"myID"`
	StartTime time.Time `json:"startTime"`
}
