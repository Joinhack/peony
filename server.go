package peony

import (
	"net/http"
)

type Filter func(*Controller, []Filter)

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
	output      http.ResponseWriter
}

func NewRequest(r *http.Request) *Request {
	return &Request{Request: r}
}

func NewResponse(r http.ResponseWriter) *Response {
	return &Response{output: r}
}

func (server *Server) handler(w http.ResponseWriter, r *http.Request) {
	server.handlerInner(w, r)
}

func (server *Server) handlerInner(w http.ResponseWriter, r *http.Request) {
	c := NewController(NewResponse(w), NewRequest(r))
	server.filters[0](c, server.filters[1:])
}

func (server *Server) Run() error {
	server.httpServer = &http.Server{Addr: server.Addr, Handler: http.HandlerFunc(server.handler)}
	return server.httpServer.ListenAndServe()
}
