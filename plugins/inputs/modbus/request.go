//go:build !openbsd

package modbus

import "sort"

type request struct {
	address uint16
	length  uint16
	fields  []field
}

func newRequestsFromFields(fields []field, maxBatchSize uint16) []request {
	if len(fields) == 0 {
		return nil
	}

	// Sort the fields by address (ascending) and length
	sort.Slice(fields, func(i, j int) bool {
		addrI := fields[i].address
		addrJ := fields[j].address
		return addrI < addrJ || (addrI == addrJ && fields[i].length > fields[j].length)
	})

	// Construct the consecutive register chunks for the addresses and construct Modbus requests.
	// For field addresses like [1, 2, 3, 5, 6, 10, 11, 12, 14] we should construct the following
	// requests (1, 3) , (5, 2) , (10, 3), (14 , 1). Furthermore, we should respect field boundaries
	// and the given maximum chunk sizes.
	var requests []request

	current := request{
		address: fields[0].address,
		length:  fields[0].length,
		fields:  []field{fields[0]},
	}

	for _, f := range fields[1:] {
		// Check if we need to interrupt the current chunk and require a new one
		needInterrupt := f.address != current.address+current.length            // not consecutive
		needInterrupt = needInterrupt || f.length+current.length > maxBatchSize // too large

		if !needInterrupt {
			// Still save to add the field to the current request
			current.length += f.length
			current.fields = append(current.fields, f) // TODO: omit the field with a future flag
			continue
		}

		// Finish the current request, add it to the list and construct a new one
		requests = append(requests, current)
		current = request{
			address: f.address,
			length:  f.length,
			fields:  []field{f},
		}
	}
	requests = append(requests, current)

	return requests
}
