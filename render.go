package peony

import (
	"net/http"
)

type Render interface {
	Apply(req *Request, resp *Response)
}

type TextRender struct {
	Render
	ContentType string
	Text        string
}

func (r *TextRender) Apply(req *Request, resp *Response) {
	if r.ContentType == "" {
		r.ContentType = "text/pain"
	}
	resp.WriteHeader(http.StatusOK, r.ContentType)
	resp.Output.Write([]byte(r.Text))
}
