package pdparser

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/zhiduoke/gapi/metadata"
	annotation "github.com/zhiduoke/gapi/proto"
)

type pdServiceOption struct {
	server         string
	defaultHandler string
	defaultTimeout int32
	pathPrefix     string
}

type pdService struct {
	name     string
	fullname string
	opt      pdServiceOption
	methods  []*pdMethod
}

type pdMethodOption struct {
	method  string
	path    string
	use     []string
	timeout int32
	handler string
}

type pdMethod struct {
	name string
	opt  pdMethodOption
	in   *metadata.Message
	out  *metadata.Message
}

type Parser struct {
	ns           []string
	nsstr        string
	msgs         map[string]*metadata.Message
	isEntry      map[string]bool
	enums        map[string]bool
	services     []*pdService
	extraHandler func(msg *metadata.Message, md *descriptor.DescriptorProto)
}

func (p *Parser) enter(ns string) {
	p.ns = append(p.ns, ns)
	p.nsstr = "." + strings.Join(p.ns, ".")
}

func (p *Parser) leave() {
	p.ns = p.ns[:len(p.ns)-1]
	p.nsstr = "." + strings.Join(p.ns, ".")
}

func (p *Parser) AddFile(file *descriptor.FileDescriptorProto) error {
	p.enter(file.GetPackage())
	defer p.leave()
	for _, enumType := range file.EnumType {
		err := p.parseEnum(enumType)
		if err != nil {
			return err
		}
	}
	for _, m := range file.MessageType {
		err := p.parseMessage(m)
		if err != nil {
			return err
		}
	}
	for _, srv := range file.Service {
		service, err := p.parseService(srv)
		if err != nil {
			return err
		}
		p.services = append(p.services, service)
	}
	return nil
}

func (p *Parser) parseService(sd *descriptor.ServiceDescriptorProto) (*pdService, error) {
	svc := &pdService{
		name:     sd.GetName(),
		fullname: strings.TrimLeft(p.nsstr, ".") + "." + sd.GetName(),
	}
	if sd.Options != nil {
		opts, err := proto.GetExtensions(sd.Options, []*proto.ExtensionDesc{
			annotation.E_Server,
			annotation.E_DefaultHandler,
			annotation.E_DefaultTimeout,
			annotation.E_PathPrefix,
		})
		if err != nil {
			return nil, err
		}
		svc.opt = pdServiceOption{
			server:         getString(opts[0], svc.fullname),
			defaultHandler: getString(opts[1], ""),
			defaultTimeout: getInt32(opts[2], 0),
			pathPrefix:     getString(opts[3], ""),
		}
	}
	for _, md := range sd.Method {
		method, err := p.parseMethod(md)
		if err == proto.ErrMissingExtension {
			continue
		}
		if err != nil {
			return nil, err
		}
		svc.methods = append(svc.methods, method)
	}
	return svc, nil
}

func (p *Parser) parseMethod(md *descriptor.MethodDescriptorProto) (*pdMethod, error) {
	method := &pdMethod{
		name: md.GetName(),
	}
	if md.Options == nil {
		return nil, proto.ErrMissingExtension
	}
	httpOpt, err := proto.GetExtension(md.Options, annotation.E_Http)
	if err != nil {
		return nil, err
	}
	opt, ok := httpOpt.(*annotation.Http)
	if !ok {
		return nil, fmt.Errorf("unrecognized http option: %T", httpOpt)
	}
	if opt.Pattern == nil {
		return nil, fmt.Errorf("pattern is not defined")
	}

	switch t := opt.Pattern.(type) {
	case *annotation.Http_Post:
		method.opt.method = "POST"
		method.opt.path = t.Post
	case *annotation.Http_Get:
		method.opt.method = "GET"
		method.opt.path = t.Get
	case *annotation.Http_Put:
		method.opt.method = "PUT"
		method.opt.path = t.Put
	case *annotation.Http_Delete:
		method.opt.method = "DELETE"
		method.opt.path = t.Delete
	case *annotation.Http_Option:
		method.opt.method = "OPTION"
		method.opt.path = t.Option
	case *annotation.Http_Patch:
		method.opt.method = "PATCH"
		method.opt.path = t.Patch
	default:
		return nil, fmt.Errorf("unkonwn pattern %T", opt.Pattern)
	}
	method.opt.timeout = opt.Timeout
	method.opt.handler = opt.Handler
	method.opt.use = opt.Use
	method.in = p.msgs[md.GetInputType()]
	method.out = p.msgs[md.GetOutputType()]
	return method, nil
}

func (p *Parser) parseEnum(ed *descriptor.EnumDescriptorProto) error {
	p.enums[p.nsstr+"."+ed.GetName()] = true
	return nil
}

func (p *Parser) getMessage(name string) *metadata.Message {
	msg := p.msgs[name]
	if msg == nil {
		msg = &metadata.Message{
			Name: name,
		}
		p.msgs[name] = msg
	}
	return msg
}

