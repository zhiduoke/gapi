package pdparser

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/zhiduoke/gapi/metadata"
)

func getString(v interface{}, or string) string {
	if s, ok := v.(*string); ok && s != nil {
		return *s
	}
	return or
}

func getInt32(v interface{}, or int32) int32 {
	if s, ok := v.(*int32); ok && s != nil {
		return *s
	}
	return or
}

func getBool(v interface{}, or bool) bool {
	if s, ok := v.(*bool); ok && s != nil {
		return *s
	}
	return or
}

func mapTypeToKind(t descriptor.FieldDescriptorProto_Type) metadata.TypeKind {
	switch t {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return metadata.DoubleKind
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return metadata.FloatKind
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return metadata.Int64Kind
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return metadata.Uint64Kind
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return metadata.Int32Kind
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return metadata.Fixed64Kind
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return metadata.Fixed32Kind
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return metadata.BoolKind
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return metadata.StringKind
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return metadata.MessageKind
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return metadata.BytesKind
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return metadata.Uint32Kind
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return metadata.EnumKind
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return metadata.Sfixed32Kind
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return metadata.Sfixed64Kind
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return metadata.Sint32Kind
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return metadata.Sint64Kind
	default:
		return metadata.InvalidType
	}
}
