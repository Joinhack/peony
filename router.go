package peony

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ruleRE *regexp.Regexp
)

func init() {
	ruleRE = regexp.MustCompile(strings.Join([]string{
		`(?P<static>[^<]*)`, //static pattern
		`<(?:`,
		`(?P<converter>[a-zA-Z_][a-zA-Z0-9_]*)`, //converter
		`(?:\((?P<args>.*)\))?\:`,               //converter args
		`)?`,
		"(?P<variable>[a-zA-Z_][a-zA-Z0-9_]*)", //paramName
		`>`,
	}, ""))
}

type trace struct {
	isRegex  bool
	variable string
}

type Route struct {
	Method    []string //e.g. GET, POST
	Action    string   //e.g. Controller.Call
	Path      string   //e.g. /app /app/<int:name>
	traces    []*trace
	args      map[string]Converter
	PathRegex *regexp.Regexp
}

type Converter func(string) string

func IntConverter(arg string) string {
	return "-?\\d+"
}

func StringConverter(arg string) string {
	return "[^/]+"
}

type Router struct {
	route   []*Route
	convers map[string]Converter
}

func (r *Route) appendTrace(isRegex bool, variable string) {
	r.traces = append(r.traces, &trace{isRegex, variable})
}

func (router *Router) complieRoute(r *Route) error {
	if r.Path == "" {
		return errors.New("can't compile, path is nil")
	}
	matches := ruleRE.FindAllStringSubmatch(r.Path, -1)
	reg := []string{}
	usedNames := NewSet()
	r.args = make(map[string]Converter)
	if matches == nil {
		r.appendTrace(false, r.Path)
		reg = append(reg, r.Path)
	} else {
		for _, match := range matches {
			static := match[1]
			converter := match[2]
			args := match[3]
			variable := match[4]
			if static != "" {
				r.appendTrace(false, static)
				reg = append(reg, static)
			}
			if variable != "" {
				if usedNames.Has(variable) {
					return errors.New("have the same variable:" + variable)
				}
				usedNames.Add(variable)
				if converter == "" {
					converter = "string"
				}
				r.appendTrace(true, variable)
				conver := router.convers[converter]
				if conver == nil {
					return errors.New("can't compile, unknown converter:" + converter)
				}
				r.args[variable] = conver
				reg = append(reg, fmt.Sprintf("(?P<%s>%s)", variable, conver(args)))
			}
		}
	}
	var err error
	r.PathRegex, err = regexp.Compile(strings.Join(reg, ""))
	if err != nil {
		return err
	}
	return nil
}

func (r *Route) Build(params map[string]string) string {
	rs := make([]string, len(r.traces))
	for _, trace := range r.traces {
		if trace.isRegex {
			rs = append(rs, params[trace.variable])
		} else {
			rs = append(rs, trace.variable)
		}
	}
	return strings.Join(rs, "")
}

func (r *Route) Match(path string) map[string]string {
	match := r.PathRegex.FindStringSubmatch(path)
	if match == nil {
		return nil
	} else {
		rs := make(map[string]string, len(r.args))
		for idx, name := range r.PathRegex.SubexpNames() {
			if r.args[name] != nil {
				rs[name] = match[idx]
			}
		}
		return rs
	}
}

func NewRoute(method, path, action string) *Route {
	route := &Route{
		Action: action,
		Path:   path,
	}
	return route
}

func bindDefaultConverter(convers map[string]Converter) {
	convers["string"] = StringConverter
	convers["int"] = IntConverter
}

func NewRouter() *Router {
	router := &Router{}
	router.convers = make(map[string]Converter, 8)
	bindDefaultConverter(router.convers)
	return router
}

//
func (r *Router) Add(route *Route) {
	r.route = append(r.route, route)
}
