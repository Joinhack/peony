package peony

import (
	"net/http"
	"net/url"
)

type Filter func(*Action, []Filter)

type Server struct {
	Addr       string
	httpServer *http.Server
	router     *Router
	filters    []Filter
}

type Request struct {
	*http.Request
}

type Response struct {
	ContentType string
	Output      http.ResponseWriter
}

type Params struct {
	url.Values
	Route url.Values
}

func (r *Response) WriteHeader(code int, contentType string) {
	r.Output.WriteHeader(code)
	if contentType == "" {
		contentType = "text/html"
	}
	r.Output.Header().Set("Content-Type", contentType)
}

func (r *Response) SetHeader(key, value string) {
	r.Output.Header().Set(key, value)
}

func NewRequest(r *http.Request) *Request {
	return &Request{Request: r}
}

func NewResponse(r http.ResponseWriter) *Response {
	return &Response{Output: r}
}

func (server *Server) handler(w http.ResponseWriter, r *http.Request) {
	server.handlerInner(w, r)
}

func (server *Server) handlerInner(w http.ResponseWriter, r *http.Request) {
	c := NewAction(NewResponse(w), NewRequest(r))
	server.filters[0](c, server.filters[1:])
}

func (server *Server) Run() error {
	server.httpServer = &http.Server{Addr: server.Addr, Handler: http.HandlerFunc(server.handler)}
	return server.httpServer.ListenAndServe()
}