func (p *Parser) parseMessage(md *descriptor.DescriptorProto) error {
	p.enter(md.GetName())
	defer p.leave()
	fullName := p.nsstr
	msg := p.getMessage(fullName)
	for _, enumType := range md.EnumType {
		err := p.parseEnum(enumType)
		if err != nil {
			return err
		}
	}

	for _, nested := range md.NestedType {
		err := p.parseMessage(nested)
		if err != nil {
			return err
		}
	}

	if md.Options != nil {
		if md.Options.GetMapEntry() {
			p.isEntry[fullName] = true
		}
		opt, err := proto.GetExtension(md.Options, annotation.E_Flat)
		if err != nil && err != proto.ErrMissingExtension {
			return err
		}
		if err == nil {
			msg.Options.Flat = getBool(opt, false)
		}
	}

	var fields []*metadata.Field
	for _, fd := range md.Field {
		kind := mapTypeToKind(fd.GetType())
		if kind == metadata.InvalidType {
			continue
		}
		field := &metadata.Field{
			Tag:      int(fd.GetNumber()),
			Name:     fd.GetName(),
			Kind:     kind,
			Repeated: fd.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED,
		}

		if fd.Options != nil {
			opts, err := proto.GetExtensions(fd.Options, []*proto.ExtensionDesc{
				annotation.E_Alias,
				annotation.E_OmitEmpty,
				annotation.E_RawData,
				annotation.E_FromContext,
				annotation.E_Validate,
				annotation.E_Bind,
			})
			if err != nil && err != proto.ErrMissingExtension {
				return err
			}
			if err == nil {
				field.Name = getString(opts[0], field.Name)
				bind := metadata.FromDefault
				// compatibility
				if getBool(opts[3], false) {
					bind = metadata.FromContext
				}
				if v, ok := opts[5].(*annotation.FIELD_BIND); ok && v != nil {
					switch *v {
					case annotation.FIELD_BIND_FROM_DEFAULT:
						bind = metadata.FromDefault
					case annotation.FIELD_BIND_FROM_CONTEXT:
						bind = metadata.FromContext
					case annotation.FIELD_BIND_FROM_QUERY:
						bind = metadata.FromQuery
					case annotation.FIELD_BIND_FROM_HEADER:
						bind = metadata.FromHeader
					case annotation.FIELD_BIND_FROM_PARAMS:
						bind = metadata.FromParams
					}
				}
				field.Options = metadata.FieldOptions{
					OmitEmpty: getBool(opts[1], false),
					RawData:   getBool(opts[2], false),
					Validate:  getBool(opts[4], false),
					Bind:      bind,
				}
			}
		}
		if kind == metadata.MessageKind {
			msgName := fd.GetTypeName()
			if !strings.HasPrefix(msgName, ".") {
				msgName = fullName + "." + msgName
			}
			field.Message = p.getMessage(msgName)
		}
		fields = append(fields, field)
	}

	// TODO map entry field order

	msg.Fields = fields
	msg.BakeTagIndex()

	if fn := p.extraHandler; fn != nil {
		fn(msg, md)
	}

	return nil
}

func (p *Parser) SetExtraHandler(fn func(msg *metadata.Message, md *descriptor.DescriptorProto)) {
	p.extraHandler = fn
}

func (p *Parser) Resolve() {
	// resolve map kind
	for _, msg := range p.msgs {
		for _, f := range msg.Fields {
			if !f.Repeated || f.Kind != metadata.MessageKind {
				continue
			}
			if p.isEntry[f.Message.Name] {
				f.Kind = metadata.MapKind
			}
		}
	}
}

func (p *Parser) CollectRoutes() ([]*metadata.Route, error) {
	var routes []*metadata.Route
	for _, svc := range p.services {
		prefix := svc.opt.pathPrefix
		if prefix != "" {
			if prefix[0] != '/' {
				return nil, fmt.Errorf("prefix %s must start with '/'", prefix)
			}
			if prefix[len(prefix)-1] == '/' {
				prefix = prefix[:len(prefix)-1]
			}
		}
		for _, method := range svc.methods {
			handler := svc.opt.defaultHandler
			if method.opt.handler != "" {
				handler = method.opt.handler
			}
			timeout := svc.opt.defaultTimeout
			if method.opt.timeout != 0 {
				timeout = method.opt.timeout
			}
			path := method.opt.path
			if path == "" {
				return nil, fmt.Errorf("missing route path of method %s", method.name)
			}
			if path[0] != '/' {
				return nil, fmt.Errorf("path %s must start with '/'", path)
			}
			if prefix != "" {
				path = prefix + path
			}
			routes = append(routes, &metadata.Route{
				Method: method.opt.method,
				Path:   path,
				Options: metadata.RouteOptions{
					Middlewares: method.opt.use,
				},
				Call: &metadata.Call{
					Server:  svc.opt.server,
					Handler: handler,
					Name:    fmt.Sprintf("/%s/%s", svc.fullname, method.name),
					In:      method.in,
					Out:     method.out,
					Timeout: time.Duration(timeout) * time.Millisecond,
				},
			})
		}
	}
	return routes, nil
}

func NewParser() *Parser {
	return &Parser{
		msgs:    map[string]*metadata.Message{},
		isEntry: map[string]bool{},
		enums:   map[string]bool{},
	}
}

func ParseSet(data []byte) (*metadata.Metadata, error) {
	var pd descriptor.FileDescriptorSet
	err := proto.Unmarshal(data, &pd)
	if err != nil {
		return nil, err
	}
	p := NewParser()
	for _, file := range pd.File {
		err := p.AddFile(file)
		if err != nil {
			return nil, err
		}
	}
	p.Resolve()
	routes, err := p.CollectRoutes()
	if err != nil {
		return nil, err
	}
	return &metadata.Metadata{Routes: routes}, nil
}
