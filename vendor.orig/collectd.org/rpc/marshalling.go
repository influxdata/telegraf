package rpc // import "collectd.org/rpc"

import (
	"collectd.org/api"
	pb "collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// MarshalValue converts an api.Value to a pb.Value.
func MarshalValue(v api.Value) (*pb.Value, error) {
	switch v := v.(type) {
	case api.Counter:
		return &pb.Value{
			Value: &pb.Value_Counter{Counter: uint64(v)},
		}, nil
	case api.Derive:
		return &pb.Value{
			Value: &pb.Value_Derive{Derive: int64(v)},
		}, nil
	case api.Gauge:
		return &pb.Value{
			Value: &pb.Value_Gauge{Gauge: float64(v)},
		}, nil
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "%T values are not supported", v)
	}
}

// UnmarshalValue converts a pb.Value to an api.Value.
func UnmarshalValue(in *pb.Value) (api.Value, error) {
	switch v := in.GetValue().(type) {
	case *pb.Value_Counter:
		return api.Counter(v.Counter), nil
	case *pb.Value_Derive:
		return api.Derive(v.Derive), nil
	case *pb.Value_Gauge:
		return api.Gauge(v.Gauge), nil
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "%T values are not supported", v)
	}
}

// MarshalIdentifier converts an api.Identifier to a pb.Identifier.
func MarshalIdentifier(id *api.Identifier) *pb.Identifier {
	return &pb.Identifier{
		Host:           id.Host,
		Plugin:         id.Plugin,
		PluginInstance: id.PluginInstance,
		Type:           id.Type,
		TypeInstance:   id.TypeInstance,
	}
}

// UnmarshalValue converts a pb.Identifier to an api.Identifier.
func UnmarshalIdentifier(in *pb.Identifier) *api.Identifier {
	return &api.Identifier{
		Host:           in.Host,
		Plugin:         in.Plugin,
		PluginInstance: in.PluginInstance,
		Type:           in.Type,
		TypeInstance:   in.TypeInstance,
	}
}

// MarshalValueList converts an api.ValueList to a pb.ValueList.
func MarshalValueList(vl *api.ValueList) (*pb.ValueList, error) {
	t, err := ptypes.TimestampProto(vl.Time)
	if err != nil {
		return nil, err
	}

	var pbValues []*pb.Value
	for _, v := range vl.Values {
		pbValue, err := MarshalValue(v)
		if err != nil {
			return nil, err
		}

		pbValues = append(pbValues, pbValue)
	}

	return &pb.ValueList{
		Values:     pbValues,
		Time:       t,
		Interval:   ptypes.DurationProto(vl.Interval),
		Identifier: MarshalIdentifier(&vl.Identifier),
	}, nil
}

// UnmarshalValue converts a pb.ValueList to an api.ValueList.
func UnmarshalValueList(in *pb.ValueList) (*api.ValueList, error) {
	t, err := ptypes.Timestamp(in.GetTime())
	if err != nil {
		return nil, err
	}

	interval, err := ptypes.Duration(in.GetInterval())
	if err != nil {
		return nil, err
	}

	var values []api.Value
	for _, pbValue := range in.GetValues() {
		v, err := UnmarshalValue(pbValue)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return &api.ValueList{
		Identifier: *UnmarshalIdentifier(in.GetIdentifier()),
		Time:       t,
		Interval:   interval,
		Values:     values,
		DSNames:    in.DsNames,
	}, nil
}
