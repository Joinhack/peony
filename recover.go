package peony

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

//if app panic, the filter will handle the error.
func RecoverFilter(c *Controller, filter []Filter) {
	defer func() {
		if err := recover(); err != nil {
			stack := debug.Stack()
			e := NewErrorFromPanic(c.app, string(stack))
			if e != nil {
				description := "unknown error"
				if err != nil {
					description = fmt.Sprint(err)
				}
				e.Description = description
				NewErrorRender(e).Apply(c)
			} else {
				c.resp.Write(stack)
			}
		}
	}()
	filter[0](c, filter[1:])
}

var PANICRegexp = regexp.MustCompile(`^([^:#]+):(\d+)(:\d+)? (.*)$`)

func NewErrorFromPanic(app *App, stack string) *Error {
	idx := strings.Index(stack, app.BasePath)
	if idx == -1 {
		return nil
	}
	end := strings.Index(stack[idx:], "\n")
	if end == -1 {
		return nil
	}
	errorLine := stack[idx : idx+end]
	matched := PANICRegexp.FindStringSubmatch(errorLine)
	if matched == nil {
		return nil
	}
	filepath := matched[1]
	lineNO := matched[2]
	line, _ := strconv.Atoi(lineNO)
	e := &Error{
		Title:       "Panic",
		Path:        filepath[len(app.BasePath):],
		FileName:    filepath[len(app.BasePath):],
		Line:        line,
		SourceLines: MustReadLines(filepath),
	}
	return e
}
