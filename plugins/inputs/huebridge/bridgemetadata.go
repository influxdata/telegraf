package huebridge

import (
	"fmt"
	"maps"
	"net/http"

	"github.com/tdrn-org/go-hue"
)

type BridgeMetadata struct {
	resourceTree    map[string]string
	deviceNames     map[string]string
	roomAssignments map[string]string
}

func FetchMetadata(bridgeClient hue.BridgeClient, manualRoomAsignments map[string]string) (*BridgeMetadata, error) {
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
	metadata := &BridgeMetadata{
		resourceTree:    resourceTree,
		deviceNames:     deviceNames,
		roomAssignments: roomAssignments,
	}
	return metadata, nil
}

func fetchResourceTree(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getResourcesResponse, err := bridgeClient.GetResources()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge resources on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getResourcesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge resources from %q (status: %s)", bridgeClient.Url().Redacted(), getResourcesResponse.HTTPResponse.Status)
	}
	tree := make(map[string]string)
	responseData := getResourcesResponse.JSON200.Data
	if responseData != nil {
		for _, resource := range *responseData {
			resourceId := *resource.Id
			resourceOwnerId := ""
			resourceOwner := resource.Owner
			if resourceOwner != nil {
				resourceOwnerId = *resourceOwner.Rid
				tree[resourceId] = resourceOwnerId
			}
		}
	}
	return tree, nil
}

func fetchDeviceNames(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getDevicesResponse, err := bridgeClient.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge devices on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getDevicesResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge devices from %q (status: %s)", bridgeClient.Url().Redacted(), getDevicesResponse.HTTPResponse.Status)
	}
	names := make(map[string]string)
	responseData := getDevicesResponse.JSON200.Data
	if responseData != nil {
		for _, device := range *responseData {
			deviceId := *device.Id
			deviceName := *device.Metadata.Name
			names[deviceId] = deviceName
		}
	}
	return names, nil
}

func fetchRoomAssignments(bridgeClient hue.BridgeClient) (map[string]string, error) {
	getRoomsResponse, err := bridgeClient.GetRooms()
	if err != nil {
		return nil, fmt.Errorf("failed to access bridge rooms on %q (cause: %w)", bridgeClient.Url().Redacted(), err)
	}
	if getRoomsResponse.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bridge rooms from %q (status: %s)", bridgeClient.Url().Redacted(), getRoomsResponse.HTTPResponse.Status)
	}
	assignments := make(map[string]string)
	responseData := getRoomsResponse.JSON200.Data
	if responseData != nil {
		for _, roomGet := range *responseData {
			roomName := *roomGet.Metadata.Name
			for _, children := range *roomGet.Children {
				childId := *children.Rid
				assignments[childId] = roomName
			}
		}
	}
	return assignments, nil
}

func (metadata *BridgeMetadata) ResolveResourceRoom(resourceId string, resourceName string) string {
	roomName := metadata.roomAssignments[resourceName]
	if roomName == "" {
		resourceOwnerId := resourceId
		for {
			roomName = metadata.roomAssignments[resourceOwnerId]
			if roomName != "" {
				break
			}
			resourceOwnerId = metadata.resourceTree[resourceOwnerId]
			if resourceOwnerId == "" {
				break
			}
		}
	}
	if roomName == "" {
		roomName = "<unassigned>"
	}
	return roomName
}

func (metadata *BridgeMetadata) ResolveDeviceName(resourceId string) string {
	deviceName := ""
	resourceOwnerId := resourceId
	for {
		deviceName = metadata.deviceNames[resourceOwnerId]
		if deviceName != "" {
			break
		}
		resourceOwnerId = metadata.resourceTree[resourceOwnerId]
		if resourceOwnerId == "" {
			break
		}
	}
	if deviceName == "" {
		deviceName = "<undefined>"
	}
	return deviceName
}
