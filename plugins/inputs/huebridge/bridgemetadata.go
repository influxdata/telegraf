package huebridge

import (
	"fmt"
	"maps"
	"net/http"

	"github.com/tdrn-org/go-hue"
)

type bridgeMetadata struct {
	resourceTree    map[string]string
	deviceNames     map[string]string
	roomAssignments map[string]string
}

func fetchMetadata(bridgeClient hue.BridgeClient, manualRoomAsignments map[string]string) (*bridgeMetadata, error) {
	resourceTree, err := fetchResourceTree(bridgeClient)
	if err != nil {
		return nil, err
	}
	deviceNames, err := fetchDeviceNames(bridgeClient)
	if err != nil {
		return nil, err
	}
	roomAssignments, err := fetchRoomAssignments(bridgeClient)
	if err != nil {
		return nil, err
	}
	maps.Copy(roomAssignments, manualRoomAsignments)
	metadata := &bridgeMetadata{
		resourceTree:    resourceTree,
		deviceNames:     deviceNames,
		roomAssignments: roomAssignments,
	}
	return metadata, nil
}

func fetchResourceTree(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getResourcesResponse, err := bridgeClient.GetResources()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge resources on %s: %w", bridgeClient.Url().Redacted(), err)
	}
	if getResourcesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge resources from %s: %s", bridgeClient.Url().Redacted(), getResourcesResponse.HTTPResponse.Status)
	}
	responseData := getResourcesResponse.JSON200.Data
	if responseData == nil {
		return make(map[string]string), nil
	}
	tree := make(map[string]string, len(*responseData))
	for _, resource := range *responseData {
		if resource.Owner != nil {
			tree[*resource.Id] = *resource.Owner.Rid
		}
	}
	return tree, nil
}

func fetchDeviceNames(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getDevicesResponse, err := bridgeClient.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge devices on %s: %w", bridgeClient.Url().Redacted(), err)
	}
	if getDevicesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge devices from %s: %s", bridgeClient.Url().Redacted(), getDevicesResponse.HTTPResponse.Status)
	}
	responseData := getDevicesResponse.JSON200.Data
	if responseData == nil {
		return make(map[string]string), nil
	}
	names := make(map[string]string, len(*responseData))
	for _, device := range *responseData {
		names[*device.Id] = *device.Metadata.Name
	}
	return names, nil
}

func fetchRoomAssignments(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getRoomsResponse, err := bridgeClient.GetRooms()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge rooms on %s: %w", bridgeClient.Url().Redacted(), err)
	}
	if getRoomsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge rooms from %s: %s", bridgeClient.Url().Redacted(), getRoomsResponse.HTTPResponse.Status)
	}
	responseData := getRoomsResponse.JSON200.Data
	if responseData == nil {
		return make(map[string]string), nil
	}
	assignments := make(map[string]string, len(*responseData))
	for _, roomGet := range *responseData {
		for _, children := range *roomGet.Children {
			assignments[*children.Rid] = *roomGet.Metadata.Name
		}
	}
	return assignments, nil
}

func (metadata *bridgeMetadata) resolveResourceRoom(resourceId string, resourceName string) string {
	roomName := metadata.roomAssignments[resourceName]
	if roomName != "" {
		return roomName
	}
	// If resource does not have a room assigned directly, iterate upwards via
	// its owners until we find a room or there is no more owner. The latter
	// may happen (e.g. for Motion Sensors) resulting in room name
	// "<unassigned>".
	currentResourceId := resourceId
	for {
		// Try next owner
		currentResourceId = metadata.resourceTree[currentResourceId]
		if currentResourceId == "" {
			// No owner left but no room found
			break
		}
		roomName = metadata.roomAssignments[currentResourceId]
		if roomName != "" {
			// Room name found, done
			return roomName
		}
	}
	return "<unassigned>"
}

func (metadata *bridgeMetadata) resolveDeviceName(resourceId string) string {
	deviceName := metadata.deviceNames[resourceId]
	if deviceName != "" {
		return deviceName
	}
	// If resource does not have a device name assigned directly, iterate
	// upwards via its owners until we find a room or there is no more
	// owner. The latter may happen resulting in device name "<undefined>".
	currentResourceId := resourceId
	for {
		// Try next owner
		currentResourceId = metadata.resourceTree[currentResourceId]
		if currentResourceId == "" {
			// No owner left but no device found
			break
		}
		deviceName = metadata.deviceNames[currentResourceId]
		if deviceName != "" {
			// Device name found, done
			return deviceName
		}
	}
	return "<undefined>"
}
