package peony

import (
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
)

var (
	ActionExist      = errors.New("Action already exist")
	NoSuchMethod     = errors.New("No Such method")
	ShouldTypeAction = errors.New("Action should be TypeAction")
)

type Filter func(*Controller, []Filter)

type Server struct {
	Addr           string
	httpServer     *http.Server
	router         *Router
	filters        []Filter
	converter      *Converter
	actions        *ActionContainer
	notifier       *Notifier
	Interceptors   Interceptors
	templateLoader *TemplateLoader
	App            *App
	SessionManager SessionManager
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
	ParseParems(c.params, c.Req)
	filter[0](c, filter[1:])
}

func (r *Response) WriteContentTypeCode(code int, contentType string) {
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

func NewController(w http.ResponseWriter, r *http.Request, tl *TemplateLoader) *Controller {
	return &Controller{Resp: NewResponse(w), Req: NewRequest(r), templateLoader: tl}
}

func (server *Server) handlerInner(w http.ResponseWriter, r *http.Request) {
	c := NewController(w, r, server.templateLoader)
	c.Server = server
	server.filters[0](c, server.filters[1:])
	if c.render != nil {
		c.render.Apply(c)
	}
}

func (s *Server) BindDefaultFilters() {
	s.filters = []Filter{
		RecoverFilter,
		GetNotifyFilter(s),
		GetRouterFilter(s),
		GetSessionFilter(s),
		ParamsFilter,
		GetInterceptorFilter(s),
		GetActionInvokeFilter(s),
	}
}

//mapper the func, e.g. func Index() ...
func (s *Server) FuncMapper(expr string, function interface{}, action *Action) error {
	if s.actions.FindAction(action.Name) != nil {
		return ActionExist
	}
	if err := s.actions.RegisterFuncAction(function, action); err != nil {
		return err
	}
	return s.router.AddRule(&Rule{Path: expr, Action: action.Name})
}

//mapper the func with recv, e.g. func (c *C) Index() ...
func (s *Server) MethodMapper(expr string, method interface{}, action *Action) error {
	if s.actions.FindAction(action.Name) != nil {
		return ActionExist
	}
	if err := s.actions.RegisterMethodAction(method, action); err != nil {
		return err
	}
	return s.router.AddRule(&Rule{Path: expr, Action: action.Name})
}

func NewServer(app *App) *Server {
	s := &Server{Addr: app.BindAddr}
	s.App = app
	return s
}

var serverInitHooks = []func(*Server){}

func OnServerInit(f func(*Server)) {
	serverInitHooks = append(serverInitHooks, f)
}

func (s *Server) Init() {
	s.router = NewRouter()
	s.actions = NewActionContainer()
	s.converter = NewConverter()
	s.Interceptors = NewInterceptors()
	//the default views is priority, used for render error, follower template loader error.
	s.templateLoader = NewTemplateLoader([]string{path.Join(PEONYPATH, "views"), s.App.ViewPath})
	s.notifier = NewNotifier()
	s.notifier.Watch(s.templateLoader)

	s.BindDefaultFilters()
	for _, f := range serverInitHooks {
		f(s)
	}
}

func (server *Server) Run() error {
	server.httpServer = &http.Server{Addr: server.Addr, Handler: http.HandlerFunc(server.handler)}
	return server.httpServer.ListenAndServe()
}
