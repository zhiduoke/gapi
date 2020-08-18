package metadata

import (
	"sort"
	"time"
)

type TypeKind int

const (
	InvalidType TypeKind = iota
	Int32Kind
	Uint32Kind
	Int64Kind
	Uint64Kind
	BoolKind
	FloatKind
	DoubleKind
	Fixed32Kind
	Fixed64Kind
	EnumKind
	Sfixed32Kind
	Sfixed64Kind
	Sint32Kind
	Sint64Kind
	StringKind
	BytesKind
	MessageKind
	MapKind
	MaxTypeKind
)

const (
	FromDefault = iota
	FromContext
	FromQuery
	FromHeader
	FromParams
)

type FieldOptions struct {
	OmitEmpty bool
	RawData   bool
	Validate  bool
	Bind      int
}

type Field struct {
	Tag      int
	Name     string
	Kind     TypeKind
	Message  *Message
	Repeated bool
	Options  FieldOptions
}

type MessageOptions struct {
	Flat      bool
	ExtraInfo interface{}
}

type Message struct {
	Name      string
	Fields    []*Field
	tagIndex  []int
	nameField []*Field
	Options   MessageOptions
}

func (m *Message) BakeTagIndex() {
	fields := m.Fields
	if len(fields) == 0 {
		m.tagIndex = nil
		return
	}
	maxTag := 0
	for _, f := range fields {
		if f.Tag > maxTag {
			maxTag = f.Tag
		}
	}
	if maxTag-len(m.Fields) < 3 {
		// dense
		tagIndex := make([]int, maxTag+1)
		for i := range tagIndex {
			tagIndex[i] = -1
		}
		for i, f := range fields {
			tagIndex[f.Tag] = i
		}
		m.tagIndex = tagIndex
		return
	}
	// sparse
	tagIndex := make([]int, len(fields))
	for i := range fields {
		tagIndex[i] = i
	}
	sort.Slice(tagIndex, func(i, j int) bool {
		return fields[tagIndex[i]].Tag < fields[tagIndex[j]].Tag
	})
	m.tagIndex = tagIndex
}

func (m *Message) TagIndex(tag int) int {
	if len(m.tagIndex) > len(m.Fields) {
		if tag < len(m.tagIndex) {
			return m.tagIndex[tag]
		} else {
			return -1
		}
	}
	l, r := 0, len(m.tagIndex)-1
	for l <= r {
		mid := (l + r) / 2
		i := m.tagIndex[mid]
		x := m.Fields[i].Tag
		if x == tag {
			return i
		} else if x > tag {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return -1
}

func (m *Message) BakeNameField() {
	if len(m.Fields) == 0 {
		return
	}
	fields := m.Fields
	nameField := make([]*Field, len(fields))
	copy(nameField, fields)
	sort.Slice(nameField, func(i, j int) bool {
		return nameField[i].Name < nameField[j].Name
	})
	m.nameField = nameField
}

func (m *Message) GetField(fieldName string) *Field {
	l, r := 0, len(m.nameField)-1
	for l <= r {
		mid := (l + r) / 2
		field := m.nameField[mid]
		if field.Name == fieldName {
			return field
		} else if field.Name > fieldName {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return nil
}

type Call struct {
	Server  string
	Handler string
	Name    string
	In      *Message
	Out     *Message
	Timeout time.Duration
}

type RouteOptions struct {
	Middlewares []string
}

type Route struct {
	Method  string
	Path    string
	Options RouteOptions
	Call    *Call
}

type Metadata struct {
	Routes []*Route
}
