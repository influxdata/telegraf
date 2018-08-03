//+build linux

package procstat

func getOSSpecificMetrics() map[string]interface{} {
	fields := map[string]interface{}{}
	mmaps, err := proc.MemoryMaps()
	if err == nil {
		fields[prefix+"memory_pss"] = mmaps.Pss
	}
}
