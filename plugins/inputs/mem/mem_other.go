//go:build !linux && !windows

package mem

const extendedMemorySupported = false

func getExtendedMemoryFields() (map[string]any, error) {
	return nil, nil
}
