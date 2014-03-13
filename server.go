package peony

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
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
	Interceptors
	*http.Server
	*Router
	*ActionContainer
	filters        []Filter
	convertors     *Convertors
	notifier       *Notifier
	templateLoader *TemplateLoader
	App            *App
	SessionManager SessionManager
	Addr           string
	listener       net.Listener
}

func GetRandomListenPort() int {
	ipcon, err := net.Listen("tcp", ":0")
	if err != nil {
		ERROR.Fatalln("getListenPort error:", err)
	}
	ipcon.Close()
	return ipcon.Addr().(*net.TCPAddr).Port
}

type Request struct {
	*http.Request
	ContentType string
	Accept      string
	WSConn      *websocket.Conn
}

type Response struct {
	http.ResponseWriter
	ContentType string
}

type Params struct {
	url.Values
	Router   url.Values //e.g. /xx/<int:name>/ param for router.
	Url      url.Values
	Form     url.Values
	Files    map[string][]*multipart.FileHeader
	tmpFiles []*os.File
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
		values = make(url.Values, l)
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
	ParseParems(c.Params, c.Req)
	filter[0](c, filter[1:])
	for _, tmp := range c.Params.tmpFiles {
		if err := os.Remove(tmp.Name()); err != nil {
			WARN.Println("remove temp file error,", err)
		}
	}
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
	upgrade := r.Header.Get("Upgrade")
	if upgrade == "websocket" || upgrade == "Websocket" {
		websocket.Handler(func(wsconn *websocket.Conn) {
			req := NewRequest(r)
			req.Method = "WS"
			req.WSConn = wsconn
			server.handlerInner(NewResponse(w), req)
		}).ServeHTTP(w, r)
	} else {
		server.handlerInner(NewResponse(w), NewRequest(r))
	}
}

func NewController(w http.ResponseWriter, r *http.Request, loader *TemplateLoader) *Controller {
	return &Controller{Resp: NewResponse(w), Req: NewRequest(r), templateLoader: loader}
}

func (server *Server) handlerInner(resp *Response, req *Request) {
	c := &Controller{Resp: resp, Req: req, templateLoader: server.templateLoader}
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
		//static filter must be location after notify filter, the static used the templateloader should inited in notify.
		GetStaticFilter(s),
		GetRouterFilter(s),
		GetSessionFilter(s),
		ParamsFilter,
		GetInterceptorFilter(s),
		GetActionInvokeFilter(s),
	}
}

//mapper the func, e.g. func Index() ...
func (s *Server) FuncMapper(expr string, httpMethods []string, function interface{}, action *Action) error {
	if err := s.RegisterFuncAction(function, action); err != nil {
		return err
	}
	return s.AddRule(&Rule{Path: expr, Action: action.Name, HttpMethods: httpMethods})
}

//mapper the func with recv, e.g. func (c *C) Index() ...
func (s *Server) MethodMapper(expr string, httpMethods []string, method interface{}, action *Action) error {
	if err := s.RegisterMethodAction(method, action); err != nil {
		return err
	}
	return s.AddRule(&Rule{Path: expr, Action: action.Name, HttpMethods: httpMethods})
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
	s.Router = NewRouter()
	s.ActionContainer = NewActionContainer()
	s.convertors = NewConvertors()
	s.Interceptors = NewInterceptors()
	//the peony project's views is priority, used for render error, follower template loader error.
	if s.App.DevMode {
		s.templateLoader = NewTemplateLoader([]string{s.App.ViewPath, path.Join(GetPeonyPath(), "views")})
	}
	s.templateLoader.BindServerTemplateFunc(s)
	s.notifier = NewNotifier()
	s.notifier.Watch(s.templateLoader)

	s.BindDefaultFilters()
	for _, f := range serverInitHooks {
		f(s)
	}
}

func (svr *Server) Listen() error {
	addr := svr.Addr
	if addr == "" {
		addr = ":http"
	}
	l, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}
	svr.listener = l
	return nil
}

func (svr *Server) CloseListener() {
	if svr.listener != nil {
		svr.listener.Close()
		svr.listener = nil
	}
}

func (svr *Server) RegisterSessionManager(sm SessionManager) {
	svr.SessionManager = sm
}

func (svr *Server) Run() <-chan error {
	svr.Server = &http.Server{Addr: svr.Addr, Handler: http.HandlerFunc(svr.handler)}
	if svr.listener == nil {
		err := svr.Listen()
		if err != nil {
			return nil
		}
	}
	waitChan := make(chan error, 1)
	go func() {
		waitChan <- svr.Server.Serve(svr.listener)
	}()
	return waitChan
}
