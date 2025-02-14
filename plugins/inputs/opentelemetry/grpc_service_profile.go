package opentelemetry

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	service "go.opentelemetry.io/proto/otlp/collector/profiles/v1experimental"
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

func (s *profileService) Export(_ context.Context, req *service.ExportProfilesServiceRequest) (*service.ExportProfilesServiceResponse, error) {
	// Output the received message for debugging
	buf, err := protojson.Marshal(req)
	if err != nil {
		s.logger.Errorf("marshalling received profile failed: %v", err)
	} else {
		s.logger.Debugf("received profile: %s", string(buf))
	}

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
				for i, sample := range p.Profile.Sample {
					for j := sample.LocationsStartIndex; j < sample.LocationsStartIndex+sample.LocationsLength; j++ {
						for validx, value := range sample.Value {
							loc := p.Profile.Location[j]
							locations := make([]string, 0, len(loc.Line))
							for _, line := range loc.Line {
								f := p.Profile.Function[line.FunctionIndex]
								fileloc := p.Profile.StringTable[f.Filename]
								if f.StartLine > 0 {
									if fileloc != "" {
										fileloc += " "
									}
									fileloc += "line " + strconv.FormatInt(f.StartLine, 10)
								}
								l := p.Profile.StringTable[f.Name]
								if fileloc != "" {
									l += "(" + fileloc + ")"
								}
								locations = append(locations, l)
							}
							mapping := p.Profile.Mapping[loc.MappingIndex]
							tags := map[string]string{
								"profile_id":       hex.EncodeToString(p.ProfileId),
								"sample":           strconv.Itoa(i),
								"sample_name":      p.Profile.StringTable[p.Profile.PeriodType.Type],
								"sample_unit":      p.Profile.StringTable[p.Profile.PeriodType.Unit],
								"sample_type":      p.Profile.StringTable[p.Profile.SampleType[validx].Type],
								"sample_type_unit": p.Profile.StringTable[p.Profile.SampleType[validx].Unit],
								"address":          "0x" + strconv.FormatUint(loc.Address, 16),
							}
							for k, v := range attrtags {
								tags[k] = v
							}
							fields := map[string]interface{}{
								"start_time_unix_nano": p.StartTimeUnixNano,
								"end_time_unix_nano":   p.EndTimeUnixNano,
								"location":             strings.Join(locations, ","),
								"frame_type":           p.Profile.StringTable[loc.TypeIndex],
								"stack_trace_id":       p.Profile.StringTable[sample.StacktraceIdIndex],
								"memory_start":         mapping.MemoryStart,
								"memory_limit":         mapping.MemoryLimit,
								"filename":             p.Profile.StringTable[mapping.Filename],
								"file_offset":          mapping.FileOffset,
								"build_id":             p.Profile.StringTable[mapping.BuildId],
								"build_id_type":        mapping.BuildIdKind.String(),
								"value":                value,
							}
							for _, idx := range sample.Attributes {
								attr := p.Profile.AttributeTable[idx]
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
