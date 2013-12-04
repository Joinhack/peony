package peony

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	//"path/filepath"
	"reflect"
	"strconv"
	"time"
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
	Status int
	Error  error
}

type BinaryRender struct {
	Render
	Reader             io.Reader
	Name               string
	ContentDisposition string
	Len                int64
	ModTime            time.Time
}

func NewErrorRender(err error) *ErrorRender {
	return &ErrorRender{Error: err}
}

func (b *BinaryRender) Apply(c *Controller) {
	resp := c.Resp
	contentDisposition := fmt.Sprintf("%s; filename=%s", b.ContentDisposition, b.Name)
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

func NewFileRender(path string) Render {
	var err error
	var finfo os.FileInfo
	var file *os.File
	if finfo, err = os.Stat(path); err != nil {
		notFound := &Error{Title: "Not Found", Description: err.Error()}
		render := NewErrorRender(notFound)
		render.Status = http.StatusNotFound
		return render
	}
	if finfo.IsDir() {
		render := NewErrorRender(&Error{Title: "Forbidden", Description: "Directory listing not allowed"})
		render.Status = http.StatusForbidden
		return render
	}
	if file, err = os.Open(path); err != nil {
		return NewErrorRender(err)
	}
	return &BinaryRender{ModTime: finfo.ModTime(), Reader: file}
}

func (r *ErrorRender) Apply(c *Controller) {
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
