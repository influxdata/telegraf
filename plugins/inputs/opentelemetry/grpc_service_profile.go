package opentelemetry

import (
	"context"
	"encoding/hex"
	"fmt"
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

	pd := req.Dictionary

	for _, rp := range req.ResourceProfiles {
		// Extract the requested attributes that should be added as tags
		attrtags := make(map[string]string)
		for _, attr := range rp.Resource.Attributes {
			if s.filter.Match(attr.Key) {
				attrtags[attr.Key] = attr.GetValue().GetStringValue()
			}
		}

		for _, sp := range rp.ScopeProfiles {
			for _, p := range sp.Profiles {
				for i, sample := range p.Sample {
					for j := sample.LocationsStartIndex; j < sample.LocationsStartIndex+sample.LocationsLength; j++ {
						for validx, value := range sample.Value {
							locIdx := p.LocationIndices[j]
							loc := pd.LocationTable[locIdx]
							locations := make([]string, 0, len(loc.Line))
							for _, line := range loc.Line {
								f := pd.FunctionTable[line.FunctionIndex]
								fileloc := pd.StringTable[f.FilenameStrindex]
								if f.StartLine > 0 {
									if fileloc != "" {
										fileloc += " "
									}
									fileloc += "line " + strconv.FormatInt(f.StartLine, 10)
								}
								l := pd.StringTable[f.NameStrindex]
								if fileloc != "" {
									l += "(" + fileloc + ")"
								}
								locations = append(locations, l)
							}
							mapping := &otlp.Mapping{}
							if loc.MappingIndex != nil {
								mapping = pd.MappingTable[*loc.MappingIndex]
							}
							tags := map[string]string{
								"profile_id":       hex.EncodeToString(p.ProfileId),
								"sample":           strconv.Itoa(i),
								"sample_name":      pd.StringTable[p.PeriodType.TypeStrindex],
								"sample_unit":      pd.StringTable[p.PeriodType.UnitStrindex],
								"sample_type":      pd.StringTable[p.SampleType[validx].TypeStrindex],
								"sample_type_unit": pd.StringTable[p.SampleType[validx].UnitStrindex],
								"address":          "0x" + strconv.FormatUint(loc.Address, 16),
							}
							for k, v := range attrtags {
								tags[k] = v
							}
							fields := map[string]interface{}{
								"start_time_unix_nano": p.TimeNanos,
								"end_time_unix_nano":   p.TimeNanos + p.DurationNanos,
								"location":             strings.Join(locations, ","),
								"memory_start":         mapping.MemoryStart,
								"memory_limit":         mapping.MemoryLimit,
								"filename":             pd.StringTable[mapping.FilenameStrindex],
								"file_offset":          mapping.FileOffset,
								"value":                value,
							}
							for _, idx := range sample.AttributeIndices {
								attr := pd.AttributeTable[idx]
								fields[attr.Key] = attr.GetValue().Value
							}
							ts := sample.TimestampsUnixNano[validx]
							s.acc.AddFields("profiles", fields, tags, time.Unix(0, int64(ts)))
						}
					}
				}
			}
		}
	}
	return &service.ExportProfilesServiceResponse{}, nil
}
