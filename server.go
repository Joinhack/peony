package peony

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Filter func(*Controller, []Filter)

type Server struct {
	Addr       string
	httpServer *http.Server
	router     *Router
	filters    []Filter
}

type Request struct {
	ContentType string
	*http.Request
}

type Response struct {
	ContentType string
	Output      http.ResponseWriter
}

type Params struct {
	url.Values
	Router url.Values //e.g. /xx/<int:name>/ param for router.
	Url    url.Values
	Form   url.Values
	Files  map[string][]*multipart.FileHeader
}

func ResolveContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return "text/html"
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

func ParseParems(params *Params, req *Request) {
	params.Url = req.URL.Query()

	// Parse the body depending on the content type.
	switch req.ContentType {
	case "application/x-www-form-urlencoded":
		// Typical form.
		if err := req.ParseForm(); err != nil {
			ERROR.Println("parse form error, detail:", err)
		} else {
			params.Form = req.Form
		}

	case "multipart/form-data":
		// Multipart form.
		// TODO: Extract the multipart form param so app can set it.
		if err := req.ParseMultipartForm(32 << 20 /* 32 MB */); err != nil {
			ERROR.Println("parse form error, detail:", err)
		} else {
			params.Form = req.MultipartForm.Value
			params.Files = req.MultipartForm.File
		}
	}
	params.mergeValues()
}

func (p *Params) mergeValues() {
	l := len(p.Url) + len(p.Form) + len(p.Router)
	var values url.Values = nil
	switch l {
	case len(p.Url):
		values = p.Url
	case len(p.Router):
		values = p.Router
	case len(p.Form):
		values = p.Form
	}
	if values == nil {
		values := make(url.Values, l)
		for k, v := range p.Url {
			values[k] = append(values[k], v...)
		}
		for k, v := range p.Router {
			values[k] = append(values[k], v...)
		}
		for k, v := range p.Form {
			values[k] = append(values[k], v...)
		}
	}
	p.Values = values
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
	return &Request{Request: r, ContentType: ResolveContentType(r)}
}

func NewResponse(r http.ResponseWriter) *Response {
	return &Response{Output: r}
}

func (server *Server) handler(w http.ResponseWriter, r *http.Request) {
	server.handlerInner(w, r)
}

func (server *Server) handlerInner(w http.ResponseWriter, r *http.Request) {
	c := NewController(NewResponse(w), NewRequest(r))
	server.filters[0](c, server.filters[1:])
}

func (s *Server) BindDefaultFilters() {
	s.filters = []Filter{
		GetRouterFilter(s.router),
	}
}

func NewServer() *Server {
	s := &Server{Addr: ":8080"}
	s.router = NewRouter()
	s.BindDefaultFilters()
	return s
}

func (server *Server) Run() error {
	server.httpServer = &http.Server{Addr: server.Addr, Handler: http.HandlerFunc(server.handler)}
	return server.httpServer.ListenAndServe()
}
