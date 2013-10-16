package peony

import (
	"runtime/debug"
)

//if app panic, the filter will handle the error.
func RecoverFilter(c *Controller, filter []Filter) {
	defer func() {
		if err := recover(); err != nil {
			c.resp.WriteHeader(500, "text/plain")
			stack := debug.Stack()
			ERROR.Println(string(stack))
			c.resp.Write(stack)
		}
	}()
	filter[0](c, filter[1:])
}
