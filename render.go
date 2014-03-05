package peony

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	RendererType reflect.Type = reflect.TypeOf((*Renderer)(nil)).Elem()
	Attachment   string       = "attachment"
	Inline       string       = "inline"
)

type Renderer interface {
	Apply(c *Controller)
}

type TextRenderer struct {
	Renderer
	ContentType string
	Text        string
}

type JsonRenderer struct {
	Renderer
	Json interface{}
}

type RedirectRenderer struct {
	Renderer
	Url    string
	action interface{} //e.g. Controller.Method (*Controller).Method function
	param  interface{}
}

type XmlRenderer struct {
	Renderer
	Xml interface{}
}

type autoRenderer struct {
	Renderer
	Params interface{}
}

type TemplateRenderer struct {
	Renderer
	RenderParam  interface{}
	TemplateName string
}

type ErrorRenderer struct {
	Renderer
	Status int
	Error  error
}

type BinaryRenderer struct {
	Renderer
	Reader             io.Reader
	Name               string
	ContentDisposition string
	Len                int64
	ModTime            time.Time
}

func RenderError(err error) *ErrorRenderer {
	return &ErrorRenderer{Error: err}
}

func NotFound(msg string, args ...interface{}) *ErrorRenderer {
	text := msg
	if len(args) > 0 {
		text = fmt.Sprintf(msg, args)
	}
	render := RenderError(&Error{Title: "Not Found", Description: text})
	render.Status = 404
	return render
}

func (a *autoRenderer) Apply(c *Controller) {
	switch c.Req.Accept {
	case "json":
		RenderJson(a.Params).Apply(c)
	case "xml":
		RenderXml(a.Params).Apply(c)
	default:
		RenderTemplate(a.Params).Apply(c)
	}
}

func (r *RedirectRenderer) getRedirctUrl(svr *Server) (string, error) {
	if r.Url != "" {
		return r.Url, nil
	}
	locType := reflect.TypeOf(r.action)
	actionName := ""
	if locType.NumIn() > 0 {
		recvType := locType.In(0)
		//support Controller.Method as redirect argument.
		if recvType.Kind() != reflect.Ptr {
			meth := FindMethod(recvType, reflect.ValueOf(r.action))
			if meth != nil {
				actionName = recvType.Name() + "." + meth.Name
			}
		}
	}
	var action *Action
	if actionName != "" {
		action = svr.FindAction(actionName)
	} else {
		action = svr.FindActionByType(locType)
	}
	if action == nil {
		return "", NoSuchAction
	}

	var err error
	var rsurl string
	var buildParams map[string]string
	if params, ok := r.param.(map[string]interface{}); ok {
		buildParams = make(map[string]string, len(params))
		for k, v := range params {
			svr.convertors.ReverseConvert(buildParams, k, v)
		}
	}
	if params, ok := r.param.(map[string]string); ok {
		buildParams = params
	}
	err, rsurl = svr.Router.Build(action.Name, buildParams)
	if err != nil {
		return "", err
	}
	queryValues := make(url.Values)
	for k, v := range buildParams {
		queryValues.Set(k, v)
	}
	if len(queryValues) > 0 {
		rsurl += "?" + queryValues.Encode()
	}
	return rsurl, nil
}

func (r *RedirectRenderer) Apply(c *Controller) {
	url, err := r.getRedirctUrl(c.Server)
	if err != nil {
		RenderError(err).Apply(c)
		return
	}
	c.Resp.Header().Set("Location", url)
	c.Resp.WriteContentTypeCode(http.StatusFound, "")
}

func ParseAction(action string) string {
	return strings.Replace(action, ".", "/", 1)
}

func (b *BinaryRenderer) Apply(c *Controller) {
	resp := c.Resp
	prefix := b.ContentDisposition
	if b.ContentDisposition == "" {
		prefix = Inline
	}
	contentDisposition := fmt.Sprintf("%s; filename=%s", prefix, b.Name)
	resp.Header().Set("Content-Disposition", contentDisposition)
	if readSeeker, ok := b.Reader.(io.ReadSeeker); ok {
		http.ServeContent(c.Resp.ResponseWriter, c.Req.Request, b.Name, b.ModTime, readSeeker)
	} else {
		if b.Len >= 0 {
			resp.Header().Set("Content-Length", strconv.FormatInt(b.Len, 10))
		}
		io.Copy(resp, b.Reader)
	}
	if closer, ok := b.Reader.(io.Closer); ok {
		closer.Close()
	}
}

