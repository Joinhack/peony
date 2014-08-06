package peony

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	ruleRE            *regexp.Regexp
	argRE             *regexp.Regexp
	HttpMethods       = []string{"GET", "POST", "PUT", "DELETE"}
	ExtendHttpMethods = append(HttpMethods, "WS")
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

	argRE = regexp.MustCompile(strings.Join([]string{
		`((?P<name>\w+)\s*=\s*)?`,
		`(?P<value>`,
		`true|false|`, //bool
		`\d+.\d+|`,    //float
		`\d+.|`,       //float
		`\d+|`,        //int
		`\w+|`,
		`(?P<stringval>\"[^\"]*?")`,
		`)\s*,?`,
	}, ""))
}

type trace struct {
	isRegex  bool
	variable string
}

type Rule struct {
	HttpMethods []string //e.g. GET, POST
	Action      string   //e.g. Controller.Call
	Path        string   //e.g. /app /app/<int:name>
	traces      []*trace
	numRegTrace int
	args        map[string]Parser
	PathRegex   *regexp.Regexp
}

type Parser func(string) (error, string)

func arg2map(arg string) map[string]string {
	m := map[string]string{}
	matched := argRE.FindAllStringSubmatch(arg, -1)
	for _, match := range matched {
		if match[4] != "" {
			m[match[2]] = match[4]
		} else {
			m[match[2]] = match[3]
		}
	}
	return m
}

func IntParser(arg string) (error, string) {
	return nil, "\\d+"
}

func FloatParser(arg string) (error, string) {
	return nil, `\d+.\d+`
}

//e.g. string(10) string(len=10) string(maxlen=10) string(minlen=10)
//string(minlen=10, maxlen=20)
func StringParser(arg string) (err error, expr string) {
	if strings.Index(arg, "=") != -1 {
		var maxlen, minlen string
		var ok bool
		argmap := arg2map(arg)
		if maxlen, ok = argmap["maxlen"]; !ok {
			maxlen = ""
		}
		if minlen, ok = argmap["minlen"]; !ok {
			minlen = ""
		}
		if l, ok := argmap["len"]; ok {
			expr = fmt.Sprintf("[^/]{%s}", l)
			return
		}
		expr = fmt.Sprintf("[^/]{%s, %s}", minlen, maxlen)
		return
	} else if len(arg) > 0 {
		var l int
		if l, err = strconv.Atoi(arg); err != nil {
			return
		}
		expr = fmt.Sprintf("[^/]{%d}", l)
		return
	}
	expr = "[^/]+"
	return
}

func REParser(arg string) (error, string) {
	return nil, arg
}

type Router struct {
	rules         []*Rule
	rulesByAction map[string][]*Rule
	parsers       map[string]Parser
}

func (r *Rule) appendTrace(isRegex bool, variable string) {
	r.traces = append(r.traces, &trace{isRegex, variable})
	if isRegex {
		r.numRegTrace++
	}
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
				var parsedExpr string
				var err error
				if err, parsedExpr = parse(args); err != nil {
					panic(fmt.Sprintf("please check expr: %s, detail error: %s", r.Path, err))
				}
				reg = append(reg, fmt.Sprintf("(?P<%s>%s)", variable, parsedExpr))
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

var (
	NoSuchAction          = errors.New("no such action.")
	CannotBuildWithParams = errors.New("can't build url by these params")
)

func (r *Rule) canbeBuild(params map[string]string) bool {
	if r.PathRegex == nil && (params == nil || len(params) == 0) {
		return true
	}
	for k, _ := range r.args {
		if _, ok := params[k]; !ok {
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
			delete(params, trace.variable)
		} else {
			rs = append(rs, trace.variable)
		}
	}
	return strings.Join(rs, "")
}

func (r *Rule) Match(httpmethod string, path string) (string, map[string]string) {
	if !StringSliceContain(r.HttpMethods, httpmethod) {
		return "", nil
	}
	if r.numRegTrace == 0 {
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
	parsers["float"] = FloatParser
	parsers["re"] = REParser
}

func NewRouter() *Router {
	router := &Router{}
	router.parsers = make(map[string]Parser, 8)
	router.rulesByAction = make(map[string][]*Rule, 0)
	bindDefaultConverter(router.parsers)
	return router
}

func (r *Router) Match(httpmethod string, path string) (string, map[string]string) {
	var action string
	var match map[string]string
	for _, rule := range r.rules {
		if action, match = rule.Match(httpmethod, path); action != "" {
			return action, match
		}
	}
	return "", nil
}

func (r *Router) Refresh() {
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
		return NoSuchAction, ""
	}
	for _, rule := range rules {
		if rule.canbeBuild(params) {
			return nil, rule.Build(params)
		}
	}
	WARN.Println(CannotBuildWithParams, params)
	return CannotBuildWithParams, ""
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

func GetStaticFilter(svr *Server) Filter {
	return func(controller *Controller, filter []Filter) {
		urlpath := controller.Req.URL.Path
		staticInfo := svr.App.StaticInfo
		if staticInfo != nil && strings.HasPrefix(urlpath, staticInfo.UriPrefix) {
			controller.render = RenderFile(path.Join(staticInfo.Path, urlpath[len(staticInfo.UriPrefix):]))
			return
		}
		filter[0](controller, filter[1:])
	}
}

func GetRouterFilter(svr *Server) Filter {
	return func(controller *Controller, filter []Filter) {
		router := svr.Router
		urlpath := controller.Req.URL.Path
		actionName, routerParams := router.Match(controller.Req.Method, urlpath)
		if actionName == "" {
			controller.NotFound("No matched rule found")
			return
		}
		controller.actionName = actionName
		controller.Params = &Params{}
		controller.Params.Router = url.Values{}
		for k, v := range routerParams {
			controller.Params.Router[k] = append(controller.Params.Router[k], []string{v}...)
		}

		// bind actionMethod to controller
		action := svr.FindAction(controller.actionName)
		if action == nil {
			controller.NotFound("Intenal error.")
			ERROR.Println("can't find action method by name:", controller.actionName)
			return
		}
		controller.action = action.Dup()
		filter[0](controller, filter[1:])
	}
}
