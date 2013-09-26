package peony

import ()

type Controller struct {
	resp *Response
	req  *Request
}

func NewController(w *Response, r *Request) *Controller {
	c := &Controller{resp: w, req: r}

	return c
}
