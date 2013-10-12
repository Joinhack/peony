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
	Apply(req *Request, resp *Response)
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
	RenderParams []interface{}
	TemplateName string
}

type ErrorRender struct {
	Render
	Error error
}

func (r *ErrorRender) Apply(req *Request, resp *Response) {
	resp.WriteHeader(http.StatusInternalServerError, "text/"+req.Accept)
	resp.Write([]byte(r.Error.Error()))
}

func (j *JsonRender) Apply(req *Request, resp *Response) {
	rs, err := json.Marshal(j.Json)
	if err != nil {
		(&ErrorRender{Error: err}).Apply(req, resp)
		return
	}
	resp.WriteHeader(http.StatusOK, "application/json")
	resp.Write(rs)
}

func (r *TextRender) Apply(req *Request, resp *Response) {
	if r.ContentType == "" {
		r.ContentType = "text/pain"
	}
	resp.WriteHeader(http.StatusOK, r.ContentType)
	resp.Write([]byte(r.Text))
}
