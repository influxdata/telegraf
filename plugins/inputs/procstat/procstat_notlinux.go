//+build !linux

package procstat

func getOSSpecificMetrics() map[string]interface{} {
	return nil
}
