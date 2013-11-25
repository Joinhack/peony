package peony

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	ruleRE *regexp.Regexp
)

func init() {
	ruleRE = regexp.MustCompile(strings.Join([]string{
		`([^<]*)`, //static pattern
		`<(?:`,
		`([a-zA-Z_][a-zA-Z0-9_]*)`, //parser
		`(?:\((.*)\))?\:`,          //parser args
		`)?`,
		"([a-zA-Z_][a-zA-Z0-9_]*)", //paramName
		`>`,
	}, ""))
}

type trace struct {
	isRegex  bool
	variable string
}

type Rule struct {
	Method    []string //e.g. GET, POST
	Action    string   //e.g. Controller.Call
	Path      string   //e.g. /app /app/<int:name>
	traces    []*trace
	args      map[string]Parser
	PathRegex *regexp.Regexp
}

type Parser func(string) string

func IntParser(arg string) string {
	return "\\d+"
}

func StringParser(arg string) string {
	return "[^/]+"
}

func REParser(arg string) string {
	return arg
}

type Router struct {
	rules         []*Rule
	rulesByAction map[string][]*Rule
	parsers       map[string]Parser
}

func (r *Rule) appendTrace(isRegex bool, variable string) {
	r.traces = append(r.traces, &trace{isRegex, variable})
}

func (router *Router) complieRule(r *Rule) error {
	if r.Path == "" {
		return errors.New("can't compile, path is nil")
	}
	matches := ruleRE.FindAllStringSubmatch(r.Path, -1)
	reg := []string{}
	usedNames := NewSet()
	r.args = make(map[string]Parser)
	if matches == nil {
		r.appendTrace(false, r.Path)
		reg = append(reg, r.Path)
	} else {
		var lastMatched string
		for _, match := range matches {
			lastMatched = match[0]
			static := match[1]
			parser := match[2]
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
				if parser == "" {
					parser = "string"
				}
				r.appendTrace(true, variable)
				parse := router.parsers[parser]
				if parse == nil {
					return errors.New("can't compile, unknown parser:" + parser)
				}
				r.args[variable] = parse
				reg = append(reg, fmt.Sprintf("(?P<%s>%s)", variable, parse(args)))
			}
		}
		//last suffix not mutched add to reg slice and collection to trace
		if idx := strings.LastIndex(r.Path, lastMatched); idx+len(lastMatched) < len(r.Path) {
			suffix := r.Path[idx+len(lastMatched):]
			r.appendTrace(false, suffix)
			reg = append(reg, suffix)
		}
	}
	var err error
	r.PathRegex, err = regexp.Compile(fmt.Sprintf("^%s$", strings.Join(reg, "")))
	if err != nil {
		return err
	}
	return nil
}

func (r *Rule) canbeBuild(params map[string]string) bool {
	if r.PathRegex == nil && (params == nil || len(params) == 0) {
		return true
	}
	if len(r.args) != len(params) {
		return false
	}
	for k, _ := range params {
		if r.args[k] == nil {
			return false
		}
	}
	return true
}

func (r *Rule) Build(params map[string]string) string {
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

func (r *Rule) Match(path string) (string, map[string]string) {
	if len(r.args) == 0 {
		if r.Path == path {
			return r.Action, nil
		} else {
			return "", nil
		}
	}
	match := r.PathRegex.FindStringSubmatch(path)
	if match == nil {
		return "", nil
	} else {
		rs := make(map[string]string, len(r.args))
		for idx, name := range r.PathRegex.SubexpNames() {
			if r.args[name] != nil {
				rs[name] = match[idx]
			}
		}
		return r.Action, rs
	}
}

func bindDefaultConverter(parsers map[string]Parser) {
	parsers["string"] = StringParser
	parsers["int"] = IntParser
	parsers["re"] = REParser
}

func NewRouter() *Router {
	router := &Router{}
	router.parsers = make(map[string]Parser, 8)
	router.rulesByAction = make(map[string][]*Rule, 0)
	bindDefaultConverter(router.parsers)
	return router
}

func (r *Router) Match(path string) (string, map[string]string) {
	var action string
	var match map[string]string
	for _, rule := range r.rules {
		if action, match = rule.Match(path); action != "" {
			return action, match
		}
	}
	return "", nil
}

func (r *Router) Update() {
	sort.Sort(r)
}

func (r *Router) Less(i, j int) bool {
	//let static to front
	if len(r.rules[i].args) == 0 && len(r.rules[j].args) > 0 {
		return true
	}
	if len(r.rules[i].args) > 0 && len(r.rules[j].args) == 0 {
		return false
	}
	if len(r.rules[i].args) == 0 && len(r.rules[j].args) == 0 {
		return r.rules[i].Path < r.rules[j].Path
	}
	//complex rule to front
	return len(r.rules[i].args) >= len(r.rules[j].args)
}

func (r *Router) Len() int {
	return len(r.rules)
}

func (r *Router) Swap(i, j int) {
	r.rules[i], r.rules[j] = r.rules[j], r.rules[i]
}

func (r *Router) AddRules(rules []*Rule) error {
	var err error
	for _, rule := range rules {
		if err = r.AddRule(rule); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Build(action string, params map[string]string) (error, string) {
	rules := r.rulesByAction[action]
	if rules == nil {
		return errors.New("no such action:" + action), ""
	}

	for _, rule := range rules {
		if rule.canbeBuild(params) {
			return nil, rule.Build(params)
		}
	}
	return errors.New("can't build by these params"), ""
}

func (r *Router) AddRule(rule *Rule) error {
	if err := r.complieRule(rule); err != nil {
		return err
	}
	r.rules = append(r.rules, rule)
	rules := r.rulesByAction[rule.Action]
	if rules == nil {
		rules = make([]*Rule, 0)
	}
	rules = append(rules, rule)
	r.rulesByAction[rule.Action] = rules
	return nil
}

func GetRouterFilter(svr *Server) Filter {
	return func(controller *Controller, filter []Filter) {
		router := svr.router
		actionName, routerParams := router.Match(controller.Req.URL.Path)
		if actionName == "" {
			controller.NotFound("Not Found")
			return
		}
		controller.actionName = actionName
		controller.params = &Params{}
		controller.params.Router = url.Values{}
		for k, v := range routerParams {
			controller.params.Router[k] = append(controller.params.Router[k], []string{v}...)
		}

		// bind actionMethod to controller
		controller.action = svr.actions.FindAction(controller.actionName)
		if controller.action == nil {
			controller.NotFound("intenal error")
			ERROR.Println("can't find action method by name:", controller.actionName)
			return
		}
		filter[0](controller, filter[1:])
	}
}