func RenderFile(path string) Renderer {
	var err error
	var finfo os.FileInfo
	var file *os.File
	if finfo, err = os.Stat(path); err != nil {
		notFound := &Error{Title: "Not Found", Description: err.Error()}
		render := RenderError(notFound)
		render.Status = http.StatusNotFound
		return render
	}
	if finfo.IsDir() {
		render := RenderError(&Error{Title: "Forbidden", Description: "Directory listing not allowed"})
		render.Status = http.StatusForbidden
		return render
	}
	if file, err = os.Open(path); err != nil {
		return RenderError(err)
	}
	return &BinaryRenderer{ModTime: finfo.ModTime(), Name: finfo.Name(), Reader: file, Len: finfo.Size()}
}

func (r *ErrorRenderer) Apply(c *Controller) {
	resp := c.Resp
	req := c.Req
	status := r.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	tplName := fmt.Sprintf("errors/%d.%s", status, req.Accept)
	tpl := c.templateLoader.Lookup(tplName)
	if tpl == nil {
		resp.WriteContentTypeCode(status, "text/"+req.Accept)
		resp.Write([]byte(r.Error.Error()))
		WARN.Println("can't find template", tplName)
		return
	}
	resp.WriteContentTypeCode(status, "text/html")
	var err *Error
	switch r.Error.(type) {
	case *Error:
		err = r.Error.(*Error)
	default:
		err = &Error{
			Title:       "error",
			Description: r.Error.Error(),
		}
	}
	if err := tpl.Execute(c.Resp, err); err != nil {
		ERROR.Println(err)
	}
}

func (j *JsonRenderer) Apply(c *Controller) {
	resp := c.Resp
	rs, err := json.Marshal(j.Json)
	if err != nil {
		(&ErrorRenderer{Error: err}).Apply(c)
		return
	}
	resp.WriteContentTypeCode(http.StatusOK, "application/json")
	resp.Write(rs)
}

func (r *XmlRenderer) Apply(c *Controller) {
	resp := c.Resp
	bs, err := xml.Marshal(r.Xml)
	if err != nil {
		(&ErrorRenderer{Error: err}).Apply(c)
		return
	}
	resp.WriteContentTypeCode(http.StatusOK, "application/xml")
	resp.Write(bs)
}

func (r *TextRenderer) Apply(c *Controller) {
	resp := c.Resp
	contentType := r.ContentType
	if contentType == "" {
		contentType = "text/pain"
	}
	resp.WriteContentTypeCode(http.StatusOK, r.ContentType)
	resp.Write([]byte(r.Text))
}

func (t *TemplateRenderer) Apply(c *Controller) {
	resp := c.Resp
	templateLoader := c.templateLoader
	resp.WriteContentTypeCode(http.StatusOK, "text/html")
	tmplName := t.TemplateName
	//if user choose a template, use the choosed, esle use the default rule for find tempate
	if tmplName == "" {
		tmplName = ParseAction(c.actionName) + ".html"
	}
	template := templateLoader.Lookup(tmplName)
	if template == nil {
		ERROR.Println("can't find template", tmplName)
		resp.Write([]byte("can't find template " + tmplName))
		return
	}
	err := template.Execute(resp, t.RenderParam)
	if err != nil {
		//TODO parse error
		resp.Write([]byte(err.Error()))
	}
}

func RenderJson(json interface{}) Renderer {
	return &JsonRenderer{Json: json}
}

func RenderXml(xml interface{}) Renderer {
	return &XmlRenderer{Xml: xml}
}

func RenderText(s string) Renderer {
	return &TextRenderer{Text: s}
}

func Render(param ...interface{}) Renderer {
	var renderParam interface{}
	if len(param) == 1 {
		renderParam = param[0]
	} else {
		renderParam = param
	}
	return &autoRenderer{Params: renderParam}
}

//renderParam for is the parameter for template execute. templateName is for point the template.
func RenderTemplate(param interface{}, templateName ...string) Renderer {
	name := ""
	if len(templateName) > 0 {
		name = templateName[0]
	}
	return &TemplateRenderer{RenderParam: param, TemplateName: name}
}

var descript = `Redirect parameter must be (function, map[string]string or map[string]interface{}) or ("/%s/%i", "index", 1)`

func Redirect(r interface{}, param ...interface{}) Renderer {
	if loc, ok := r.(string); ok {
		var url string
		if len(param) == 0 {
			url = loc
		} else {
			url = fmt.Sprintf(loc, param)
		}
		return &RedirectRenderer{Url: url}
	}
	var p interface{}

	if reflect.TypeOf(r).Kind() != reflect.Func || len(param) > 1 {
		goto ERR
	}

	if len(param) == 1 {
		p = param[0]
		switch p.(type) {
		case map[string]string, map[string]interface{}:
		default:
			goto ERR
		}
	}

	return &RedirectRenderer{action: r, param: param}
ERR:
	_, f, l, _ := runtime.Caller(1)
	lines, _ := ReadLines(f)

	return RenderError(&Error{
		Title:       "Parameter error",
		Description: descript,
		Path:        f,
		Line:        l,
		SourceLines: lines,
	})
}
