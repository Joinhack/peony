package peony

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
)

var renderType reflect.Type

func init() {
	renderType = reflect.TypeOf((*Render)(nil)).Elem()
}

type Render interface {
	Apply(c *Controller)
}

type TextRender struct {
	Render
	ContentType string
	Text        string
}

type JsonRender struct {
	Render
	Json interface{}
}

type XmlRender struct {
	Render
	Xml interface{}
}

type TemplateRender struct {
	Render
	RenderParam  interface{}
	TemplateName string
}

type ErrorRender struct {
	Render
	Error error
}

func NewErrorRender(err error) Render {
	return &ErrorRender{Error: err}
}

func (r *ErrorRender) Apply(c *Controller) {
	resp := c.Resp
	req := c.Req
	tplName := "errors/500.html"
	tpl := c.templateLoader.Lookup(tplName)
	if tpl == nil {
		resp.WriteContentTypeCode(http.StatusInternalServerError, "text/"+req.Accept)
		resp.Write([]byte(r.Error.Error()))
		WARN.Println("can't find template", tplName)
		return
	}
	resp.WriteContentTypeCode(http.StatusInternalServerError, "text/html")
	var errorlist ErrorList
	switch r.Error.(type) {
	case ErrorList:
		errorlist = r.Error.(ErrorList)
	case *Error:
		err := r.Error.(*Error)
		errorlist = ErrorList{err}
	default:
		errorlist = ErrorList{&Error{
			Title:       "error",
			Description: r.Error.Error(),
		}}
	}
	tpl.Execute(c.Resp, errorlist)
}

func (j *JsonRender) Apply(c *Controller) {
	resp := c.Resp
	rs, err := json.Marshal(j.Json)
	if err != nil {
		(&ErrorRender{Error: err}).Apply(c)
		return
	}
	resp.WriteContentTypeCode(http.StatusOK, "application/json")
	resp.Write(rs)
}

func (r *XmlRender) Apply(c *Controller) {
	resp := c.Resp
	bs, err := xml.Marshal(r.Xml)
	if err != nil {
		(&ErrorRender{Error: err}).Apply(c)
		return
	}
	resp.WriteContentTypeCode(http.StatusOK, "application/xml")
	resp.Write(bs)
}

func (r *TextRender) Apply(c *Controller) {
	resp := c.Resp
	contentType := r.ContentType
	if contentType == "" {
		contentType = "text/pain"
	}
	resp.WriteContentTypeCode(http.StatusOK, r.ContentType)
	resp.Write([]byte(r.Text))
}

func (t *TemplateRender) Apply(c *Controller) {
	resp := c.Resp
	templateLoader := c.templateLoader
	resp.WriteContentTypeCode(http.StatusOK, "text/html")
	tmplName := t.TemplateName
	//if user choose a template, use the choosed, esle use the default rule for find tempate
	if tmplName == "" {
		tmplName = c.actionName
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

func NewJsonRender(json interface{}) *JsonRender {
	return &JsonRender{Json: json}
}

func NewXmlRender(xml interface{}) *XmlRender {
	return &XmlRender{Xml: xml}
}

func NewTextRender(s string) *TextRender {
	return &TextRender{Text: s}
}

//renderParam for is the parameter for template execute. templateName is for point the template.
func NewTemplateRender(param interface{}, templateName ...string) *TemplateRender {
	name := ""
	if len(templateName) > 0 {
		name = templateName[0]
	}
	return &TemplateRender{RenderParam: param, TemplateName: name}
}
