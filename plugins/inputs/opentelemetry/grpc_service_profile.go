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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if pd == nil {
		return nil, status.Error(codes.InvalidArgument, "profile dictionary is required")
	}

	for _, rp := range req.ResourceProfiles {
		// Extract the requested attributes that should be added as tags
		attrtags := make(map[string]string)
		if rp.Resource != nil {
			for _, attr := range rp.Resource.Attributes {
				if s.filter.Match(attr.Key) {
					attrtags[attr.Key] = attr.GetValue().GetStringValue()
				}
			}
		}

		for _, sp := range rp.ScopeProfiles {
			for _, p := range sp.Profiles {
				// These four are constant for the profile, so resolve them once
				// and drop the whole profile if any of them cannot be resolved.
				sampleName, nameOK := tableEntry(pd.StringTable, p.GetPeriodType().GetTypeStrindex())
				sampleUnit, unitOK := tableEntry(pd.StringTable, p.GetPeriodType().GetUnitStrindex())
				sampleType, typeOK := tableEntry(pd.StringTable, p.GetSampleType().GetTypeStrindex())
				sampleTypeUnit, typeUnitOK := tableEntry(pd.StringTable, p.GetSampleType().GetUnitStrindex())
				if !nameOK || !unitOK || !typeOK || !typeUnitOK {
					s.logger.Errorf("dropping profile %s: period or sample type names out of range", hex.EncodeToString(p.ProfileId))
					continue
				}

				for i, sample := range p.Samples {
					stack, ok := tableEntry(pd.StackTable, sample.StackIndex)
					if !ok {
						s.logger.Errorf("dropping sample %d: stack index %d out of range", i, sample.StackIndex)
						continue
					}
					for _, locIdx := range stack.LocationIndices {
						loc, ok := tableEntry(pd.LocationTable, locIdx)
						if !ok {
							s.logger.Errorf("dropping sample %d: location index %d out of range", i, locIdx)
							continue
						}
						locations, ok := s.resolveLocation(pd, loc)
						if !ok {
							s.logger.Errorf("dropping sample %d: location %d references an out-of-range function or string", i, locIdx)
							continue
						}
						mapping := &otlp.Mapping{}
						// MappingIndex of 0 means unknown or unapplicable mapping, as the
						// first entry in the  mapping table is always a null mapping.
						if loc.MappingIndex != 0 {
							mapping, ok = tableEntry(pd.MappingTable, loc.MappingIndex)
							if !ok {
								s.logger.Errorf("dropping sample %d: mapping index %d out of range", i, loc.MappingIndex)
								continue
							}
						}
						mappingFilename, ok := tableEntry(pd.StringTable, mapping.FilenameStrindex)
						if !ok {
							s.logger.Errorf("dropping sample %d: mapping filename index %d out of range", i, mapping.FilenameStrindex)
							continue
						}
						sampleAttributes, ok := s.resolveAttributes(pd, sample.AttributeIndices)
						if !ok {
							s.logger.Errorf("dropping sample %d: attribute index out of range", i)
							continue
						}
						for validx, value := range sample.Values {
							ts, ok := tableEntry(sample.TimestampsUnixNano, int32(validx))
							if !ok {
								s.logger.Errorf("dropping sample %d: no timestamp for value %d", i, validx)
								continue
							}
							tags := map[string]string{
								"profile_id":       hex.EncodeToString(p.ProfileId),
								"sample":           strconv.Itoa(i),
								"sample_name":      sampleName,
								"sample_unit":      sampleUnit,
								"sample_type":      sampleType,
								"sample_type_unit": sampleTypeUnit,
								"address":          "0x" + strconv.FormatUint(loc.Address, 16),
							}
							for k, v := range attrtags {
								tags[k] = v
							}
							fields := map[string]interface{}{
								"start_time_unix_nano": int64(p.TimeUnixNano),
								"end_time_unix_nano":   int64(p.TimeUnixNano + p.DurationNano),
								"location":             strings.Join(locations, ","),
								"memory_start":         mapping.MemoryStart,
								"memory_limit":         mapping.MemoryLimit,
								"filename":             mappingFilename,
								"file_offset":          mapping.FileOffset,
								"value":                value,
							}
							for k, v := range sampleAttributes {
								fields[k] = v
							}
							s.acc.AddFields("profiles", fields, tags, time.Unix(0, int64(ts)))
						}
					}
				}
			}
		}
	}
	return &service.ExportProfilesServiceResponse{}, nil
}

// tableEntry reads an entry from one of the cross-reference tables carried in a
// profile request. Every index in the payload is chosen by the sender, so an
// out-of-range one is a malformed message rather than a bug in the sender, and
// it must not reach a slice index expression: the plugin's gRPC server installs
// no recovery interceptor, so the resulting panic ends the Telegraf process.
func tableEntry[T any](table []T, index int32) (T, bool) {
	if index < 0 || int(index) >= len(table) {
		var zero T
		return zero, false
	}
	return table[index], true
}

// resolveLocation renders the function names and file positions of a location,
// reporting false if any index in it is out of range.
func (s *profileService) resolveLocation(pd *otlp.ProfilesDictionary, loc *otlp.Location) ([]string, bool) {
	locations := make([]string, 0, len(loc.Lines))
	for _, line := range loc.Lines {
		f, ok := tableEntry(pd.FunctionTable, line.FunctionIndex)
		if !ok {
			return nil, false
		}
		fileloc, ok := tableEntry(pd.StringTable, f.FilenameStrindex)
		if !ok {
			return nil, false
		}
		if f.StartLine > 0 {
			if fileloc != "" {
				fileloc += " "
			}
			fileloc += "line " + strconv.FormatInt(f.StartLine, 10)
		}
		l, ok := tableEntry(pd.StringTable, f.NameStrindex)
		if !ok {
			return nil, false
		}
		if fileloc != "" {
			l += "(" + fileloc + ")"
		}
		locations = append(locations, l)
	}
	return locations, true
}

// resolveAttributes reads the sample attributes, reporting false if any index is
// out of range.
func (s *profileService) resolveAttributes(pd *otlp.ProfilesDictionary, indices []int32) (map[string]interface{}, bool) {
	attributes := make(map[string]interface{}, len(indices))
	for _, idx := range indices {
		attr, ok := tableEntry(pd.AttributeTable, idx)
		if !ok {
			return nil, false
		}
		key, ok := tableEntry(pd.StringTable, attr.KeyStrindex)
		if !ok {
			return nil, false
		}
		attributes[key] = attr.GetValue().Value
	}
	return attributes, true
}
