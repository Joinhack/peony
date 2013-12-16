package peony

import (
	"testing"
)

func TestRE(t *testing.T) {
	re := ruleRE
	arg := re.FindAllStringSubmatch("aaas", -1)
	t.Logf("%q\n", re.SubexpNames())
	t.Logf("%q\n", arg)
	arg = re.FindAllStringSubmatch("asdad<aw(sse):a><w>", -1)
	t.Logf("%q\n", re.SubexpNames())
	t.Logf("%q\n", arg)
}

func TestRule(t *testing.T) {
	router := NewRouter()
	r := &Rule{Path: "/path/12<string:oo>-<int:m>"}
	router.complieRule(r)

	path := "/path/12-9090-123"
	_, params := r.Match("GET", path)
	t.Logf("%t\n", r.Build(params) == path)

	r = &Rule{Path: "/path"}
	err := router.complieRule(r)
	if err != nil {
		panic(err)
	}

	path = "/path"
	_, params = r.Match("GET", path)
	t.Logf("%t\n", r.Build(params) == path)
}

func TestRouter(t *testing.T) {
	router := NewRouter()

	r := &Rule{Path: "/path/12<string:p1>-<int:p2>", Action: "aa.ss", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path/12<string:p1>-<int:p2>-", Action: "e.ssmm", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path/13<re(\\d{1,5}):p3>-<int:p2>--", Action: "e.ssmm", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path", Action: "static", HttpMethods: HttpMethods}
	router.AddRule(r)

	router.Update()

	path := "/path/12-9090-123"
	action, params := router.Match("GET", path)
	_, p := router.Build(action, params)
	t.Logf("%q, %t \n", action, p == path)

	path = "/path"
	action, params = router.Match("GET", path)
	_, p = router.Build(action, params)
	t.Logf("%q, %t\n", action, p == path)

	path = "/path/12-9090-123-"
	action, params = router.Match("GET", path)
	_, p = router.Build(action, params)
	t.Logf("%q, %t\n", action, p == path)

	path = "/path/13909-123--"
	action, params = router.Match("GET", path)
	_, p = router.Build(action, params)
	t.Logf("%q, %t\n", action, p == path)
}
