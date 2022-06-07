package modbus

import "sort"

type request struct {
	address uint16
	length  uint16
	fields  []field
	tags    map[string]string
}

func newRequest(f field, tags map[string]string) request {
	r := request{
		address: f.address,
		length:  f.length,
		fields:  []field{f},
		tags:    map[string]string{},
	}

	// Copy the tags
	for k, v := range tags {
		r.tags[k] = v
	}
	return r
}

func groupFieldsToRequests(fields []field, tags map[string]string, maxBatchSize uint16) []request {
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

	current := newRequest(fields[0], tags)
	for _, f := range fields[1:] {
		// Check if we need to interrupt the current chunk and require a new one
		needInterrupt := f.address != current.address+current.length            // not consecutive
		needInterrupt = needInterrupt || f.length+current.length > maxBatchSize // too large

		if !needInterrupt {
			// Still safe to add the field to the current request
			current.length += f.length
			if !f.omit {
				// Omit adding the field but use it for constructing the request.
				current.fields = append(current.fields, f)
			}
			continue
		}

		// Finish the current request, add it to the list and construct a new one
		requests = append(requests, current)
		current = newRequest(f, tags)
	}
	requests = append(requests, current)

	return requests
}
