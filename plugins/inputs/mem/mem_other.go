//go:build !linux && !windows

package mem

const extendedMemorySupported = false

func getExtendedMemoryFields() (map[string]interface{}, error) {
	return nil, nil
}
