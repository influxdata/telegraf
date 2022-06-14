//go:build linux && amd64
// +build linux,amd64

package intel_pmu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

// Maximum size of core IDs or socket IDs (8192). Based on maximum value of CPUs that linux kernel supports.
const maxIDsSize = 1 << 13

type entitiesParser interface {
	parseEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) (err error)
}

type configParser struct {
	log telegraf.Logger
	sys sysInfoProvider
}

func (cp *configParser) parseEntities(coreEntities []*CoreEventEntity, uncoreEntities []*UncoreEventEntity) (err error) {
	if len(coreEntities) == 0 && len(uncoreEntities) == 0 {
		return fmt.Errorf("neither core nor uncore entities configured")
	}

	for _, coreEntity := range coreEntities {
		if coreEntity == nil {
			return fmt.Errorf("core entity is nil")
		}
		if coreEntity.Events == nil {
			if cp.log != nil {
				cp.log.Debug("all core events from provided files will be configured")
			}
			coreEntity.allEvents = true
		} else {
			events := cp.parseEvents(coreEntity.Events)
			if events == nil {
				return fmt.Errorf("an empty list of core events was provided")
			}
			coreEntity.parsedEvents = events
		}

		coreEntity.parsedCores, err = cp.parseCores(coreEntity.Cores)
		if err != nil {
			return fmt.Errorf("error during cores parsing: %v", err)
		}
	}

	for _, uncoreEntity := range uncoreEntities {
		if uncoreEntity == nil {
			return fmt.Errorf("uncore entity is nil")
		}
		if uncoreEntity.Events == nil {
			if cp.log != nil {
				cp.log.Debug("all uncore events from provided files will be configured")
			}
			uncoreEntity.allEvents = true
		} else {
			events := cp.parseEvents(uncoreEntity.Events)
			if events == nil {
				return fmt.Errorf("an empty list of uncore events was provided")
			}
			uncoreEntity.parsedEvents = events
		}

		uncoreEntity.parsedSockets, err = cp.parseSockets(uncoreEntity.Sockets)
		if err != nil {
			return fmt.Errorf("error during sockets parsing: %v", err)
		}
	}
	return nil
}

func (cp *configParser) parseEvents(events []string) []*eventWithQuals {
	if len(events) == 0 {
		return nil
	}

	events, duplications := removeDuplicateStrings(events)
	for _, duplication := range duplications {
		if cp.log != nil {
			cp.log.Warnf("duplicated event `%s` will be removed", duplication)
		}
	}
	return parseEventsWithQualifiers(events)
}

func (cp *configParser) parseCores(cores []string) ([]int, error) {
	if cores == nil {
		if cp.log != nil {
			cp.log.Debug("all possible cores will be configured")
		}
		if cp.sys == nil {
			return nil, fmt.Errorf("system info provider is nil")
		}
		cores, err := cp.sys.allCPUs()
		if err != nil {
			return nil, fmt.Errorf("cannot obtain all cpus: %v", err)
		}
		return cores, nil
	}
	if len(cores) == 0 {
		return nil, fmt.Errorf("an empty list of cores was provided")
	}

	result, err := cp.parseIntRanges(cores)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (cp *configParser) parseSockets(sockets []string) ([]int, error) {
	if sockets == nil {
		if cp.log != nil {
			cp.log.Debug("all possible sockets will be configured")
		}
		if cp.sys == nil {
			return nil, fmt.Errorf("system info provider is nil")
		}
		sockets, err := cp.sys.allSockets()
		if err != nil {
			return nil, fmt.Errorf("cannot obtain all sockets: %v", err)
		}
		return sockets, nil
	}
	if len(sockets) == 0 {
		return nil, fmt.Errorf("an empty list of sockets was provided")
	}

	result, err := cp.parseIntRanges(sockets)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (cp *configParser) parseIntRanges(ranges []string) ([]int, error) {
	var ids []int
	var duplicatedIDs []int
	var err error
	ids, err = parseIDs(ranges)
	if err != nil {
		return nil, err
	}
	ids, duplicatedIDs = removeDuplicateValues(ids)
	for _, duplication := range duplicatedIDs {
		if cp.log != nil {
			cp.log.Warnf("duplicated id number `%d` will be removed", duplication)
		}
	}
	return ids, nil
}

func parseEventsWithQualifiers(events []string) []*eventWithQuals {
	var result []*eventWithQuals

	for _, event := range events {
		newEventWithQualifiers := &eventWithQuals{}

		split := strings.Split(event, ":")
		newEventWithQualifiers.name = split[0]

		if len(split) > 1 {
			newEventWithQualifiers.qualifiers = split[1:]
		}
		result = append(result, newEventWithQualifiers)
	}
	return result
}

func parseIDs(allIDsStrings []string) ([]int, error) {
	var result []int
	for _, idsString := range allIDsStrings {
		ids := strings.Split(idsString, ",")

		for _, id := range ids {
			id := strings.TrimSpace(id)
			// a-b support
			var start, end uint
			n, err := fmt.Sscanf(id, "%d-%d", &start, &end)
			if err == nil && n == 2 {
				if start >= end {
					return nil, fmt.Errorf("`%d` is equal or greater than `%d`", start, end)
				}
				for ; start <= end; start++ {
					if len(result)+1 > maxIDsSize {
						return nil, fmt.Errorf("requested number of IDs exceeds max size `%d`", maxIDsSize)
					}
					result = append(result, int(start))
				}
				continue
			}
			// Single value
			num, err := strconv.Atoi(id)
			if err != nil {
				return nil, fmt.Errorf("wrong format for id number `%s`: %v", id, err)
			}
			if len(result)+1 > maxIDsSize {
				return nil, fmt.Errorf("requested number of IDs exceeds max size `%d`", maxIDsSize)
			}
			result = append(result, num)
		}
	}
	return result, nil
}

func removeDuplicateValues(intSlice []int) (result []int, duplicates []int) {
	keys := make(map[int]bool)

	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			result = append(result, entry)
		} else {
			duplicates = append(duplicates, entry)
		}
	}
	return result, duplicates
}

func removeDuplicateStrings(strSlice []string) (result []string, duplicates []string) {
	keys := make(map[string]bool)

	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			result = append(result, entry)
		} else {
			duplicates = append(duplicates, entry)
		}
	}
	return result, duplicates
}
