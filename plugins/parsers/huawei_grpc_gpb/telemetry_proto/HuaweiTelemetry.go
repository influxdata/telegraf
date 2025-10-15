package telemetry

import (
	"fmt"
	"reflect"

	"github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_devm"
	"github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_ifm"
	"google.golang.org/protobuf/proto"
)

// struct reflect.Type set
type ProtoTypes struct {
	typeSet []reflect.Type // Array can not repeat
}

// the key struct in map
type PathKey struct {
	ProtoPath string
	Version   string
}

// proto key value
type ProtoOrganizeType int

const (
	// proto type mark represents huawei, put them under one proto, and encaps them into type
	PROTO_HUAWEI_TYPE = 0
	PROTO_IETF_TYPE   = 1
)

const (
	DEFAULT_VERSION = "1.0"
)

// get all ProtoPath
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

// get reflect.Type set pointer by protokey
func GetProtoTypeSetByKey(p *PathKey) *ProtoTypes {
	set := &ProtoTypes{
		typeSet: pathTypeMap[*p],
	}
	if set.typeSet == nil {
		return nil
	}
	return set
}

// get protoPath with protoPath and version
func GetTypeByProtoPath(protoPath string, version string) (proto.Message, error) {
	if version == "" {
		version = DEFAULT_VERSION
	}
	mapping := GetProtoTypeSetByKey(
		&PathKey{
			ProtoPath: protoPath,
			Version:   DEFAULT_VERSION})
	if mapping == nil {
		return nil, fmt.Errorf("the proto type is nil , protoPath is %s", protoPath)
	}
	typeInMap := mapping.GetTypesByProtoOrg(PROTO_HUAWEI_TYPE) // using reflect
	elem := typeInMap.Elem()
	reflectType := reflect.New(elem)
	contentType := reflectType.Interface().(proto.Message)
	return contentType, nil
}

// get proto type by proto
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

// one map key: protoPath + version, value : reflect[]
var pathTypeMap = map[PathKey][]reflect.Type{

	{ProtoPath: "huawei_ifm.Ifm", Version: "1.0"}:   {reflect.TypeOf((*huawei_ifm.Ifm)(nil))},
	{ProtoPath: "huawei_devm.Devm", Version: "1.0"}: {reflect.TypeOf((*huawei_devm.Devm)(nil))},
}
