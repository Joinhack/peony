package peony

import (
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

var (
	NotAction = errors.New("action should be a func")
)

type Filter func(*Controller, []Filter)

type Server struct {
	Addr           string
	httpServer     *http.Server
	router         *Router
	filters        []Filter
	converter      *Converter
	actions        *ActionMethods
	templateLoader *TemplateLoader
}

type Request struct {
	*http.Request
	ContentType string
	Accept      string
}

type Response struct {
	http.ResponseWriter
	ContentType string
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

func ResolveAccept(req *http.Request) string {
	accept := req.Header.Get("accept")

	switch {
	case accept == "",
		strings.HasPrefix(accept, "*/*"), // */
		strings.Contains(accept, "application/xhtml"),
		strings.Contains(accept, "text/html"):
		return "html"
	case strings.Contains(accept, "application/xml"),
		strings.Contains(accept, "text/xml"):
		return "xml"
	case strings.Contains(accept, "text/plain"):
		return "txt"
	case strings.Contains(accept, "application/json"),
		strings.Contains(accept, "text/javascript"):
		return "json"
	}

	return "html"
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

func ParamsFilter(c *Controller, filter []Filter) {
	ParseParems(c.params, c.req)
	filter[0](c, filter[1:])
}

func (r *Response) WriteHeader(code int, contentType string) {
	if contentType == "" {
		contentType = "text/html"
	}
	r.Header().Set("Content-Type", contentType)
	r.ResponseWriter.WriteHeader(code)
}

func (r *Response) SetHeader(key, value string) {
	r.Header().Set(key, value)
}

func NewRequest(r *http.Request) *Request {
	return &Request{Request: r,
		ContentType: ResolveContentType(r),
		Accept:      ResolveAccept(r),
	}
}

func NewResponse(r http.ResponseWriter) *Response {
	return &Response{ResponseWriter: r}
}

func (server *Server) handler(w http.ResponseWriter, r *http.Request) {
	server.handlerInner(w, r)
}

func (server *Server) handlerInner(w http.ResponseWriter, r *http.Request) {
	c := &Controller{resp: NewResponse(w),
		req:            NewRequest(r),
		templateLoader: server.templateLoader,
	}
	server.filters[0](c, server.filters[1:])
}

func (s *Server) BindDefaultFilters() {
	s.filters = []Filter{
		RecoverFilter,
		GetRouterFilter(s.router),
		ParamsFilter,
		GetActionFilter(s),
	}
}

func (s *Server) Mapper(exp string, action interface{}, actionMethod *ActionMethod) error {
	actionType := reflect.TypeOf(action)
	if actionType.Kind() != reflect.Func {
		return NotAction
	}
	s.router.AddRule(&Rule{Path: exp, Action: actionMethod.Name})
	s.actions.RegisterAction(action, actionMethod)
	return nil
}

func NewServer() *Server {
	s := &Server{Addr: ":8080"}
	s.router = NewRouter()
	s.actions = NewActionMethods()
	s.converter = NewConverter()
	s.BindDefaultFilters()
	return s
}

func (server *Server) Run() error {
	server.httpServer = &http.Server{Addr: server.Addr, Handler: http.HandlerFunc(server.handler)}
	return server.httpServer.ListenAndServe()
}
