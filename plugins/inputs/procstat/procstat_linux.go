package procstat

// Add Linux-specific metrics into fields map
func (p *Procstat) getOSMetrics(proc Process, prefix string) map[string]interface{} {

	fields := map[string]interface{}{}

	mmaps, err := proc.MemoryMaps(false)
	if err == nil {
		memory_pss := 0
		for _, mmap := range *mmaps {
			// pss is returned in kbytes
			memory_pss += int(mmap.Pss * 1024)
		}
		fields[prefix+"memory_pss"] = memory_pss
	}

	return fields
}
