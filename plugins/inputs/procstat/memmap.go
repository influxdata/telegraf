//go:build linux

package procstat

func collectMemmap(proc Process, prefix string, fields map[string]any) {
	memMapStats, err := proc.MemoryMaps(true)
	if err == nil && len(*memMapStats) == 1 {
		memMap := (*memMapStats)[0]
		fields[prefix+"memory_size"] = memMap.Size
		fields[prefix+"memory_pss"] = memMap.Pss
		fields[prefix+"memory_shared_clean"] = memMap.SharedClean
		fields[prefix+"memory_shared_dirty"] = memMap.SharedDirty
		fields[prefix+"memory_private_clean"] = memMap.PrivateClean
		fields[prefix+"memory_private_dirty"] = memMap.PrivateDirty
		fields[prefix+"memory_referenced"] = memMap.Referenced
		fields[prefix+"memory_anonymous"] = memMap.Anonymous
		fields[prefix+"memory_swap"] = memMap.Swap
	}
}
