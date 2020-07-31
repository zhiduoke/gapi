package gapi

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/julienschmidt/httprouter"
	"github.com/zhiduoke/gapi/metadata"
	"google.golang.org/grpc"
)

type CallHandler interface {
	HandleRequest(call *metadata.Call, ctx *Context) ([]byte, error)
	WriteResponse(call *metadata.Call, ctx *Context, data []byte) error
}

type HandleFunc func(ctx *Context) error

type Server struct {
	router      atomic.Value
	ctxpool     sync.Pool
	routeLock   sync.Mutex
	clients     map[string]*grpc.ClientConn
	middlewares struct {
		sync.RWMutex
		inner map[string]HandleFunc
	}
	callHandlers struct {
		sync.RWMutex
		inner map[string]CallHandler
	}

	globalUses []HandleFunc
	Dial       func(string) (*grpc.ClientConn, error)
	NotFound   http.Handler
}

func (s *Server) getCallHandler(name string) CallHandler {
	s.callHandlers.RLock()
	h := s.callHandlers.inner[name]
	s.callHandlers.RUnlock()
	return h
}

func (s *Server) RegisterHandler(name string, h CallHandler) {
	s.callHandlers.Lock()
	s.callHandlers.inner[name] = h
	s.callHandlers.Unlock()
}

func (s *Server) Use(handle HandleFunc) {
	s.globalUses = append(s.globalUses, handle)
}

func (s *Server) generateMiddlewareChain(mws []string, exec HandleFunc) ([]HandleFunc, error) {
	n := len(mws) + 1
	hs := make([]HandleFunc, 0, n+len(s.globalUses))
	if len(s.globalUses) > 0 {
		hs = append(hs, s.globalUses...)
	}
	s.middlewares.RLock()
	for _, name := range mws {
		mw := s.middlewares.inner[name]
		if mw == nil {
			s.middlewares.RUnlock()
			return nil, fmt.Errorf("no such middleware: %s", name)
		}
		hs = append(hs, mw)
	}
	s.middlewares.RUnlock()
	hs = append(hs, exec)
	return hs, nil
}

func (s *Server) RegisterMiddleware(name string, h HandleFunc) {
	s.middlewares.Lock()
	s.middlewares.inner[name] = h
	s.middlewares.Unlock()
}

func (s *Server) UpdateRoute(md *metadata.Metadata) error {
	s.routeLock.Lock()
	defer s.routeLock.Unlock()

	old := s.clients
	clients := map[string]*grpc.ClientConn{}
	defer func() {
		// close new connections when error occurred
		for server, cc := range clients {
			if old[server] == nil {
				cc.Close()
			}
		}
	}()

	dial := s.Dial
	if dial == nil {
		dial = defaultDial
	}
	// register routes from metadata
	router := httprouter.New()
	for _, route := range md.Routes {
		ch := s.getCallHandler(route.Call.Handler)
		if ch == nil {
			return fmt.Errorf("no such handler: %s", route.Call.Handler)
		}
		// reuse existed connection
		client := old[route.Call.Server]
		if client == nil {
			var err error
			client, err = dial(route.Call.Server)
			if err != nil {
				return err
			}
		}
		clients[route.Call.Server] = client
		rh := &routeHandler{
			s:      s,
			call:   route.Call,
			ch:     ch,
			client: client,
		}
		chain, err := s.generateMiddlewareChain(route.Options.Middlewares, rh.invoke)
		if err != nil {
			return err
		}
		rh.chain = chain
		router.Handle(route.Method, route.Path, rh.handle)
	}
	router.NotFound = s.NotFound
	s.router.Store(router)
	s.clients = clients

	for server, cc := range old {
		if clients[server] == nil {
			cc.Close()
		}
	}
	// don't clean conns
	clients = nil

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.Load().(*httprouter.Router).ServeHTTP(w, req)
}

func NewServer() *Server {
	s := &Server{
		ctxpool: sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
		NotFound: http.NotFoundHandler(),
	}
	s.router.Store(httprouter.New())
	s.middlewares.inner = map[string]HandleFunc{}
	s.callHandlers.inner = map[string]CallHandler{}
	return s
}
