package telemetry

import (
	"fmt"
	"reflect"

	"github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_devm"
	"github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_ifm"
	"google.golang.org/protobuf/proto"
)

// ProtoTypes holds a set of reflect.Type
type ProtoTypes struct {
	typeSet []reflect.Type // Array can not repeat
}

// PathKey is the key struct in map
type PathKey struct {
	ProtoPath string
	Version   string
}

// ProtoOrganizeType represents the proto key value
type ProtoOrganizeType int

const (
	// ProtoHuaweiType represents huawei type, put them under one proto, and encaps them into type
	ProtoHuaweiType ProtoOrganizeType = 0
	// ProtoIetfType represents IETF type
	ProtoIetfType ProtoOrganizeType = 1
)

const (
	// DefaultVersion is the default version for proto
	DefaultVersion = "1.0"
)

// GetProtoPaths returns all ProtoPath
func GetProtoPaths() []*PathKey {
	paths := make([]*PathKey, len(pathTypeMap))
	i := 0
	for key := range pathTypeMap {
		path := key
		paths[i] = &path
		i++
	}
	return paths
}

// GetProtoTypeSetByKey returns reflect.Type set pointer by protokey
func GetProtoTypeSetByKey(p *PathKey) *ProtoTypes {
	set := &ProtoTypes{
		typeSet: pathTypeMap[*p],
	}
	if set.typeSet == nil {
		return nil
	}
	return set
}

// GetTypeByProtoPath returns proto.Message with protoPath and version
func GetTypeByProtoPath(protoPath string, version string) (proto.Message, error) {
	if version == "" {
		version = DefaultVersion
	}
	mapping := GetProtoTypeSetByKey(
		&PathKey{
			ProtoPath: protoPath,
			Version:   DefaultVersion})
	if mapping == nil {
		return nil, fmt.Errorf("the proto type is nil, protoPath is %s", protoPath)
	}
	typeInMap := mapping.GetTypesByProtoOrg(ProtoHuaweiType) // using reflect
	elem := typeInMap.Elem()
	reflectType := reflect.New(elem)
	contentType := reflectType.Interface().(proto.Message)
	return contentType, nil
}

// GetTypesByProtoOrg returns proto type by proto
func (p *ProtoTypes) GetTypesByProtoOrg(orgType ProtoOrganizeType) reflect.Type {
	varTypes := p.typeSet
	if varTypes == nil {
		return nil
	}
	if len(varTypes) > int(orgType) {
		return varTypes[orgType]
	}
	return nil
}

// pathTypeMap maps protoPath + version to reflect.Type array
var pathTypeMap = map[PathKey][]reflect.Type{

	{ProtoPath: "huawei_ifm.Ifm", Version: "1.0"}:   {reflect.TypeOf((*huawei_ifm.Ifm)(nil))},
	{ProtoPath: "huawei_devm.Devm", Version: "1.0"}: {reflect.TypeOf((*huawei_devm.Devm)(nil))},
}
