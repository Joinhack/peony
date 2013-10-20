package peony

import (
	"encoding/json"
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

type TemplateRender struct {
	Render
	RenderParam  interface{}
	TemplateName string
}

type ErrorRender struct {
	Render
	Error error
}

func (r *ErrorRender) Apply(c *Controller) {
	resp := c.resp
	req := c.req
	resp.WriteHeader(http.StatusInternalServerError, "text/"+req.Accept)
	resp.Write([]byte(r.Error.Error()))
}

func (j *JsonRender) Apply(c *Controller) {
	resp := c.resp
	rs, err := json.Marshal(j.Json)
	if err != nil {
		(&ErrorRender{Error: err}).Apply(c)
		return
	}
	resp.WriteHeader(http.StatusOK, "application/json")
	resp.Write(rs)
}

func (r *TextRender) Apply(c *Controller) {
	resp := c.resp
	contentType := r.ContentType
	if contentType == "" {
		contentType = "text/pain"
	}
	resp.WriteHeader(http.StatusOK, r.ContentType)
	resp.Write([]byte(r.Text))
}

func NewJsonRender(json interface{}) Render {
	return &JsonRender{Json: json}
}

func NewTemplateRender(param interface{}, names ...string) Render {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	return &TemplateRender{RenderParam: param, TemplateName: name}
}

func (t *TemplateRender) Apply(c *Controller) {
	resp := c.resp
	templateLoader := c.templateLoader
	resp.WriteHeader(http.StatusOK, "text/html")
	tmplName := t.TemplateName
	//if user choose a template, use the choosed, esle use the default rule for find tempate
	if tmplName == "" {
		tmplName = c.action
	}
	template, err := templateLoader.Lookup(tmplName)
	if err != nil {
		//TODO parse error
		resp.Write([]byte(err.Error()))
		return
	}
	err = template.Execute(resp, t.RenderParam)
	if err != nil {
		//TODO parse error
		resp.Write([]byte(err.Error()))
	}
}
