package jtop

import "github.com/zhiduoke/gapi/metadata"

var numericKinds = [...]bool{
	metadata.DoubleKind:   true,
	metadata.FloatKind:    true,
	metadata.Int32Kind:    true,
	metadata.Int64Kind:    true,
	metadata.Uint32Kind:   true,
	metadata.Uint64Kind:   true,
	metadata.Sint32Kind:   true,
	metadata.Sint64Kind:   true,
	metadata.Fixed32Kind:  true,
	metadata.Fixed64Kind:  true,
	metadata.Sfixed32Kind: true,
	metadata.Sfixed64Kind: true,
	metadata.BoolKind:     true,
}

func isNumeric(kind metadata.TypeKind) bool {
	if kind >= metadata.TypeKind(len(numericKinds)) || kind < 0 {
		return false
	}
	return numericKinds[kind]
}

func fieldNullable(filed *metadata.Field) bool {
	return filed.Repeated ||
		filed.Kind == metadata.BytesKind ||
		filed.Kind == metadata.MapKind ||
		filed.Kind == metadata.MessageKind
}

func isValueToken(token *Token) bool {
	switch token.Kind {
	case Null, True, False, Number, String, ObjectBegin, ArrayBegin:
		return true
	default:
		return false
	}
}
