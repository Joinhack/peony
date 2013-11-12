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

func (r *XmlRender) Apply(c *Controller) {
	resp := c.resp
	bs, err := xml.Marshal(r.Xml)
	if err != nil {
		(&ErrorRender{Error: err}).Apply(c)
		return
	}
	resp.WriteHeader(http.StatusOK, "application/xml")
	resp.Write(bs)
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

func (t *TemplateRender) Apply(c *Controller) {
	resp := c.resp
	templateLoader := c.templateLoader
	resp.WriteHeader(http.StatusOK, "text/html")
	tmplName := t.TemplateName
	//if user choose a template, use the choosed, esle use the default rule for find tempate
	if tmplName == "" {
		tmplName = c.actionName
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

func NewJsonRender(json interface{}) Render {
	return &JsonRender{Json: json}
}

func NewXmlRender(xml interface{}) Render {
	return &XmlRender{Xml: xml}
}

func NewTextRender(s string) Render {
	return &TextRender{Text: s}
}

//renderParam for is the parameter for template execute. templateName is for point the template.
func NewTemplateRender(renderParam interface{}, templateName ...string) Render {
	name := ""
	if len(templateName) > 0 {
		name = templateName[0]
	}
	return &TemplateRender{RenderParam: renderParam, TemplateName: name}
}
