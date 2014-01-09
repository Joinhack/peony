package peony

import (
	"testing"
)

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

	r := &Rule{Path: "/path/12<string:p1>-<int:p2>", Action: "action.method1", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path/12<string:p1>-<int:p2>-", Action: "action.method2", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path/13<re(\\d{1,5}):p3>-<int:p2>--", Action: "action.method3", HttpMethods: HttpMethods}
	router.AddRule(r)

	r = &Rule{Path: "/path/15<p3>-<int:p2>--", Action: "action.method4", HttpMethods: HttpMethods}
	router.AddRule(r)
	r = &Rule{Path: "/path", Action: "static", HttpMethods: HttpMethods}
	router.AddRule(r)

	router.Refresh()

	path := "/path/12-9090-123"
	action, params := router.Match("GET", path)
	t.Log(params)
	_, p := router.Build(action, params)
	t.Logf("%q, %t \n", action, p == path)

	path = "/path"
	action, params = router.Match("GET", path)
	t.Log(params)
	_, p = router.Build(action, params)
	t.Logf("%q, %t\n", action, p == path)

	path = "/path/12-9090-123-"
	action, params = router.Match("GET", path)
	t.Log(params)
	_, p = router.Build(action, params)
	t.Logf("%q, %t\n", action, p == path)

	path = "/path/15909-123--"
	action, params = router.Match("GET", path)
	t.Log(params)
	_, p = router.Build(action, params)
	t.Logf("%q, %t, %q\n", action, p == path)
}
