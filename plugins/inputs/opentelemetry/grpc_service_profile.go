package opentelemetry

import (
	"context"
	"encoding/hex"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	service "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	otlp "go.opentelemetry.io/proto/otlp/profiles/v1development"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

type profileService struct {
	service.UnimplementedProfilesServiceServer

	acc    telegraf.Accumulator
	filter filter.Filter
	logger telegraf.Logger
}

func newProfileService(acc telegraf.Accumulator, logger telegraf.Logger, dimensions []string) (*profileService, error) {
	// Check for duplicate dimensions
	seen := make(map[string]bool, len(dimensions))
	duplicates := make([]string, 0)
	dims := make([]string, 0, len(dimensions))
	for _, d := range dimensions {
		if seen[d] {
			duplicates = append(duplicates, d)
			continue
		}
		dims = append(dims, d)
		seen[d] = true
	}
	if len(duplicates) > 0 {
		return nil, fmt.Errorf("duplicate profile dimension(s) configured: %s", strings.Join(duplicates, ","))
	}
	f, err := filter.Compile(dims)
	if err != nil {
		return nil, fmt.Errorf("compiling dimensions filter failed: %w", err)
	}

	return &profileService{
		acc:    acc,
		filter: f,
		logger: logger,
	}, nil
}

// Export processes and exports the received profile data.
func (s *profileService) Export(_ context.Context, req *service.ExportProfilesServiceRequest) (*service.ExportProfilesServiceResponse, error) {
	// Output the received message for debugging
	buf, err := protojson.Marshal(req)
	if err != nil {
		s.logger.Errorf("marshalling received profile failed: %v", err)
	} else {
		s.logger.Debugf("received profile: %s", string(buf))
	}

	pd := req.GetDictionary()
	logged := make(map[string]bool)

	for _, rp := range req.GetResourceProfiles() {
		if rp == nil {
			continue
		}

		// Extract the requested attributes that should be added as tags
		attrtags := make(map[string]string)
		for _, attr := range rp.GetResource().GetAttributes() {
			if attr != nil && s.filter.Match(attr.Key) {
				attrtags[attr.Key] = attr.GetValue().GetStringValue()
			}
		}

		for _, sp := range rp.GetScopeProfiles() {
			if sp == nil {
				continue
			}
			for _, p := range sp.GetProfiles() {
				if p == nil {
					continue
				}
				for i, sample := range p.GetSamples() {
					if sample == nil {
						continue
					}
					if pd.GetStackTable() == nil {
						if !logged["nil stack table"] {
							s.logger.Errorf("invalid nil stack table: %s", buf)
							logged["nil stack table"] = true
						}
						continue
					}
					if sample.StackIndex < 0 || int(sample.StackIndex) >= len(pd.GetStackTable()) {
						if !logged["stack index"] {
							s.logger.Errorf("invalid stack index %d in request: %s", sample.StackIndex, buf)
							logged["stack index"] = true
						}
						continue
					}
					stack := pd.GetStackTable()[sample.StackIndex]
					if stack == nil {
						continue
					}
					for _, locIdx := range stack.GetLocationIndices() {
						if pd.GetLocationTable() == nil {
							if !logged["nil location table"] {
								s.logger.Errorf("invalid nil location table: %s", buf)
								logged["nil location table"] = true
							}
							continue
						}
						if locIdx < 0 || int(locIdx) >= len(pd.GetLocationTable()) {
							if !logged["location table index"] {
								s.logger.Errorf("invalid location table index %d in request: %s", locIdx, buf)
								logged["location table index"] = true
							}
							continue
						}
						loc := pd.GetLocationTable()[locIdx]
						if loc == nil {
							continue
						}
						for validIdx, value := range sample.GetValues() {
							locations := make([]string, 0, len(loc.GetLines()))
							for _, line := range loc.GetLines() {
								if line == nil {
									continue
								}
								if pd.GetFunctionTable() == nil {
									if !logged["nil function table"] {
										s.logger.Errorf("invalid nil function table: %s", buf)
										logged["nil function table"] = true
									}
									continue
								}
								if line.FunctionIndex < 0 || int(line.FunctionIndex) >= len(pd.GetFunctionTable()) {
									if !logged["function index"] {
										s.logger.Errorf("invalid function index %d in request: %s", line.FunctionIndex, buf)
										logged["function index"] = true
									}
									continue
								}
								f := pd.GetFunctionTable()[line.FunctionIndex]
								if f == nil {
									if !logged["nil function"] {
										s.logger.Errorf("invalid nil function: %s", buf)
										logged["nil function"] = true
									}
									continue
								}
								if pd.GetStringTable() == nil {
									if !logged["nil string table"] {
										s.logger.Errorf("invalid nil string table: %s", buf)
										logged["nil string table"] = true
									}
									continue
								}
								if f.FilenameStrindex < 0 || int(f.FilenameStrindex) >= len(pd.GetStringTable()) {
									if !logged["filename index"] {
										s.logger.Errorf("invalid filename index %d in request: %s", f.FilenameStrindex, buf)
										logged["filename index"] = true
									}
									continue
								}
								fileloc := pd.GetStringTable()[f.FilenameStrindex]
								if f.StartLine > 0 {
									if fileloc != "" {
										fileloc += " "
									}
									fileloc += "line " + strconv.FormatInt(f.StartLine, 10)
								}
								if f.NameStrindex < 0 || int(f.NameStrindex) >= len(pd.GetStringTable()) {
									if !logged["function name index"] {
										s.logger.Errorf("invalid function name index %d in request: %s", f.NameStrindex, buf)
										logged["function name index"] = true
									}
									continue
								}
								l := pd.GetStringTable()[f.NameStrindex]
								if fileloc != "" {
									l += "(" + fileloc + ")"
								}
								locations = append(locations, l)
							}
							mapping := &otlp.Mapping{}

							// Check the indices of the following lookups
							if p.PeriodType == nil {
								if !logged["nil period type"] {
									s.logger.Errorf("invalid nil period type in request: %s", buf)
									logged["nil period type"] = true
								}
								continue
							}
							if p.SampleType == nil {
								if !logged["nil sample type"] {
									s.logger.Errorf("invalid nil sample type in request: %s", buf)
									logged["nil sample type"] = true
								}
								continue
							}
							if p.PeriodType.TypeStrindex < 0 || int(p.PeriodType.TypeStrindex) >= len(pd.GetStringTable()) {
								if !logged["period name index"] {
									s.logger.Errorf("invalid mapping period name index %d in request: %s", p.PeriodType.TypeStrindex, buf)
									logged["period name index"] = true
								}
								continue
							}
							if p.PeriodType.UnitStrindex < 0 || int(p.PeriodType.UnitStrindex) >= len(pd.GetStringTable()) {
								if !logged["period unit index"] {
									s.logger.Errorf("invalid mapping period unit index %d in request: %s", p.PeriodType.UnitStrindex, buf)
									logged["period unit index"] = true
								}
								continue
							}
							if p.SampleType.TypeStrindex < 0 || int(p.SampleType.TypeStrindex) >= len(pd.GetStringTable()) {
								if !logged["sample name index"] {
									s.logger.Errorf("invalid mapping sample name index %d in request: %s", p.SampleType.TypeStrindex, buf)
									logged["sample name index"] = true
								}
								continue
							}
							if p.SampleType.UnitStrindex < 0 || int(p.SampleType.UnitStrindex) >= len(pd.GetStringTable()) {
								if !logged["sample unit index"] {
									s.logger.Errorf("invalid mapping sample unit index %d in request: %s", p.SampleType.UnitStrindex, buf)
									logged["sample unit index"] = true
								}
								continue
							}
							if mapping.FilenameStrindex < 0 || int(mapping.FilenameStrindex) >= len(pd.GetStringTable()) {
								if !logged["mapping filename index"] {
									s.logger.Errorf("invalid mapping filename index %d in request: %s", mapping.FilenameStrindex, buf)
									logged["mapping filename index"] = true
								}
								continue
							}

							// MappingIndex of 0 means unknown or unapplicable mapping, as the
							// first entry in the mapping table is always a null mapping.
							if loc.MappingIndex != 0 {
								if pd.GetMappingTable() == nil {
									if !logged["nil mapping table"] {
										s.logger.Errorf("invalid nil mapping table: %s", buf)
										logged["nil mapping table"] = true
									}
									continue
								}
								if loc.MappingIndex < 0 || int(loc.MappingIndex) >= len(pd.GetMappingTable()) {
									if !logged["mapping index"] {
										s.logger.Errorf("invalid mapping index %d in request: %s", loc.MappingIndex, buf)
										logged["mapping index"] = true
									}
									continue
								}
								mapping = pd.GetMappingTable()[loc.MappingIndex]
							}

							filename := "unknown"
							if mapping != nil &&
								pd.GetStringTable() != nil &&
								mapping.FilenameStrindex >= 0 &&
								int(mapping.FilenameStrindex) < len(pd.GetStringTable()) {
								filename = pd.GetStringTable()[mapping.FilenameStrindex]
							}

							tags := map[string]string{
								"profile_id":       hex.EncodeToString(p.ProfileId),
								"sample":           strconv.Itoa(i),
								"sample_name":      pd.StringTable[p.PeriodType.TypeStrindex],
								"sample_unit":      pd.StringTable[p.PeriodType.UnitStrindex],
								"sample_type":      pd.StringTable[p.SampleType.TypeStrindex],
								"sample_type_unit": pd.StringTable[p.SampleType.UnitStrindex],
								"address":          "0x" + strconv.FormatUint(loc.Address, 16),
							}
							maps.Copy(tags, attrtags)
							fields := map[string]any{
								"start_time_unix_nano": int64(p.TimeUnixNano),
								"end_time_unix_nano":   int64(p.TimeUnixNano + p.DurationNano),
								"location":             strings.Join(locations, ","),
								"memory_start":         mapping.MemoryStart,
								"memory_limit":         mapping.MemoryLimit,
								"filename":             filename,
								"file_offset":          mapping.FileOffset,
								"value":                value,
							}
							for _, idx := range sample.GetAttributeIndices() {
								if idx < 0 || int(idx) >= len(pd.GetAttributeTable()) {
									if !logged["attribute table index"] {
										s.logger.Errorf("invalid attribute table index %d in request: %s", idx, buf)
										logged["attribute table index"] = true
									}
									continue
								}
								attr := pd.GetAttributeTable()[idx]
								if attr == nil || attr.Value == nil {
									continue
								}
								if attr.KeyStrindex < 0 || int(attr.KeyStrindex) >= len(pd.GetStringTable()) {
									if !logged["attribute key index"] {
										s.logger.Errorf("invalid attribute key index %d in request: %s", attr.KeyStrindex, buf)
										logged["attribute key index"] = true
									}
									continue
								}
								key := pd.GetStringTable()[attr.KeyStrindex]
								fields[key] = attr.GetValue().Value
							}
							if validIdx >= len(sample.GetTimestampsUnixNano()) {
								if !logged["timestamp index"] {
									s.logger.Errorf("invalid timestamp index %d in request: %s", validIdx, buf)
									logged["timestamp index"] = true
								}
								continue
							}
							ts := sample.GetTimestampsUnixNano()[validIdx]
							s.acc.AddFields("profiles", fields, tags, time.Unix(0, int64(ts)))
						}
					}
				}
			}
		}
	}
	return &service.ExportProfilesServiceResponse{}, nil
}
